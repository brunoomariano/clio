package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

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

type ticker interface {
	Chan() <-chan time.Time
	Stop()
}

type realTicker struct {
	t *time.Ticker
}

func (r realTicker) Chan() <-chan time.Time { return r.t.C }
func (r realTicker) Stop()                  { r.t.Stop() }

var runApp = run
var exitFn = os.Exit
var loadConfig = model.LoadOrCreateConfig
var startWatcher = store.StartWatcher
var userHomeDir = os.UserHomeDir
var newProgram = func(m tea.Model, opts ...tea.ProgramOption) teaProgram {
	return tea.NewProgram(m, opts...)
}
var newTicker = func(d time.Duration) ticker {
	return realTicker{t: time.NewTicker(d)}
}

func main() {
	if err := runApp(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run: %v\n", err)
		exitFn(1)
	}
}

func run() error {
	configPath := defaultConfigPath()
	cfg, err := loadConfig(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	st := store.NewNoteStore(cfg.NotesDir)
	if err := st.EnsureDir(); err != nil {
		return fmt.Errorf("ensure notes dir: %w", err)
	}

	notes, err := st.LoadAll()
	if err != nil {
		return fmt.Errorf("load notes: %w", err)
	}

	idx := index.NewIndex()
	idx.Reset(notes)

	now := time.Now().UTC()
	if removed, err := st.PurgeExpired(now); err == nil && len(removed) > 0 {
		for _, id := range removed {
			idx.Remove(id)
		}
	}

	modelUI := ui.New(cfg, st, idx)
	program := newProgram(modelUI, tea.WithAltScreen())

	watchCh, closeWatch, err := startWatcher(st.Dir())
	if err != nil {
		return fmt.Errorf("start watcher: %w", err)
	}
	defer func() { _ = closeWatch() }()

	done := make(chan struct{})
	var wg sync.WaitGroup
	defer func() {
		close(done)
		wg.Wait()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for ev := range watchCh {
			select {
			case <-done:
				return
			default:
			}
			if ev.Err != nil {
				program.Send(ui.WatcherMsg{Err: ev.Err})
				continue
			}
			program.Send(ui.WatcherMsg{Path: ev.Path, Op: ev.Op.String()})
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := newTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.Chan():
				if removed, err := st.PurgeExpired(time.Now().UTC()); err == nil {
					for _, id := range removed {
						idx.Remove(id)
					}
					program.Send(ui.WatcherMsg{Path: ""})
				}
			}
		}
	}()

	if _, err := program.Run(); err != nil {
		return fmt.Errorf("run ui: %w", err)
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
