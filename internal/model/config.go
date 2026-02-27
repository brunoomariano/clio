package model

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	NotesDir    string  `yaml:"notes_dir"`
	BM25K1      float64 `yaml:"bm25_k1"`
	BM25B       float64 `yaml:"bm25_b"`
	BoostWeight float64 `yaml:"boost_weight"`
	DebounceMS  int     `yaml:"debounce_ms"`
	MaxResults  int     `yaml:"max_results"`
}

func DefaultConfig(home string) Config {
	return Config{
		NotesDir:    "~/.local/share/clio/notes",
		BM25K1:      1.2,
		BM25B:       0.75,
		BoostWeight: 2.0,
		DebounceMS:  100,
		MaxResults:  200,
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
		defaultCfg.NotesDir = ExpandPath(defaultCfg.NotesDir, home)
		return defaultCfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	cfg := defaultCfg
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	cfg.NotesDir = ExpandPath(cfg.NotesDir, home)
	return cfg, nil
}
