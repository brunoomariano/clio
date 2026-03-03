package main

import (
	"os"
	"path/filepath"
	"testing"

	"clio/internal/model"
)

func TestParsePatternListCSVAndInvalidJSON(t *testing.T) {
	values, err := parsePatternList("*.md,*.json")
	if err != nil || len(values) != 2 {
		t.Fatalf("unexpected parse result: %#v err=%v", values, err)
	}

	if _, err := parsePatternList("["); err == nil {
		t.Fatalf("expected invalid json error")
	}
}

func TestRunEditor(t *testing.T) {
	err := runEditor(model.Config{Editor: "/bin/true"}, "/tmp/nonexistent")
	if err != nil {
		t.Fatalf("expected /bin/true to run: %v", err)
	}
}

func TestRunEditorFallbackToNano(t *testing.T) {
	dir := t.TempDir()
	nanoPath := filepath.Join(dir, "nano")
	if err := os.WriteFile(nanoPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write nano failed: %v", err)
	}
	oldPath := os.Getenv("PATH")
	if err := os.Setenv("PATH", dir); err != nil {
		t.Fatalf("set PATH failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Setenv("PATH", oldPath) })

	if err := runEditor(model.Config{Editor: "/not/found/editor"}, "/tmp/any"); err != nil {
		t.Fatalf("expected fallback nano to run: %v", err)
	}
}

func TestRunEditorUsesEditorEnv(t *testing.T) {
	oldEditor := os.Getenv("EDITOR")
	if err := os.Setenv("EDITOR", "/bin/true"); err != nil {
		t.Fatalf("set EDITOR failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Setenv("EDITOR", oldEditor) })
	if err := runEditor(model.Config{}, "/tmp/any"); err != nil {
		t.Fatalf("expected EDITOR env to run: %v", err)
	}
}

func TestParseRunOptionsUnexpectedArgs(t *testing.T) {
	if _, err := parseRunOptions([]string{"--cwd", "extra"}); err == nil {
		t.Fatalf("expected error for unexpected args")
	}
	if _, err := parseRunOptions([]string{"--ignore_paths=['x']"}); err == nil {
		t.Fatalf("expected error when --cwd is missing")
	}
}

func TestApplyRunOptionsNoCWD(t *testing.T) {
	cfg := model.Config{SearchDirs: []model.SearchDir{{Path: "/tmp/a"}}}
	if err := applyRunOptions(&cfg, runOptions{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.SearchDirs) != 1 || cfg.SearchDirs[0].Path != "/tmp/a" {
		t.Fatalf("expected unchanged search dirs")
	}
}

func TestParsePatternListEmpty(t *testing.T) {
	values, err := parsePatternList("   ")
	if err != nil || values != nil {
		t.Fatalf("expected nil list on empty input")
	}
}
