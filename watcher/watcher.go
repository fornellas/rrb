package watcher

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type WatcherConfig struct {
	RootPath string
	Pattern  string
}

type Watcher struct {
	Config         WatcherConfig
	ChangedFilesCn chan string
	ErrorsCn       chan error
	fsnWatcher     *fsnotify.Watcher
}

func NewWatcher(cfg WatcherConfig) (*Watcher, error) {
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
		ChangedFilesCn: make(chan string),
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

func (w *Watcher) add(path string) error {
	if err := w.fsnWatcher.Add(path); err != nil {
		if path == w.Config.RootPath {
			return fmt.Errorf("failed to watch root %s: %s", path, err)
		}
		log.Printf("failed to watch %s: %s", path, err)
	}
	log.Printf("watching: %s", path)
	return nil
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
		if err := w.add(path); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (w *Watcher) watch() {
	for {
		select {
		// TODO cancellation contexts
		case event, ok := <-w.fsnWatcher.Events:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}
			log.Println("event:", event)
			if event.Has(fsnotify.Write) {
				log.Println("modified file:", event.Name)
			}
		case err, ok := <-w.fsnWatcher.Errors:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}
			w.ErrorsCn <- err
		}
	}
}

func (w *Watcher) Close() error {
	// TODO cancellation context
	return w.fsnWatcher.Close()
}
