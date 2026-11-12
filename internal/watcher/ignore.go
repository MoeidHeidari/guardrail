package watcher

import (
	"github.com/MoeidHeidari/guardrail/internal/ignore"
	"github.com/fsnotify/fsnotify"
)

func (w *Watcher) reloadIgnore(pending map[string]fsnotify.Op) (Update, error) {
	matcher, err := ignore.Load(w.Root)
	if err != nil {
		return Update{}, err
	}
	w.ignore = matcher
	var before []Entry
	for _, entry := range w.index {
		before = append(before, entry)
	}
	entries, err := w.scan()
	if err != nil {
		return Update{}, err
	}
	clear(w.index)
	clear(w.children)
	for _, entry := range entries {
		w.put(entry)
	}
	writes := make(map[string]bool)
	for path, op := range pending {
		writes[path] = op.Has(fsnotify.Write)
	}
	return Update{Entries: entries, Changes: diff(before, entries, writes)}, nil
}
