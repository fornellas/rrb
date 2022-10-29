package watcher

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Config struct {
	RootPath         string
	Pattern          string
	DebounceDuration time.Duration
}

type Watcher struct {
	Config         Config
	ChangedFilesCn chan struct{}
	ErrorsCn       chan error
	fsnWatcher     *fsnotify.Watcher
}

func NewWatcher(cfg Config) (*Watcher, error) {
	var err error

	if !filepath.IsAbs(cfg.RootPath) {
		cfg.RootPath, err = filepath.Abs(cfg.RootPath)
		if err != nil {
			return nil, fmt.Errorf("filepath.Abs: %s", err.Error())
		}
	}
	cfg.RootPath = filepath.Clean(cfg.RootPath)

	w := &Watcher{
		Config:         cfg,
		ChangedFilesCn: make(chan struct{}),
		ErrorsCn:       make(chan error),
	}
	w.fsnWatcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("NewWatcher: %s", err.Error())
	}

	if err = w.addRecursively(w.Config.RootPath); err != nil {
		if closeErr := w.fsnWatcher.Close(); closeErr != nil {
			log.Printf("Watcher.Close(): %s", closeErr.Error())
		}
		return nil, err
	}

	// TODO cancellation context
	go w.watch()

	return w, nil
}

func (w *Watcher) addRecursively(path string) error {
	if err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if path == w.Config.RootPath {
				return fmt.Errorf("failed to list root directory: %s", err)
			}
			log.Printf("failed to list: %s", err)
			return nil
		}

		if !d.IsDir() {
			return nil
		}

		if err := w.fsnWatcher.Add(path); err != nil {
			if path == w.Config.RootPath {
				return fmt.Errorf("failed to watch root %s: %s", path, err)
			}
			log.Printf("failed to watch %s: %s", path, err)
		}
		log.Printf("watching: %s", path)

		return nil
	}); err != nil {
		return err
	}
	return nil
}

func isNewDirectory(event fsnotify.Event) (bool, error) {
	if !(event.Has(fsnotify.Create) || event.Has(fsnotify.Chmod)) {
		return false, nil
	}
	fileInfo, err := os.Stat(event.Name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return fileInfo.IsDir(), nil
}

func (w *Watcher) watch() {
	forever := time.Duration(math.MaxInt64)
	debounceWait := forever
	var lastChangedTime time.Time
	for {
		select {
		// TODO cancellation contexts
		case event, ok := <-w.fsnWatcher.Events:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}
			log.Println("event:", event)

			addRecursively, err := isNewDirectory(event)
			if err != nil {
				log.Println("error:", err)
				w.ErrorsCn <- err
			} else if addRecursively {
				if err := w.addRecursively(event.Name); err != nil {
					log.Println("error:", err)
					w.ErrorsCn <- err
				}
			}

			log.Println("CHANGED")
			now := time.Now()
			if lastChangedTime.IsZero() {
				log.Println("first event, debouncing")
				debounceWait = w.Config.DebounceDuration
			} else {
				if now.Sub(lastChangedTime) > w.Config.DebounceDuration {
					log.Println("delay elapsed")
					log.Println(">>>> BUILD <<<<")
					debounceWait = forever
				} else {
					log.Println("change too short, waiting for delay")
					debounceWait = w.Config.DebounceDuration
				}
			}
			lastChangedTime = now
		case <-time.After(debounceWait):
			log.Println("debounced")
			log.Println(">>>> BUILD <<<<")
			debounceWait = forever
			lastChangedTime = time.Time{}
		case err, ok := <-w.fsnWatcher.Errors:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}
			log.Println("error:", err)
			w.ErrorsCn <- err
		}
	}
}

func (w *Watcher) Close() error {
	// TODO cancellation context for watch
	return w.fsnWatcher.Close()
}
