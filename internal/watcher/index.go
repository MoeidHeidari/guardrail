package watcher

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/fsnotify/fsnotify"
)

func (w *Watcher) put(entry Entry) {
	w.index[entry.Path] = entry
	if entry.Path != "." {
		parent := filepath.Dir(entry.Path)
		if w.children[parent] == nil {
			w.children[parent] = make(map[string]bool)
		}
		w.children[parent][entry.Path] = true
	}
}

func (w *Watcher) reconcile(pending map[string]fsnotify.Op) (Update, error) {
	touched := make(map[string]Entry)
	absent := make(map[string]bool)
	remember := func(path string) {
		if _, seen := touched[path]; seen || absent[path] {
			return
		}
		if old, exists := w.index[path]; exists {
			touched[path] = old
		} else {
			absent[path] = true
		}
	}
	var remove func(string) error
	remove = func(path string) error {
		remember(path)
		for child := range w.children[path] {
			if err := remove(child); err != nil {
				return err
			}
		}
		if old, exists := w.index[path]; exists && old.IsDir() {
			err := w.backend.Remove(filepath.Join(w.Root, path))
			if err != nil && !errors.Is(err, fsnotify.ErrNonExistentWatch) {
				return fmt.Errorf("unwatch %q: %w", path, err)
			}
		}
		delete(w.index, path)
		delete(w.children, path)
		delete(w.children[filepath.Dir(path)], path)
		return nil
	}
	paths := make([]string, 0, len(pending))
	for path := range pending {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		absolute := filepath.Join(w.Root, path)
		info, err := os.Lstat(absolute)
		if errors.Is(err, fs.ErrNotExist) {
			if path == "." {
				return Update{}, fmt.Errorf("context %q no longer exists", w.Root)
			}
			if err := remove(path); err != nil {
				return Update{}, err
			}
			continue
		}
		if err != nil {
			return Update{}, fmt.Errorf("refresh %q: %w", path, err)
		}
		old, exists := w.index[path]
		if exists && (old.IsDir() != info.IsDir() || pending[path].Has(fsnotify.Rename|fsnotify.Remove)) {
			if err := remove(path); err != nil {
				return Update{}, err
			}
			exists = false
		}
		if info.IsDir() && (!exists || pending[path].Has(fsnotify.Create)) {
			err = filepath.WalkDir(absolute, func(full string, d fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					if errors.Is(walkErr, fs.ErrNotExist) {
						return nil
					}
					return walkErr
				}
				meta, err := d.Info()
				if errors.Is(err, fs.ErrNotExist) {
					return nil
				}
				if err != nil {
					return err
				}
				relative, err := filepath.Rel(w.Root, full)
				if err != nil {
					return err
				}
				if meta.IsDir() {
					if err := w.backend.Add(full); err != nil {
						if errors.Is(err, fs.ErrNotExist) {
							return filepath.SkipDir
						}
						return err
					}
				}
				remember(relative)
				w.put(Entry{Path: relative, Mode: meta.Mode(), Size: meta.Size(), ModTime: meta.ModTime()})
				return nil
			})
			if err != nil {
				return Update{}, fmt.Errorf("discover subtree %q: %w", path, err)
			}
		} else {
			remember(path)
			w.put(Entry{Path: path, Mode: info.Mode(), Size: info.Size(), ModTime: info.ModTime()})
		}
	}
	var before, after []Entry
	writes := make(map[string]bool)
	for path, op := range pending {
		writes[path] = op.Has(fsnotify.Write)
	}
	for path, old := range touched {
		before = append(before, old)
		if entry, exists := w.index[path]; exists {
			after = append(after, entry)
		}
	}
	for path := range absent {
		if entry, exists := w.index[path]; exists {
			after = append(after, entry)
		}
	}
	update := Update{Changes: diff(before, after, writes)}
	for _, change := range update.Changes {
		if change.Kind == "deleted" {
			update.Removed = append(update.Removed, change.Path)
		} else {
			update.Upsert = append(update.Upsert, w.index[change.Path])
		}
	}
	return update, nil
}
