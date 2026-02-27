package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"clio/internal/model"
	"clio/internal/store"

	tea "github.com/charmbracelet/bubbletea"
)

type fakeProgram struct {
	runErr error
}

func (f fakeProgram) Run() (tea.Model, error) { return nil, f.runErr }
func (f fakeProgram) Send(tea.Msg)            {}

type fakeTicker struct {
	ch chan time.Time
}

func (f fakeTicker) Chan() <-chan time.Time { return f.ch }
func (f fakeTicker) Stop()                  {}

// TestRunSuccess verifies that run executes successfully with injected dependencies.
// This ensures the main orchestration can complete without errors.
func TestRunSuccess(t *testing.T) {
	oldLoad := loadConfig
	oldWatcher := startWatcher
	oldProgram := newProgram
	oldTicker := newTicker
	defer func() {
		loadConfig = oldLoad
		startWatcher = oldWatcher
		newProgram = oldProgram
		newTicker = oldTicker
	}()

	dir := t.TempDir()
	loadConfig = func(string) (model.Config, error) {
		return model.Config{NotesDir: dir, MaxResults: 10}, nil
	}
	startWatcher = func(string) (<-chan store.WatchEvent, func() error, error) {
		ch := make(chan store.WatchEvent)
		close(ch)
		return ch, func() error { return nil }, nil
	}
	newProgram = func(tea.Model, ...tea.ProgramOption) teaProgram {
		return fakeProgram{}
	}
	newTicker = func(time.Duration) ticker {
		ch := make(chan time.Time)
		close(ch)
		return fakeTicker{ch: ch}
	}

	if err := run(); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

// TestRunLoadConfigError verifies that run returns an error when config loading fails.
// This ensures startup errors are propagated.
func TestRunLoadConfigError(t *testing.T) {
	oldLoad := loadConfig
	defer func() { loadConfig = oldLoad }()
	loadConfig = func(string) (model.Config, error) {
		return model.Config{}, errors.New("boom")
	}
	if err := run(); err == nil {
		t.Fatalf("expected error")
	}
}

// TestRunEnsureDirError verifies that run returns an error when notes directory setup fails.
// This ensures filesystem failures abort startup.
func TestRunEnsureDirError(t *testing.T) {
	oldLoad := loadConfig
	defer func() { loadConfig = oldLoad }()
	tmp := t.TempDir()
	file := filepath.Join(tmp, "notes")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	loadConfig = func(string) (model.Config, error) {
		return model.Config{NotesDir: file}, nil
	}
	if err := run(); err == nil {
		t.Fatalf("expected error")
	}
}

// TestRunLoadAllError verifies that run returns an error when reading notes fails.
// This ensures permission issues in the notes directory are surfaced.
func TestRunLoadAllError(t *testing.T) {
	oldLoad := loadConfig
	oldWatcher := startWatcher
	oldProgram := newProgram
	oldTicker := newTicker
	defer func() {
		loadConfig = oldLoad
		startWatcher = oldWatcher
		newProgram = oldProgram
		newTicker = oldTicker
	}()

	dir := t.TempDir()
	if err := os.Chmod(dir, 0o000); err != nil {
		t.Fatalf("chmod failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	loadConfig = func(string) (model.Config, error) {
		return model.Config{NotesDir: dir}, nil
	}
	startWatcher = func(string) (<-chan store.WatchEvent, func() error, error) {
		ch := make(chan store.WatchEvent)
		close(ch)
		return ch, func() error { return nil }, nil
	}
	newProgram = func(tea.Model, ...tea.ProgramOption) teaProgram { return fakeProgram{} }
	newTicker = func(time.Duration) ticker {
		ch := make(chan time.Time)
		close(ch)
		return fakeTicker{ch: ch}
	}

	if err := run(); err == nil {
		t.Fatalf("expected error")
	}
}

// TestRunWatcherError verifies that watcher startup errors are propagated.
// This ensures filesystem watcher failures are handled.
func TestRunWatcherError(t *testing.T) {
	oldLoad := loadConfig
	oldWatcher := startWatcher
	oldProgram := newProgram
	defer func() {
		loadConfig = oldLoad
		startWatcher = oldWatcher
		newProgram = oldProgram
	}()

	dir := t.TempDir()
	loadConfig = func(string) (model.Config, error) {
		return model.Config{NotesDir: dir, MaxResults: 10}, nil
	}
	startWatcher = func(string) (<-chan store.WatchEvent, func() error, error) {
		return nil, nil, errors.New("watcher")
	}
	newProgram = func(tea.Model, ...tea.ProgramOption) teaProgram { return fakeProgram{} }

	if err := run(); err == nil {
		t.Fatalf("expected error")
	}
}

// TestRunProgramError verifies that UI run errors are propagated.
// This ensures UI failures abort startup correctly.
func TestRunProgramError(t *testing.T) {
	oldLoad := loadConfig
	oldWatcher := startWatcher
	oldProgram := newProgram
	oldTicker := newTicker
	defer func() {
		loadConfig = oldLoad
		startWatcher = oldWatcher
		newProgram = oldProgram
		newTicker = oldTicker
	}()

	dir := t.TempDir()
	loadConfig = func(string) (model.Config, error) {
		return model.Config{NotesDir: dir, MaxResults: 10}, nil
	}
	startWatcher = func(string) (<-chan store.WatchEvent, func() error, error) {
		ch := make(chan store.WatchEvent)
		close(ch)
		return ch, func() error { return nil }, nil
	}
	newProgram = func(tea.Model, ...tea.ProgramOption) teaProgram {
		return fakeProgram{runErr: errors.New("ui")}
	}
	newTicker = func(time.Duration) ticker {
		ch := make(chan time.Time)
		close(ch)
		return fakeTicker{ch: ch}
	}

	if err := run(); err == nil {
		t.Fatalf("expected error")
	}
}
