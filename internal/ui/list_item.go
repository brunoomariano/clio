package ui

import (
	"fmt"
	"strings"
	"time"

	"clio/internal/model"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

type noteItem struct {
	note *model.Note
}

func (n noteItem) Title() string { return n.note.Title }
func (n noteItem) Description() string {
	snippet := strings.TrimSpace(n.note.Body)
	if len(snippet) > 80 {
		snippet = snippet[:80] + "..."
	}
	tags := strings.Join(n.note.Tags, ", ")
	expiry := ""
	if n.note.ExpiresAt != nil {
		expiry = " · expires " + n.note.ExpiresAt.Format("2006-01-02")
	}
	return fmt.Sprintf("%s\n[%s]%s", snippet, tags, expiry)
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
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	return style.Render(strings.TrimSpace(strings.Join([]string{boostChip, excludeChip}, "  ")))
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
