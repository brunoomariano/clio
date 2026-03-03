package model

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type SearchDir struct {
	Path        string   `yaml:"path"`
	Suffixes    []string `yaml:"suffixes,omitempty"`
	IgnorePaths []string `yaml:"ignore_paths,omitempty"`
}

type Config struct {
	SearchDirs        []SearchDir `yaml:"search_dirs"`
	GlobalSuffixes    []string    `yaml:"global_suffixes"`
	GlobalIgnorePaths []string    `yaml:"global_ignore_paths"`
	BM25K1            float64     `yaml:"bm25_k1"`
	BM25B             float64     `yaml:"bm25_b"`
	BoostWeight       float64     `yaml:"boost_weight"`
	DebounceMS        int         `yaml:"debounce_ms"`
	MaxResults        int         `yaml:"max_results"`
	Editor            string      `yaml:"editor"`
	Terminal          string      `yaml:"terminal"`

	LegacyNotesDir string `yaml:"notes_dir,omitempty"`
}

func DefaultConfig(home string) Config {
	return Config{
		SearchDirs: []SearchDir{
			{Path: "~/.local/share/clio/notes"},
		},
		GlobalSuffixes: []string{
			"*.md",
			"*.txt",
			"*.json",
			"*.yaml",
		},
		GlobalIgnorePaths: []string{
			"ignore/*",
			"tests/*",
		},
		BM25K1:      1.2,
		BM25B:       0.75,
		BoostWeight: 2.0,
		DebounceMS:  100,
		MaxResults:  200,
		Editor:      "nvim",
		Terminal:    "",
	}
}

func ExpandPath(path string, home string) string {
	if path == "~" {
		return home
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, strings.TrimPrefix(path, "~/"))
	}
	if strings.HasPrefix(path, "~") {
		return filepath.Join(home, strings.TrimPrefix(path, "~"))
	}
	return path
}

func LoadOrCreateConfig(path string) (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}
	defaultCfg := DefaultConfig(home)
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return Config{}, err
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return Config{}, err
		}
		data, err := yaml.Marshal(defaultCfg)
		if err != nil {
			return Config{}, err
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return Config{}, err
		}
		return finalizeConfig(defaultCfg, home), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	cfg := defaultCfg
	cfg.SearchDirs = nil
	cfg.GlobalSuffixes = nil
	cfg.GlobalIgnorePaths = nil
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return finalizeConfig(cfg, home), nil
}

func (c Config) PrimarySearchDir() string {
	if len(c.SearchDirs) == 0 {
		return ""
	}
	return c.SearchDirs[0].Path
}

func (d SearchDir) EffectiveSuffixes(global []string) []string {
	if len(d.Suffixes) > 0 {
		return d.Suffixes
	}
	return global
}

func (d SearchDir) EffectiveIgnorePaths(global []string) []string {
	if len(d.IgnorePaths) > 0 {
		return d.IgnorePaths
	}
	return global
}

func finalizeConfig(cfg Config, home string) Config {
	if len(cfg.SearchDirs) == 0 && strings.TrimSpace(cfg.LegacyNotesDir) != "" {
		cfg.SearchDirs = []SearchDir{{Path: cfg.LegacyNotesDir}}
	}

	for i := range cfg.SearchDirs {
		cfg.SearchDirs[i].Path = ExpandPath(cfg.SearchDirs[i].Path, home)
		cfg.SearchDirs[i].Suffixes = normalizePatterns(cfg.SearchDirs[i].Suffixes)
		cfg.SearchDirs[i].IgnorePaths = normalizePatterns(cfg.SearchDirs[i].IgnorePaths)
	}

	cfg.GlobalSuffixes = normalizePatterns(cfg.GlobalSuffixes)
	cfg.GlobalIgnorePaths = normalizePatterns(cfg.GlobalIgnorePaths)

	if len(cfg.SearchDirs) == 0 {
		cfg.SearchDirs = []SearchDir{{Path: ExpandPath("~/.local/share/clio/notes", home)}}
	}
	if len(cfg.GlobalSuffixes) == 0 {
		cfg.GlobalSuffixes = []string{"*.md", "*.txt", "*.json", "*.yaml"}
	}
	if len(cfg.GlobalIgnorePaths) == 0 {
		cfg.GlobalIgnorePaths = []string{"ignore/*", "tests/*"}
	}
	return cfg
}

func normalizePatterns(patterns []string) []string {
	out := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		trimmed := strings.TrimSpace(pattern)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func MatchPattern(pattern, value, base string) bool {
	if pattern == "" {
		return false
	}
	if re, err := regexp.Compile(pattern); err == nil {
		return re.MatchString(value) || (base != "" && re.MatchString(base))
	}

	targets := []string{value}
	if base != "" {
		targets = append(targets, base)
	}
	for _, t := range targets {
		if ok, err := path.Match(pattern, t); err == nil && ok {
			return true
		}
	}
	return false
}
