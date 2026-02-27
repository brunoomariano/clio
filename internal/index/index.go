package index

import (
	"math"
	"strings"
	"sync"
	"time"

	"clio/internal/model"
)

type Index struct {
	mu        sync.RWMutex
	notesByID map[string]*model.Note
	docLen    map[string]int
	avgDocLen float64
	df        map[string]int
	tf        map[string]map[string]int
}

func NewIndex() *Index {
	return &Index{
		notesByID: make(map[string]*model.Note),
		docLen:    make(map[string]int),
		df:        make(map[string]int),
		tf:        make(map[string]map[string]int),
	}
}

func (idx *Index) Reset(notes []*model.Note) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.notesByID = make(map[string]*model.Note, len(notes))
	idx.docLen = make(map[string]int, len(notes))
	idx.df = make(map[string]int)
	idx.tf = make(map[string]map[string]int)
	for _, note := range notes {
		idx.upsertLocked(note)
	}
	idx.recalculateAvgDocLen()
}

func (idx *Index) Upsert(note *model.Note) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.removeLocked(note.ID)
	idx.upsertLocked(note)
	idx.recalculateAvgDocLen()
}

func (idx *Index) Remove(id string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.removeLocked(id)
	idx.recalculateAvgDocLen()
}

func (idx *Index) NotesCount() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.notesByID)
}

func (idx *Index) Get(id string) (*model.Note, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	note, ok := idx.notesByID[id]
	return note, ok
}

func (idx *Index) AllNotes() []*model.Note {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	out := make([]*model.Note, 0, len(idx.notesByID))
	for _, note := range idx.notesByID {
		out = append(out, note)
	}
	return out
}

func (idx *Index) upsertLocked(note *model.Note) {
	stored := cloneNote(note)
	idx.notesByID[note.ID] = stored
	content := idx.noteContent(stored)
	tokens := Tokenize(content)
	idx.docLen[stored.ID] = len(tokens)
	seen := make(map[string]int)
	for _, t := range tokens {
		seen[t]++
	}
	for token, count := range seen {
		if idx.tf[token] == nil {
			idx.tf[token] = make(map[string]int)
		}
		idx.tf[token][stored.ID] = count
		idx.df[token]++
	}
}

func (idx *Index) removeLocked(id string) {
	note, ok := idx.notesByID[id]
	if !ok {
		return
	}
	content := idx.noteContent(note)
	tokens := Tokenize(content)
	seen := make(map[string]struct{})
	for _, t := range tokens {
		seen[t] = struct{}{}
	}
	for token := range seen {
		if docs, ok := idx.tf[token]; ok {
			delete(docs, id)
			if len(docs) == 0 {
				delete(idx.tf, token)
			}
		}
		if count, ok := idx.df[token]; ok {
			if count <= 1 {
				delete(idx.df, token)
			} else {
				idx.df[token] = count - 1
			}
		}
	}
	delete(idx.notesByID, id)
	delete(idx.docLen, id)
}

func (idx *Index) recalculateAvgDocLen() {
	if len(idx.docLen) == 0 {
		idx.avgDocLen = 0
		return
	}
	var total int
	for _, length := range idx.docLen {
		total += length
	}
	idx.avgDocLen = float64(total) / float64(len(idx.docLen))
}

func (idx *Index) noteContent(note *model.Note) string {
	return strings.Join([]string{
		note.Title,
		strings.Join(note.Tags, " "),
		note.Body,
	}, " ")
}

func (idx *Index) bm25Score(token string, docID string, k1, b float64) float64 {
	N := float64(len(idx.notesByID))
	df := float64(idx.df[token])
	if df == 0 {
		return 0
	}
	idf := math.Log(1 + (N-df+0.5)/(df+0.5))
	f := float64(idx.tf[token][docID])
	if f == 0 {
		return 0
	}
	dl := float64(idx.docLen[docID])
	avgdl := idx.avgDocLen
	if avgdl == 0 {
		avgdl = 1
	}
	den := f + k1*(1-b+b*(dl/avgdl))
	return idf * (f * (k1 + 1) / den)
}

