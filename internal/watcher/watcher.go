package watcher

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/MoeidHeidari/guardrail/internal/ignore"
	"github.com/fsnotify/fsnotify"
)

type Entry struct {
	Path    string
	Mode    fs.FileMode
	Size    int64
	ModTime time.Time
}

func (e Entry) IsDir() bool { return e.Mode.IsDir() }

type Change struct {
	Path string
	Kind string
	Time time.Time
}

type Update struct {
	// Entries is a complete snapshot when non-nil; otherwise apply Upsert/Removed.
	Entries []Entry
	Changes []Change
	Upsert  []Entry
	Removed []string
}

type Watcher struct {
	ignore   *ignore.Matcher
	Root     string
	Initial  Update
	backend  *fsnotify.Watcher
	index    map[string]Entry
	children map[string]map[string]bool
}

func ResolveRoot(path string) (string, error) {
	if path == "" {
		path = "."
	}
	root, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve context: %w", err)
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("resolve context %q: %w", path, err)
	}
	info, err := os.Stat(root)
	if err != nil {
		return "", fmt.Errorf("stat context: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("context %q is not a directory", path)
	}
	return root, nil
}

func New(path string) (*Watcher, error) {
	root, err := ResolveRoot(path)
	if err != nil {
		return nil, err
	}
	backend, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create filesystem watcher: %w", err)
	}
	w := &Watcher{Root: root, backend: backend}
	w.ignore, err = ignore.Load(root)
	if err != nil {
		return nil, errors.Join(err, backend.Close())
	}
	entries, err := w.scan()
	if err != nil {
		return nil, errors.Join(err, backend.Close())
	}
	w.Initial.Entries = entries
	w.index = make(map[string]Entry, len(entries))
	w.children = make(map[string]map[string]bool)
	for _, entry := range entries {
		w.put(entry)
	}
	return w, nil
}

func (w *Watcher) Close() error { return w.backend.Close() }

// Run coalesces bursts while bounding update latency, even during continuous writes.
// The caller owns Close and must wait for Run to return before closing the watcher.
func (w *Watcher) Run(ctx context.Context, updates chan<- Update) error {
	pending := make(map[string]fsnotify.Op)
	timer := time.NewTimer(time.Hour)
	timer.Stop()
	defer timer.Stop()
	var tick <-chan time.Time
	for {
		select {
		case <-ctx.Done():
			return nil
		case err, ok := <-w.backend.Errors:
			if !ok {
				return errors.New("filesystem watcher closed unexpectedly")
			}
			return fmt.Errorf("filesystem watcher: %w", err)
		case event, ok := <-w.backend.Events:
			if !ok {
				return errors.New("filesystem watcher closed unexpectedly")
			}
			if event.Name == w.Root && event.Has(fsnotify.Remove|fsnotify.Rename) {
				return fmt.Errorf("context %q was removed or renamed", w.Root)
			}
			relative, err := filepath.Rel(w.Root, event.Name)
			if err != nil {
				return fmt.Errorf("resolve changed path: %w", err)
			}
			if w.ignore.Matches(relative, false) {
				continue
			}
			if entry, exists := w.index[relative]; exists && entry.IsDir() && event.Op == fsnotify.Write {
				continue
			}
			pending[relative] |= event.Op
			if tick == nil {
				timer.Reset(20 * time.Millisecond)
				tick = timer.C
			}
		case <-tick:
			tick = nil
			update, err := w.reconcile(pending)
			if err != nil {
				return err
			}
			clear(pending)
			if len(update.Changes) > 0 || update.Entries != nil {
				select {
				case updates <- update:
				case <-ctx.Done():
					return nil
				}
			}
		}
	}
}

func (w *Watcher) scan() ([]Entry, error) {
	watched := make(map[string]bool)
	for _, path := range w.backend.WatchList() {
		watched[path] = true
	}
	directories := make(map[string]bool)
	var entries []Entry
	err := filepath.WalkDir(w.Root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if path != w.Root && errors.Is(walkErr, fs.ErrNotExist) {
				return nil
			}
			return walkErr
		}
		info, err := d.Info()
		if err != nil {
			if path != w.Root && errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		relative, err := filepath.Rel(w.Root, path)
		if err != nil {
			return err
		}
		if w.ignore.Matches(relative, info.IsDir()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			if !watched[path] {
				if err := w.backend.Add(path); err != nil {
					if path != w.Root && errors.Is(err, fs.ErrNotExist) {
						return filepath.SkipDir
					}
					return fmt.Errorf("watch directory %q: %w", path, err)
				}
			}
			directories[path] = true
		}
		entries = append(entries, Entry{
			Path: relative, Mode: info.Mode(), Size: info.Size(), ModTime: info.ModTime(),
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan context %q: %w", w.Root, err)
	}
	for path := range watched {
		if !directories[path] {
			if err := w.backend.Remove(path); err != nil && !removedWatch(err) {
				return nil, fmt.Errorf("unwatch directory %q: %w", path, err)
			}
		}
	}
	return entries, nil
}

func diff(before, after []Entry, writes map[string]bool) []Change {
	old := make(map[string]Entry, len(before))
	for _, entry := range before {
		old[entry.Path] = entry
	}
	now := time.Now()
	var changes []Change
	for _, entry := range after {
		previous, exists := old[entry.Path]
		kind := ""
		switch {
		case !exists:
			kind = "created"
		case previous.Mode != entry.Mode:
			kind = "changed"
		case !entry.IsDir() && (previous.Size != entry.Size || !previous.ModTime.Equal(entry.ModTime) || writes[entry.Path]):
			kind = "modified"
		}
		if kind != "" {
			changes = append(changes, Change{Path: entry.Path, Kind: kind, Time: now})
		}
		delete(old, entry.Path)
	}
	for path := range old {
		changes = append(changes, Change{Path: path, Kind: "deleted", Time: now})
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].Path < changes[j].Path })
	return changes
}
