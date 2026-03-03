package main

import (
	"reflect"
	"testing"

	"clio/internal/model"
)

// TestParseRunOptionsCWD verifies --cwd with overrides is parsed.
// This ensures CLI overrides are accepted in documented format.
func TestParseRunOptionsCWD(t *testing.T) {
	opts, err := parseRunOptions([]string{
		"--cwd",
		"--suffixes=['*.md','*.json']",
		"--ignore_paths=['test.*']",
	})
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if !opts.cwd {
		t.Fatalf("expected cwd option enabled")
	}
	gotSuffixes, err := parsePatternList(opts.suffixesRaw)
	if err != nil {
		t.Fatalf("suffix parse failed: %v", err)
	}
	if !reflect.DeepEqual(gotSuffixes, []string{"*.md", "*.json"}) {
		t.Fatalf("unexpected suffixes: %#v", gotSuffixes)
	}
}

// TestParseRunOptionsRequiresCWD verifies overrides are rejected without --cwd.
// This ensures ambiguous CLI usage fails clearly.
func TestParseRunOptionsRequiresCWD(t *testing.T) {
	_, err := parseRunOptions([]string{"--suffixes=['*.md']"})
	if err == nil {
		t.Fatalf("expected error")
	}
}

// TestApplyRunOptionsCWD verifies config sources are replaced by current directory.
// This ensures --cwd limits search to the working directory.
func TestApplyRunOptionsCWD(t *testing.T) {
	oldCwd := currentDir
	defer func() { currentDir = oldCwd }()
	currentDir = func() (string, error) { return "/tmp/project", nil }

	cfg := model.Config{
		SearchDirs:        []model.SearchDir{{Path: "/tmp/notes"}},
		GlobalSuffixes:    []string{"*.md", "*.txt"},
		GlobalIgnorePaths: []string{"tests/*"},
	}
	err := applyRunOptions(&cfg, runOptions{
		cwd:         true,
		suffixesRaw: "['*.md','*.json']",
		ignoresRaw:  "['test.*']",
	})
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	want := model.SearchDir{
		Path:        "/tmp/project",
		Suffixes:    []string{"*.md", "*.json"},
		IgnorePaths: []string{"test.*"},
	}
	if !reflect.DeepEqual(cfg.SearchDirs, []model.SearchDir{want}) {
		t.Fatalf("unexpected search dirs: %#v", cfg.SearchDirs)
	}
}

// TestApplyRunOptionsCWDGlobalFallback verifies --cwd without overrides uses globals.
// This ensures global filters continue as defaults for cwd mode.
func TestApplyRunOptionsCWDGlobalFallback(t *testing.T) {
	oldCwd := currentDir
	defer func() { currentDir = oldCwd }()
	currentDir = func() (string, error) { return "/tmp/project", nil }

	cfg := model.Config{
		GlobalSuffixes:    []string{"*.md"},
		GlobalIgnorePaths: []string{"tests/*"},
	}
	err := applyRunOptions(&cfg, runOptions{cwd: true})
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	if len(cfg.SearchDirs) != 1 || cfg.SearchDirs[0].Path != "/tmp/project" {
		t.Fatalf("unexpected search dirs: %#v", cfg.SearchDirs)
	}
	if len(cfg.SearchDirs[0].Suffixes) != 0 || len(cfg.SearchDirs[0].IgnorePaths) != 0 {
		t.Fatalf("expected no per-dir overrides, got %#v", cfg.SearchDirs[0])
	}
}
