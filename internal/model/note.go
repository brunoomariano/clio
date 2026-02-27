package model

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	FrontmatterDelimiter = "---"
)

var (
	ErrMissingFrontmatter = errors.New("missing frontmatter")
)

type Note struct {
	ID        string
	Title     string
	Tags      []string
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiresAt *time.Time
	Body      string
}

type frontmatter struct {
	ID        string   `yaml:"id"`
	Title     string   `yaml:"title"`
	Tags      []string `yaml:"tags"`
	CreatedAt string   `yaml:"created_at"`
	UpdatedAt string   `yaml:"updated_at"`
	ExpiresAt *string  `yaml:"expires_at"`
}

func NewID() (string, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	enc := base32.StdEncoding.WithPadding(base32.NoPadding)
	return strings.ToLower(enc.EncodeToString(buf))[:10], nil
}

func NormalizeTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		t = strings.ToLower(t)
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

func TitleFallback(body string, now time.Time) string {
	scanner := bufio.NewScanner(strings.NewReader(body))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			return line
		}
	}
	return now.Format("2006-01-02 15:04")
}

func ParseNote(r io.Reader) (*Note, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return ParseNoteBytes(data)
}

func ParseNoteBytes(data []byte) (*Note, error) {
	content := string(data)
	if !strings.HasPrefix(content, FrontmatterDelimiter+"\n") {
		return nil, ErrMissingFrontmatter
	}
	parts := strings.SplitN(content, "\n"+FrontmatterDelimiter+"\n", 2)
	if len(parts) != 2 {
		return nil, ErrMissingFrontmatter
	}
	frontmatterRaw := strings.TrimPrefix(parts[0], FrontmatterDelimiter+"\n")
	body := parts[1]

	var fm frontmatter
	if err := yaml.Unmarshal([]byte(frontmatterRaw), &fm); err != nil {
		return nil, err
	}

	createdAt, err := time.Parse(time.RFC3339, fm.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, fm.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse updated_at: %w", err)
	}

	var expiresAt *time.Time
	if fm.ExpiresAt != nil && *fm.ExpiresAt != "" {
		parsed, err := time.Parse(time.RFC3339, *fm.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("parse expires_at: %w", err)
		}
		expiresAt = &parsed
	}

	note := &Note{
		ID:        fm.ID,
		Title:     fm.Title,
		Tags:      NormalizeTags(fm.Tags),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		ExpiresAt: expiresAt,
		Body:      strings.TrimLeft(body, "\n"),
	}
	return note, nil
}

func RenderNote(note *Note) ([]byte, error) {
	fm := frontmatter{
		ID:        note.ID,
		Title:     note.Title,
		Tags:      NormalizeTags(note.Tags),
		CreatedAt: note.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: note.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if note.ExpiresAt != nil {
		ex := note.ExpiresAt.UTC().Format(time.RFC3339)
		fm.ExpiresAt = &ex
	}
	buf := &bytes.Buffer{}
	buf.WriteString(FrontmatterDelimiter + "\n")
	enc := yaml.NewEncoder(buf)
	enc.SetIndent(2)
	if err := enc.Encode(fm); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	buf.WriteString(FrontmatterDelimiter + "\n\n")
	buf.WriteString(strings.TrimRight(note.Body, "\n"))
	buf.WriteString("\n")
	return buf.Bytes(), nil
}

func SaveNoteAtomic(path string, note *Note) error {
	data, err := RenderNote(note)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	file, err := createTempFile(dir, ".clio_tmp_*")
	if err != nil {
		return err
	}
	tmpName := file.Name()
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return renameFile(tmpName, path)
}

type tempFile interface {
	Write([]byte) (int, error)
	Sync() error
	Close() error
	Name() string
}

var createTempFile = func(dir, pattern string) (tempFile, error) {
	return os.CreateTemp(dir, pattern)
}

var renameFile = os.Rename
