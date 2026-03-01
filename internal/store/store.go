package store

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"clio/internal/model"
)

var (
	ErrNoteNotFound = errors.New("note not found")
)

const tagIndexDirName = ".clio_tags"

type NoteStore struct {
	dir     string
	sources []Source
}

func NewNoteStore(dir string) *NoteStore {
	return NewNoteStoreWithSources(dir, []Source{
		{
			Path:           dir,
			SuffixPatterns: []string{"*.md"},
		},
	})
}

type Source struct {
	Path           string
	SuffixPatterns []string
	IgnorePatterns []string
}

func NewNoteStoreWithSources(writeDir string, sources []Source) *NoteStore {
	if len(sources) == 0 {
		sources = []Source{
			{
				Path:           writeDir,
				SuffixPatterns: []string{"*.md"},
			},
		}
	}
	return &NoteStore{
		dir:     writeDir,
		sources: dedupeSources(sources),
	}
}

func (s *NoteStore) EnsureDir() error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	return os.MkdirAll(s.tagIndexDir(), 0o755)
}

func (s *NoteStore) Dir() string {
	return s.dir
}

func (s *NoteStore) SourceDirs() []string {
	dirs := make([]string, 0, len(s.sources))
	for _, source := range s.sources {
		dirs = append(dirs, source.Path)
	}
	return dirs
}

func (s *NoteStore) NotePath(id string) string {
	return filepath.Join(s.dir, id+".md")
}

func (s *NoteStore) LoadAll() ([]*model.Note, error) {
	if err := s.EnsureDir(); err != nil {
		return nil, err
	}
	var notes []*model.Note
	tagMap, err := s.loadTagMap()
	if err != nil {
		return nil, err
	}
	for _, source := range s.sources {
		if err := os.MkdirAll(source.Path, 0o755); err != nil {
			return nil, err
		}
		err := filepath.WalkDir(source.Path, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if path != source.Path && s.pathIgnored(source, source.Path, path, true) {
					return filepath.SkipDir
				}
				return nil
			}
			if !s.ShouldIndexPath(path) {
				return nil
			}
			note, err := s.LoadNote(path)
			if err != nil {
				return nil
			}
			note.Tags = append([]string(nil), tagMap[note.Path]...)
			notes = append(notes, note)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return notes, nil
}

func (s *NoteStore) LoadNote(path string) (*model.Note, error) {
	absPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if filepath.Ext(path) == ".md" {
		note, err := model.ParseNoteBytes(data)
		if err != nil {
			return nil, err
		}
		note.Path = absPath
		return note, nil
	}
	now := time.Now().UTC()
	return &model.Note{
		ID:        s.NoteIDForPath(path),
		Title:     strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
		Tags:      nil,
		Path:      absPath,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: nil,
		Body:      string(data),
	}, nil
}

func (s *NoteStore) CreateNote(title, body string, tags []string, expiresAt *time.Time, now time.Time) (*model.Note, error) {
	id, err := model.NewID()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(title) == "" {
		title = model.TitleFallback(body, now)
	}
	note := &model.Note{
		ID:        id,
		Title:     title,
		Tags:      model.NormalizeTags(tags),
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: expiresAt,
		Body:      body,
	}
	if err := model.SaveNoteAtomic(s.NotePath(id), note); err != nil {
		return nil, err
	}
	note.Path, _ = filepath.Abs(s.NotePath(id))
	if len(note.Tags) > 0 {
		if err := s.SetTagsForPath(note.Path, note.Tags); err != nil {
			return nil, err
		}
	}
	return note, nil
}

func (s *NoteStore) UpdateNote(note *model.Note) error {
	note.UpdatedAt = time.Now().UTC()
	return model.SaveNoteAtomic(s.NotePath(note.ID), note)
}

func (s *NoteStore) DeleteNote(id string) error {
	path := s.NotePath(id)
	absPath, _ := filepath.Abs(path)
	if err := os.Remove(path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return ErrNoteNotFound
		}
		return err
	}
	_ = s.removePathFromAllTags(absPath)
	return nil
}

func (s *NoteStore) PurgeExpired(now time.Time) ([]string, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	var removed []string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		path := filepath.Join(s.dir, entry.Name())
		note, err := s.LoadNote(path)
		if err != nil {
			continue
		}
		if note.ExpiresAt != nil && note.ExpiresAt.Before(now) {
			if err := os.Remove(path); err == nil {
				removed = append(removed, note.ID)
			}
		}
	}
	return removed, nil
}

func (s *NoteStore) ShouldIndexPath(filePath string) bool {
	for _, source := range s.sources {
		if include, ok := s.matchesSource(source, filePath); ok {
			return include
		}
	}
	return false
}

func (s *NoteStore) NoteIDForPath(filePath string) string {
	if strings.HasPrefix(filePath, s.dir+string(os.PathSeparator)) && filepath.Ext(filePath) == ".md" {
		return strings.TrimSuffix(filepath.Base(filePath), ".md")
	}
	sum := sha1.Sum([]byte(filepath.Clean(filePath)))
	return "ext-" + hex.EncodeToString(sum[:8])
}

