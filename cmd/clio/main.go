package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"clio/internal/index"
	"clio/internal/model"
	"clio/internal/store"
	"clio/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

type teaProgram interface {
	Run() (tea.Model, error)
	Send(tea.Msg)
}

var runApp = run
var exitFn = os.Exit
var loadConfig = model.LoadOrCreateConfig
var userHomeDir = os.UserHomeDir
var currentDir = os.Getwd
var cliArgs = func() []string { return os.Args[1:] }
var newProgram = func(m tea.Model, opts ...tea.ProgramOption) teaProgram {
	return tea.NewProgram(m, opts...)
}

func main() {
	if err := runApp(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run: %v\n", err)
		exitFn(1)
	}
}

func run() error {
	opts, err := parseRunOptions(cliArgs())
	if err != nil {
		return fmt.Errorf("parse args: %w", err)
	}

	configPath := defaultConfigPath()
	cfg, err := loadConfig(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if err := applyRunOptions(&cfg, opts); err != nil {
		return fmt.Errorf("apply args: %w", err)
	}
	sources := store.SourcesFromConfig(cfg)
	st := store.NewNoteStoreWithSources(cfg.PrimarySearchDir(), sources)
	if err := st.EnsureDir(); err != nil {
		return fmt.Errorf("ensure notes dir: %w", err)
	}

	notes, err := st.LoadAll()
	if err != nil {
		return fmt.Errorf("load notes: %w", err)
	}

	idx := index.NewIndex()
	idx.Reset(notes)

	modelUI := ui.New(cfg, st, idx)
	program := newProgram(modelUI, tea.WithAltScreen())

	finalModel, err := program.Run()
	if err != nil {
		return fmt.Errorf("run ui: %w", err)
	}
	viewModel, ok := finalModel.(*ui.Model)
	if !ok {
		return nil
	}

	if editPath := strings.TrimSpace(viewModel.PendingEditorPath()); editPath != "" {
		if err := runEditor(cfg, editPath); err != nil {
			return err
		}
	}
	return nil
}

func defaultConfigPath() string {
	home, err := userHomeDir()
	if err != nil {
		return ".clio.yaml"
	}
	return filepath.Join(home, ".config", "clio.yaml")
}

type runOptions struct {
	cwd         bool
	suffixesRaw string
	ignoresRaw  string
}

func parseRunOptions(args []string) (runOptions, error) {
	fs := flag.NewFlagSet("clio", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var opts runOptions
	fs.BoolVar(&opts.cwd, "cwd", false, "index only the current working directory")
	fs.StringVar(&opts.suffixesRaw, "suffixes", "", "override suffix patterns for --cwd")
	fs.StringVar(&opts.ignoresRaw, "ignore_paths", "", "override ignore path patterns for --cwd")

	if err := fs.Parse(args); err != nil {
		return runOptions{}, err
	}
	if len(fs.Args()) > 0 {
		return runOptions{}, fmt.Errorf("unexpected arguments: %v", fs.Args())
	}
	if !opts.cwd && (opts.suffixesRaw != "" || opts.ignoresRaw != "") {
		return runOptions{}, errors.New("--suffixes and --ignore_paths require --cwd")
	}
	return opts, nil
}

func applyRunOptions(cfg *model.Config, opts runOptions) error {
	if !opts.cwd {
		return nil
	}
	cwd, err := currentDir()
	if err != nil {
		return err
	}

	dir := model.SearchDir{Path: cwd}
	if opts.suffixesRaw != "" {
		patterns, err := parsePatternList(opts.suffixesRaw)
		if err != nil {
			return fmt.Errorf("parse --suffixes: %w", err)
		}
		dir.Suffixes = patterns
	}
	if opts.ignoresRaw != "" {
		patterns, err := parsePatternList(opts.ignoresRaw)
		if err != nil {
			return fmt.Errorf("parse --ignore_paths: %w", err)
		}
		dir.IgnorePaths = patterns
	}
	cfg.SearchDirs = []model.SearchDir{dir}
	return nil
}

func parsePatternList(raw string) ([]string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}

	// Accept JSON arrays and shell-friendly forms like ['*.md','*.json'].
	jsonCandidate := strings.ReplaceAll(trimmed, "'", "\"")
	if strings.HasPrefix(jsonCandidate, "[") {
		var out []string
		if err := json.Unmarshal([]byte(jsonCandidate), &out); err != nil {
			return nil, err
		}
		return normalizeList(out), nil
	}
	return normalizeList(strings.Split(trimmed, ",")), nil
}

func normalizeList(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		item = strings.Trim(item, "'\"")
		if item == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}

func runEditor(cfg model.Config, path string) error {
	editor := strings.TrimSpace(cfg.Editor)
	if editor == "" {
		editor = strings.TrimSpace(os.Getenv("EDITOR"))
	}
	if editor == "" {
		editor = "nvim"
	}
	if _, err := exec.LookPath(editor); err != nil {
		editor = "nano"
	}
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
