package model

import (
	"errors"
	"testing"
	"time"
)

type fakeTempFile struct {
	name        string
	writeErr    error
	syncErr     error
	closeErr    error
	writeCalled bool
}

func (f *fakeTempFile) Write(p []byte) (int, error) {
	f.writeCalled = true
	if f.writeErr != nil {
		return 0, f.writeErr
	}
	return len(p), nil
}
func (f *fakeTempFile) Sync() error  { return f.syncErr }
func (f *fakeTempFile) Close() error { return f.closeErr }
func (f *fakeTempFile) Name() string { return f.name }

// TestSaveNoteAtomicWriteFailure verifies that write errors are returned.
// This ensures partial writes don't silently succeed.
func TestSaveNoteAtomicWriteFailure(t *testing.T) {
	oldCreate := createTempFile
	defer func() { createTempFile = oldCreate }()

	createTempFile = func(dir, pattern string) (tempFile, error) {
		return &fakeTempFile{name: dir + "/tmp", writeErr: errors.New("write")}, nil
	}
	note := &Note{ID: "x", Title: "t", CreatedAt: time.Now(), UpdatedAt: time.Now(), Body: "b"}
	err := SaveNoteAtomic("/tmp/x.md", note)
	if err == nil {
		t.Fatalf("expected error")
	}
}

// TestSaveNoteAtomicSyncFailure verifies that sync errors are returned.
// This ensures fsync failures don't corrupt notes.
func TestSaveNoteAtomicSyncFailure(t *testing.T) {
	oldCreate := createTempFile
	defer func() { createTempFile = oldCreate }()

	createTempFile = func(dir, pattern string) (tempFile, error) {
		return &fakeTempFile{name: dir + "/tmp", syncErr: errors.New("sync")}, nil
	}
	note := &Note{ID: "x", Title: "t", CreatedAt: time.Now(), UpdatedAt: time.Now(), Body: "b"}
	err := SaveNoteAtomic("/tmp/x.md", note)
	if err == nil {
		t.Fatalf("expected error")
	}
}

// TestSaveNoteAtomicCloseFailure verifies that close errors are returned.
// This ensures close failures are surfaced to the caller.
func TestSaveNoteAtomicCloseFailure(t *testing.T) {
	oldCreate := createTempFile
	defer func() { createTempFile = oldCreate }()

	createTempFile = func(dir, pattern string) (tempFile, error) {
		return &fakeTempFile{name: dir + "/tmp", closeErr: errors.New("close")}, nil
	}
	note := &Note{ID: "x", Title: "t", CreatedAt: time.Now(), UpdatedAt: time.Now(), Body: "b"}
	err := SaveNoteAtomic("/tmp/x.md", note)
	if err == nil {
		t.Fatalf("expected error")
	}
}

// TestSaveNoteAtomicRenameFailure verifies that rename errors are returned.
// This ensures rename failures are propagated.
func TestSaveNoteAtomicRenameFailure(t *testing.T) {
	oldCreate := createTempFile
	oldRename := renameFile
	defer func() {
		createTempFile = oldCreate
		renameFile = oldRename
	}()

	createTempFile = func(dir, pattern string) (tempFile, error) {
		return &fakeTempFile{name: dir + "/tmp"}, nil
	}
	renameFile = func(string, string) error { return errors.New("rename") }
	note := &Note{ID: "x", Title: "t", CreatedAt: time.Now(), UpdatedAt: time.Now(), Body: "b"}
	err := SaveNoteAtomic("/tmp/x.md", note)
	if err == nil {
		t.Fatalf("expected error")
	}
}
