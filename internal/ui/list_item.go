package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"clio/internal/model"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

type noteItem struct {
	note  *model.Note
	query string
}

func (n noteItem) Title() string {
	if n.note.Path != "" {
		return filepath.Base(n.note.Path)
	}
	return n.note.Title
}
func (n noteItem) Description() string {
	snippet := previewForQuery(n.note.Body, n.query)
	fullPath := n.note.Path
	if fullPath == "" {
		fullPath = "(path unavailable)"
	}
	return fmt.Sprintf("%s\n%s", snippet, fullPath)
}
func (n noteItem) FilterValue() string { return n.note.Title }

func newNoteDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Bold(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(lipgloss.Color("7"))
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.Bold(true)
	delegate.SetHeight(3)
	delegate.SetSpacing(1)
	return delegate
}

func renderChips(boost, exclude []string) string {
	boostChip := ""
	excludeChip := ""
	if len(boost) > 0 {
		boostChip = "+" + strings.Join(boost, ",+")
	}
	if len(exclude) > 0 {
		excludeChip = "-" + strings.Join(exclude, ",-")
	}
	boostStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	excludeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("204")).Bold(true)
	parts := make([]string, 0, 2)
	if boostChip != "" {
		parts = append(parts, boostStyle.Render(boostChip))
	}
	if excludeChip != "" {
		parts = append(parts, excludeStyle.Render(excludeChip))
	}
	return strings.TrimSpace(strings.Join(parts, "  "))
}

func previewForQuery(body, query string) string {
	content := strings.TrimSpace(body)
	if content == "" {
		return "(empty)"
	}
	if query == "" {
		if len(content) > 120 {
			return content[:120] + "..."
		}
		return content
	}
	lower := strings.ToLower(content)
	q := strings.ToLower(strings.TrimSpace(query))
	idx := strings.Index(lower, q)
	if idx < 0 {
		if len(content) > 120 {
			return content[:120] + "..."
		}
		return content
	}
	start := idx - 40
	if start < 0 {
		start = 0
	}
	end := idx + len(q) + 80
	if end > len(content) {
		end = len(content)
	}
	snippet := strings.TrimSpace(content[start:end])
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(content) {
		snippet += "..."
	}
	return snippet
}

func formatExpiry(note *model.Note) string {
	if note.ExpiresAt == nil {
		return ""
	}
	if note.ExpiresAt.Before(time.Now()) {
		return "expired"
	}
	return note.ExpiresAt.Format(time.RFC3339)
}
