package model

import "testing"

func TestSearchDirEffectiveOverrides(t *testing.T) {
	d := SearchDir{}
	if got := d.EffectiveSuffixes([]string{"*.md"}); len(got) != 1 || got[0] != "*.md" {
		t.Fatalf("expected global suffix fallback")
	}
	if got := d.EffectiveIgnorePaths([]string{"tests/*"}); len(got) != 1 || got[0] != "tests/*" {
		t.Fatalf("expected global ignore fallback")
	}

	d.Suffixes = []string{"*.txt"}
	d.IgnorePaths = []string{"tmp/*"}
	if got := d.EffectiveSuffixes([]string{"*.md"}); got[0] != "*.txt" {
		t.Fatalf("expected local suffix override")
	}
	if got := d.EffectiveIgnorePaths([]string{"tests/*"}); got[0] != "tmp/*" {
		t.Fatalf("expected local ignore override")
	}
}

func TestMatchPatternRegexAndGlob(t *testing.T) {
	if !MatchPattern(`.*\.md$`, "docs/file.md", "file.md") {
		t.Fatalf("expected regex match")
	}
	if MatchPattern(`.*\.md$`, "docs/file.txt", "file.txt") {
		t.Fatalf("did not expect regex match")
	}
	if !MatchPattern("*.txt", "docs/file.txt", "file.txt") {
		t.Fatalf("expected glob match")
	}
	if MatchPattern("*.txt", "docs/file.md", "file.md") {
		t.Fatalf("did not expect glob mismatch")
	}
}

func TestPrimarySearchDirEmpty(t *testing.T) {
	var cfg Config
	if got := cfg.PrimarySearchDir(); got != "" {
		t.Fatalf("expected empty primary dir")
	}
}
