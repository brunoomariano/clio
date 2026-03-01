package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"clio/internal/model"

	tea "github.com/charmbracelet/bubbletea"
)

type fakeProgram struct {
	model  tea.Model
	runErr error
}

func (f fakeProgram) Run() (tea.Model, error) { return f.model, f.runErr }
func (f fakeProgram) Send(tea.Msg)            {}

type dummyModel struct{}

func (d dummyModel) Init() tea.Cmd                           { return nil }
func (d dummyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return d, nil }
func (d dummyModel) View() string                            { return "" }

// TestRunSuccess verifies that run executes successfully with injected dependencies.
// This ensures orchestration completes without watchers and background goroutines.
func TestRunSuccess(t *testing.T) {
	oldLoad := loadConfig
	oldProgram := newProgram
	oldArgs := cliArgs
	defer func() {
		loadConfig = oldLoad
		newProgram = oldProgram
		cliArgs = oldArgs
	}()
	cliArgs = func() []string { return nil }

	dir := t.TempDir()
	loadConfig = func(string) (model.Config, error) {
		return model.Config{
			SearchDirs:        []model.SearchDir{{Path: dir}},
			GlobalSuffixes:    []string{"*.md"},
			GlobalIgnorePaths: []string{"ignore/*", "tests/*"},
			MaxResults:        10,
		}, nil
	}
	newProgram = func(m tea.Model, _ ...tea.ProgramOption) teaProgram {
		return fakeProgram{model: m}
	}

	if err := run(); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

// TestRunLoadConfigError verifies that run returns an error when config loading fails.
func TestRunLoadConfigError(t *testing.T) {
	oldLoad := loadConfig
	oldArgs := cliArgs
	defer func() {
		loadConfig = oldLoad
		cliArgs = oldArgs
	}()
	cliArgs = func() []string { return nil }
	loadConfig = func(string) (model.Config, error) {
		return model.Config{}, errors.New("boom")
	}
	if err := run(); err == nil {
		t.Fatalf("expected error")
	}
}

// TestRunEnsureDirError verifies that run returns an error when notes directory setup fails.
func TestRunEnsureDirError(t *testing.T) {
	oldLoad := loadConfig
	oldArgs := cliArgs
	defer func() {
		loadConfig = oldLoad
		cliArgs = oldArgs
	}()
	cliArgs = func() []string { return nil }
	tmp := t.TempDir()
	file := filepath.Join(tmp, "notes")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	loadConfig = func(string) (model.Config, error) {
		return model.Config{
			SearchDirs:        []model.SearchDir{{Path: file}},
			GlobalSuffixes:    []string{"*.md"},
			GlobalIgnorePaths: []string{"ignore/*", "tests/*"},
		}, nil
	}
	if err := run(); err == nil {
		t.Fatalf("expected error")
	}
}

// TestRunProgramError verifies that UI run errors are propagated.
func TestRunProgramError(t *testing.T) {
	oldLoad := loadConfig
	oldProgram := newProgram
	oldArgs := cliArgs
	defer func() {
		loadConfig = oldLoad
		newProgram = oldProgram
		cliArgs = oldArgs
	}()
	cliArgs = func() []string { return nil }

	dir := t.TempDir()
	loadConfig = func(string) (model.Config, error) {
		return model.Config{
			SearchDirs:        []model.SearchDir{{Path: dir}},
			GlobalSuffixes:    []string{"*.md"},
			GlobalIgnorePaths: []string{"ignore/*", "tests/*"},
			MaxResults:        10,
		}, nil
	}
	newProgram = func(m tea.Model, _ ...tea.ProgramOption) teaProgram {
		return fakeProgram{model: m, runErr: errors.New("ui")}
	}

	if err := run(); err == nil {
		t.Fatalf("expected error")
	}
}

func TestRunIgnoresUnknownFinalModel(t *testing.T) {
	oldLoad := loadConfig
	oldProgram := newProgram
	oldArgs := cliArgs
	defer func() {
		loadConfig = oldLoad
		newProgram = oldProgram
		cliArgs = oldArgs
	}()
	cliArgs = func() []string { return nil }

	dir := t.TempDir()
	loadConfig = func(string) (model.Config, error) {
		return model.Config{
			SearchDirs:        []model.SearchDir{{Path: dir}},
			GlobalSuffixes:    []string{"*.md"},
			GlobalIgnorePaths: []string{"ignore/*", "tests/*"},
			MaxResults:        10,
		}, nil
	}
	newProgram = func(m tea.Model, _ ...tea.ProgramOption) teaProgram {
		return fakeProgram{model: dummyModel{}}
	}
	if err := run(); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}
