package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"clio/internal/model"
)

func TestTagFilesAndReload(t *testing.T) {
	dir := t.TempDir()
	st := NewNoteStore(dir)
	note, err := st.CreateNote("Title", "Body", nil, nil, time.Now().UTC())
	if err != nil {
		t.Fatalf("create note failed: %v", err)
	}

	if err := st.SetTagsForPath(note.Path, []string{"work", "team"}); err != nil {
		t.Fatalf("set tags failed: %v", err)
	}
	for _, tag := range []string{"work", "team"} {
		if _, err := os.Stat(filepath.Join(dir, ".clio_tags", tag)); err != nil {
			t.Fatalf("expected tag file for %s: %v", tag, err)
		}
	}

	if err := st.SetTagsForPath(note.Path, []string{"team"}); err != nil {
		t.Fatalf("update tags failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".clio_tags", "work")); err == nil {
		t.Fatalf("expected removed work tag file")
	}

	notes, err := st.LoadAll()
	if err != nil {
		t.Fatalf("load all failed: %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("expected 1 note")
	}
	if len(notes[0].Tags) != 1 || notes[0].Tags[0] != "team" {
		t.Fatalf("expected rehydrated team tag, got %#v", notes[0].Tags)
	}
}

func TestTagHelpers(t *testing.T) {
	dir := t.TempDir()
	st := NewNoteStore(dir)
	if err := st.EnsureDir(); err != nil {
		t.Fatalf("ensure dir failed: %v", err)
	}

	if got := sanitizeTag("Eng/Platform"); !strings.Contains(got, "_") {
		t.Fatalf("expected sanitized tag, got %s", got)
	}

	abs := filepath.Join(dir, "x.md")
	if err := st.writeTagFile("a", []string{abs}); err != nil {
		t.Fatalf("writeTagFile failed: %v", err)
	}
	paths, err := st.readTagFile("a")
	if err != nil || len(paths) != 1 || paths[0] != abs {
		t.Fatalf("unexpected readTagFile result: %#v err=%v", paths, err)
	}

	if err := st.writeTagFile("a", nil); err != nil {
		t.Fatalf("writeTagFile remove failed: %v", err)
	}
	if _, err := os.Stat(st.tagFilePath("a")); err == nil {
		t.Fatalf("expected file removed")
	}
}

func TestEnsureDirTagIndexConflict(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "notes")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".clio_tags"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	st := NewNoteStore(dir)
	if err := st.EnsureDir(); err == nil {
		t.Fatalf("expected ensure dir error")
	}
}

func TestSourceDirsAndDefaults(t *testing.T) {
	st := NewNoteStoreWithSources(t.TempDir(), nil)
	if len(st.SourceDirs()) != 1 {
		t.Fatalf("expected one source dir by default")
	}
	if st.NoteIDForPath(filepath.Join(st.Dir(), "a.md")) != "a" {
		t.Fatalf("expected id from markdown filename")
	}
	if id := st.NoteIDForPath("/tmp/elsewhere.txt"); !strings.HasPrefix(id, "ext-") {
		t.Fatalf("expected hashed external id")
	}
	if st.ShouldIndexPath("/tmp/not-in-source.md") {
		t.Fatalf("did not expect indexing outside source")
	}
}

func TestDeleteNoteRemovesTagReferences(t *testing.T) {
	dir := t.TempDir()
	st := NewNoteStore(dir)
	n := &model.Note{ID: "n1", Title: "t", CreatedAt: time.Now(), UpdatedAt: time.Now(), Body: "b"}
	if err := model.SaveNoteAtomic(st.NotePath(n.ID), n); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	abs, _ := filepath.Abs(st.NotePath(n.ID))
	if err := st.SetTagsForPath(abs, []string{"cleanup"}); err != nil {
		t.Fatalf("set tags failed: %v", err)
	}
	if err := st.DeleteNote(n.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".clio_tags", "cleanup")); err == nil {
		t.Fatalf("expected cleanup tag file removed")
	}
}

func TestCreateNotePersistsInitialTags(t *testing.T) {
	dir := t.TempDir()
	st := NewNoteStore(dir)
	n, err := st.CreateNote("Title", "Body", []string{"A", "A", "b"}, nil, time.Now().UTC())
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, ".clio_tags", "a"))
	if err != nil {
		t.Fatalf("expected tag file: %v", err)
	}
	if !strings.Contains(string(data), n.Path) {
		t.Fatalf("expected path in tag file")
	}
}

func TestSetTagsForPathEdgeCases(t *testing.T) {
	st := NewNoteStore(t.TempDir())
	if err := st.SetTagsForPath(" ", []string{"x"}); err != nil {
		t.Fatalf("expected nil for empty path")
	}

	p := filepath.Join(t.TempDir(), "x.md")
	if err := st.SetTagsForPath(p, []string{"team", "team"}); err != nil {
		t.Fatalf("set tags failed: %v", err)
	}
	paths, err := st.readTagFile("team")
	if err != nil {
		t.Fatalf("read tag failed: %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("expected deduped paths, got %#v", paths)
	}
	if err := st.SetTagsForPath(p, nil); err != nil {
		t.Fatalf("clear tags failed: %v", err)
	}
	if _, err := os.Stat(st.tagFilePath("team")); err == nil {
		t.Fatalf("expected team file removed")
	}
}

func TestLoadTagMapWithoutIndexDir(t *testing.T) {
	st := NewNoteStore(filepath.Join(t.TempDir(), "notes"))
	got, err := st.loadTagMap()
	if err != nil {
		t.Fatalf("loadTagMap failed: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty map")
	}
}

func TestDedupeSourcesAndStrings(t *testing.T) {
	sources := dedupeSources([]Source{
		{Path: "/tmp/a"},
		{Path: "/tmp/a"},
		{Path: " "},
		{Path: "/tmp/b"},
	})
	if len(sources) != 2 {
		t.Fatalf("expected deduped sources")
	}

	got := dedupeStrings([]string{"a", "a", " ", "b"})
	if len(got) != 2 {
		t.Fatalf("expected deduped strings")
	}
}