func (s *NoteStore) SetTagsForPath(notePath string, tags []string) error {
	if strings.TrimSpace(notePath) == "" {
		return nil
	}
	if err := s.EnsureDir(); err != nil {
		return err
	}
	absPath, err := filepath.Abs(filepath.Clean(notePath))
	if err != nil {
		return err
	}
	normalized := model.NormalizeTags(tags)
	if err := s.removePathFromAllTags(absPath); err != nil {
		return err
	}
	for _, tag := range normalized {
		paths, err := s.readTagFile(tag)
		if err != nil {
			return err
		}
		paths = append(paths, absPath)
		if err := s.writeTagFile(tag, dedupeStrings(paths)); err != nil {
			return err
		}
	}
	return nil
}

func (s *NoteStore) matchesSource(source Source, filePath string) (bool, bool) {
	cleanPath := filepath.Clean(filePath)
	root := filepath.Clean(source.Path)
	rel, err := filepath.Rel(root, cleanPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return false, false
	}
	rel = filepath.ToSlash(rel)
	base := path.Base(rel)
	if s.pathIgnored(source, root, cleanPath, false) {
		return false, true
	}
	for _, pattern := range source.SuffixPatterns {
		if model.MatchPattern(pattern, rel, base) {
			return true, true
		}
	}
	return false, true
}

func (s *NoteStore) pathIgnored(source Source, root, fullPath string, isDir bool) bool {
	rel, err := filepath.Rel(root, fullPath)
	if err != nil {
		return false
	}
	rel = filepath.ToSlash(rel)
	if rel == "." {
		return false
	}
	relForDir := rel
	if isDir && !strings.HasSuffix(relForDir, "/") {
		relForDir += "/"
	}
	base := path.Base(rel)
	for _, pattern := range source.IgnorePatterns {
		if model.MatchPattern(pattern, rel, base) || (isDir && model.MatchPattern(pattern, relForDir, base)) {
			return true
		}
	}
	return false
}

func dedupeSources(sources []Source) []Source {
	out := make([]Source, 0, len(sources))
	seen := map[string]struct{}{}
	for _, source := range sources {
		source.Path = filepath.Clean(source.Path)
		if strings.TrimSpace(source.Path) == "" {
			continue
		}
		if _, ok := seen[source.Path]; ok {
			continue
		}
		seen[source.Path] = struct{}{}
		out = append(out, source)
	}
	return out
}

func SourcesFromConfig(cfg model.Config) []Source {
	sources := make([]Source, 0, len(cfg.SearchDirs))
	for _, dir := range cfg.SearchDirs {
		sources = append(sources, Source{
			Path:           dir.Path,
			SuffixPatterns: append([]string(nil), dir.EffectiveSuffixes(cfg.GlobalSuffixes)...),
			IgnorePatterns: append([]string(nil), dir.EffectiveIgnorePaths(cfg.GlobalIgnorePaths)...),
		})
	}
	return sources
}

func (s *NoteStore) tagIndexDir() string {
	return filepath.Join(s.dir, tagIndexDirName)
}

func (s *NoteStore) tagFilePath(tag string) string {
	return filepath.Join(s.tagIndexDir(), sanitizeTag(tag))
}

func sanitizeTag(tag string) string {
	tag = strings.TrimSpace(strings.ToLower(tag))
	if tag == "" {
		return "untagged"
	}
	var b strings.Builder
	for _, r := range tag {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}

func (s *NoteStore) readTagFile(tag string) ([]string, error) {
	data, err := os.ReadFile(s.tagFilePath(tag))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out, nil
}

func (s *NoteStore) writeTagFile(tag string, paths []string) error {
	filePath := s.tagFilePath(tag)
	if len(paths) == 0 {
		if err := os.Remove(filePath); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		return nil
	}
	sort.Strings(paths)
	content := strings.Join(paths, "\n") + "\n"
	return os.WriteFile(filePath, []byte(content), 0o644)
}

func (s *NoteStore) removePathFromAllTags(absPath string) error {
	entries, err := os.ReadDir(s.tagIndexDir())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fullPath := filepath.Join(s.tagIndexDir(), entry.Name())
		data, err := os.ReadFile(fullPath)
		if err != nil {
			return err
		}
		lines := strings.Split(string(data), "\n")
		filtered := make([]string, 0, len(lines))
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || line == absPath {
				continue
			}
			filtered = append(filtered, line)
		}
		if len(filtered) == 0 {
			if err := os.Remove(fullPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return err
			}
			continue
		}
		sort.Strings(filtered)
		content := strings.Join(filtered, "\n") + "\n"
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func (s *NoteStore) loadTagMap() (map[string][]string, error) {
	out := map[string][]string{}
	entries, err := os.ReadDir(s.tagIndexDir())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return out, nil
		}
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		tag := entry.Name()
		data, err := os.ReadFile(filepath.Join(s.tagIndexDir(), entry.Name()))
		if err != nil {
			return nil, err
		}
		for _, line := range strings.Split(string(data), "\n") {
			filePath := strings.TrimSpace(line)
			if filePath == "" {
				continue
			}
			out[filePath] = append(out[filePath], tag)
		}
	}
	for filePath := range out {
		out[filePath] = dedupeStrings(out[filePath])
		sort.Strings(out[filePath])
	}
	return out, nil
}

func dedupeStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