type SearchOptions struct {
	Query       string
	MaxResults  int
	BoostTags   []string
	ExcludeTags []string
	BoostWeight float64
	K1          float64
	B           float64
	Regex       bool
	Now         time.Time
}

type SearchResult struct {
	Note  *model.Note
	Score float64
}

func (idx *Index) Search(opts SearchOptions) ([]SearchResult, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	if opts.MaxResults <= 0 {
		opts.MaxResults = 200
	}
	excludeSet := normalizeTagSet(opts.ExcludeTags)
	boostSet := normalizeTagSet(opts.BoostTags)

	results := make([]SearchResult, 0)
	if opts.Regex {
		return idx.searchRegex(opts, excludeSet, boostSet)
	}
	tokens := Tokenize(opts.Query)
	candidates := make(map[string]struct{})
	for _, token := range tokens {
		for docID := range idx.tf[token] {
			candidates[docID] = struct{}{}
		}
	}
	for docID := range candidates {
		note := idx.notesByID[docID]
		if note == nil {
			continue
		}
		if isExpired(note, opts.Now) {
			continue
		}
		if hasExcludedTag(note, excludeSet) {
			continue
		}
		score := 0.0
		for _, token := range tokens {
			score += idx.bm25Score(token, docID, opts.K1, opts.B)
		}
		if score == 0 {
			continue
		}
		boostCount := countBoostTags(note, boostSet)
		score += opts.BoostWeight * float64(boostCount)
		results = append(results, SearchResult{Note: note, Score: score})
	}
	SortResults(results)
	if len(results) > opts.MaxResults {
		return results[:opts.MaxResults], nil
	}
	return results, nil
}

func (idx *Index) searchRegex(opts SearchOptions, excludeSet, boostSet map[string]struct{}) ([]SearchResult, error) {
	re, err := CompileRegex(opts.Query)
	if err != nil {
		return nil, err
	}
	results := make([]SearchResult, 0)
	for _, note := range idx.notesByID {
		if isExpired(note, opts.Now) {
			continue
		}
		if hasExcludedTag(note, excludeSet) {
			continue
		}
		content := idx.noteContent(note)
		if !re.MatchString(content) {
			continue
		}
		score := 1.0
		boostCount := countBoostTags(note, boostSet)
		score += opts.BoostWeight * float64(boostCount)
		results = append(results, SearchResult{Note: note, Score: score})
	}
	SortResults(results)
	if len(results) > opts.MaxResults {
		return results[:opts.MaxResults], nil
	}
	return results, nil
}

func normalizeTagSet(tags []string) map[string]struct{} {
	set := make(map[string]struct{}, len(tags))
	for _, t := range tags {
		t = strings.TrimSpace(strings.ToLower(t))
		if t == "" {
			continue
		}
		set[t] = struct{}{}
	}
	return set
}

func hasExcludedTag(note *model.Note, excludeSet map[string]struct{}) bool {
	for _, tag := range note.Tags {
		if _, ok := excludeSet[strings.ToLower(tag)]; ok {
			return true
		}
	}
	return false
}

func countBoostTags(note *model.Note, boostSet map[string]struct{}) int {
	count := 0
	for _, tag := range note.Tags {
		if _, ok := boostSet[strings.ToLower(tag)]; ok {
			count++
		}
	}
	return count
}

func isExpired(note *model.Note, now time.Time) bool {
	if note.ExpiresAt == nil {
		return false
	}
	return note.ExpiresAt.Before(now)
}

func cloneNote(note *model.Note) *model.Note {
	if note == nil {
		return nil
	}
	cloned := *note
	if note.Tags != nil {
		cloned.Tags = append([]string{}, note.Tags...)
	}
	return &cloned
}
