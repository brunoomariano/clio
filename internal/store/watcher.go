package store

import (
	"os"
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
	return startWatcherWithFilter([]string{dir}, isLegacyWatchFile)
}

func StartWatcherDirs(dirs []string) (<-chan WatchEvent, func() error, error) {
	return startWatcherWithFilter(dirs, isWatchedFile)
}

func startWatcherWithFilter(dirs []string, allowed func(path string) bool) (<-chan WatchEvent, func() error, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, nil, err
	}
	for _, dir := range dirs {
		if err := addWatcherDirs(watcher, dir); err != nil {
			_ = watcher.Close()
			return nil, nil, err
		}
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
				if event.Op&fsnotify.Create == fsnotify.Create {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						_ = addWatcherDirs(watcher, event.Name)
						continue
					}
				}
				if !allowed(event.Name) {
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

func isLegacyWatchFile(path string) bool {
	if filepath.Ext(path) != ".md" {
		return false
	}
	return isWatchedFile(path)
}

func isNoteFile(path string) bool {
	return isLegacyWatchFile(path)
}

func isWatchedFile(path string) bool {
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".clio_tmp_") {
		return false
	}
	return true
}

func addWatcherDirs(w *fsnotify.Watcher, root string) error {
	if _, err := os.Stat(root); err != nil {
		return err
	}
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		return w.Add(path)
	})
}
