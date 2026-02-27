package store

import (
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type WatchEvent struct {
	Path string
	Op   fsnotify.Op
	Err  error
}

func StartWatcher(dir string) (<-chan WatchEvent, func() error, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, nil, err
	}
	if err := watcher.Add(dir); err != nil {
		return nil, nil, err
	}
	ch := make(chan WatchEvent, 16)
	go func() {
		defer close(ch)
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if !isNoteFile(event.Name) {
					continue
				}
				ch <- WatchEvent{Path: event.Name, Op: event.Op}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				ch <- WatchEvent{Err: err}
			}
		}
	}()
	closeFn := func() error {
		return watcher.Close()
	}
	return ch, closeFn, nil
}

func isNoteFile(path string) bool {
	if filepath.Ext(path) != ".md" {
		return false
	}
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".clio_tmp_") {
		return false
	}
	return true
}
