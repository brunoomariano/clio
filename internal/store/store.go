package store

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"clio/internal/model"
)

var (
	ErrNoteNotFound = errors.New("note not found")
)

type NoteStore struct {
	dir string
}

func NewNoteStore(dir string) *NoteStore {
	return &NoteStore{dir: dir}
}

func (s *NoteStore) EnsureDir() error {
	return os.MkdirAll(s.dir, 0o755)
}

func (s *NoteStore) Dir() string {
	return s.dir
}

func (s *NoteStore) NotePath(id string) string {
	return filepath.Join(s.dir, id+".md")
}

func (s *NoteStore) LoadAll() ([]*model.Note, error) {
	if err := s.EnsureDir(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	var notes []*model.Note
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		path := filepath.Join(s.dir, entry.Name())
		note, err := s.LoadNote(path)
		if err != nil {
			continue
		}
		notes = append(notes, note)
	}
	return notes, nil
}

func (s *NoteStore) LoadNote(path string) (*model.Note, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return model.ParseNoteBytes(data)
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
	return note, nil
}

func (s *NoteStore) UpdateNote(note *model.Note) error {
	note.Tags = model.NormalizeTags(note.Tags)
	note.UpdatedAt = time.Now().UTC()
	return model.SaveNoteAtomic(s.NotePath(note.ID), note)
}

func (s *NoteStore) DeleteNote(id string) error {
	path := s.NotePath(id)
	if err := os.Remove(path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return ErrNoteNotFound
		}
		return err
	}
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
