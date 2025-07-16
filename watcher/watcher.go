package watcher

import (
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/fsnotify/fsnotify"
)

const forever = time.Duration(math.MaxInt64)

type Config struct {
	RootPath         string
	Patterns         []string
	IgnorePatterns   []string
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

	for _, pattern := range append(cfg.Patterns, cfg.IgnorePatterns...) {
		if !doublestar.ValidatePattern(pattern) {
			return nil, fmt.Errorf("invalid pattern: %v", pattern)
		}
	}

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
			logrus.Errorf("Watcher.Close(): %s", closeErr.Error())
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
			logrus.Debugf("failed to list: %s", err)
			return nil
		}

		if !d.IsDir() {
			return nil
		}

		match, err := w.match(w.Config.Patterns, path)
		if err != nil {
			panic(fmt.Sprintf("bug detected: pattern match failed: %s: %s", path, err))
		}
		if !match {
			logrus.Debugf("not a match: %s", path)
			return nil
		}

		match, err = w.match(w.Config.IgnorePatterns, path)
		if err != nil {
			panic(fmt.Sprintf("bug detected: pattern match failed: %s: %s", path, err))
		}
		if match {
			logrus.Debugf("ignoring: %s", path)
			return nil
		}

		if err := w.fsnWatcher.Add(path); err != nil {
			if path == w.Config.RootPath {
				return fmt.Errorf("failed to watch root %s: %s", path, err)
			}
			logrus.Errorf("failed to watch %s: %s", path, err)
		}
		logrus.Debugf("watching: %s", path)

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

func (w *Watcher) match(patterns []string, path string) (bool, error) {
	for _, pattern := range patterns {
		pathSeparator := string(os.PathSeparator)
		absPattern := w.Config.RootPath + pathSeparator + strings.TrimPrefix(pattern, pathSeparator)
		match, err := doublestar.PathMatch(
			absPattern,
			path,
		)
		if err != nil {
			return false, err
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}

func (w *Watcher) processEvent(
	event fsnotify.Event,
	lastChangedTime *time.Time,
	debounceWait *time.Duration,
) {
	logrus.Tracef("event: %v", event)

	addRecursively, err := isNewDirectory(event)
	if err != nil {
		w.ErrorsCn <- err
	} else if addRecursively {
		if err := w.addRecursively(event.Name); err != nil {
			w.ErrorsCn <- err
		}
	}

	logrus.Infof("Changed: %s (%s)", event.Name, event.Op)
	now := time.Now()
	if lastChangedTime.IsZero() {
		logrus.Trace("first event received, waiting to debounce")
		*debounceWait = w.Config.DebounceDuration
	} else {
		if now.Sub(*lastChangedTime) > w.Config.DebounceDuration {
			logrus.Trace("debounced, sending change event")
			w.ChangedFilesCn <- struct{}{}
			*debounceWait = forever
		} else {
			logrus.Trace("change was too fast, debouncing")
			*debounceWait = w.Config.DebounceDuration
		}
	}
	*lastChangedTime = now
}

func (w *Watcher) watch() {
	debounceWait := time.Duration(0)
	var lastChangedTime time.Time
	for {
		select {
		// TODO cancellation contexts
		case event, ok := <-w.fsnWatcher.Events:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}
			w.processEvent(event, &lastChangedTime, &debounceWait)
		case <-time.After(debounceWait):
			w.ChangedFilesCn <- struct{}{}
			debounceWait = forever
			lastChangedTime = time.Time{}
		case err, ok := <-w.fsnWatcher.Errors:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}
			w.ErrorsCn <- err
		}
	}
}

func (w *Watcher) Close() error {
	// TODO cancellation context for watch
	return w.fsnWatcher.Close()
}
