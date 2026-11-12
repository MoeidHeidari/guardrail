package watcher

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestIgnoreDiscoveryAndLiveReload(t *testing.T) {
	root := t.TempDir()
	write := func(path, text string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, path)), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, path), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("cache/hidden.go", "package main")
	write("visible.go", "package main")
	write(".guardignore", "cache/\n*.log\n")
	w, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := w.Close(); err != nil {
			t.Error(err)
		}
	})
	if _, exists := w.index["cache"]; exists {
		t.Fatal("ignored subtree discovered")
	}
	for _, path := range w.backend.WatchList() {
		if path == filepath.Join(root, "cache") {
			t.Fatal("ignored directory allocated a watch")
		}
	}
	write("new/cache/a.go", "package main")
	update, err := w.reconcile(map[string]fsnotify.Op{"new": fsnotify.Create})
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range update.Upsert {
		if entry.Path == "new/cache" || entry.Path == "new/cache/a.go" {
			t.Fatal("new ignored subtree included")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	updates := make(chan Update, 20)
	go func() { done <- w.Run(ctx, updates) }()
	defer func() {
		cancel()
		if err := <-done; err != nil {
			t.Error(err)
		}
	}()
	write(".guardignore", "*.log\n")
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	for {
		select {
		case update := <-updates:
			if update.Entries != nil {
				found := false
				for _, entry := range update.Entries {
					found = found || entry.Path == "cache/hidden.go"
				}
				if !found {
					t.Fatal("unignored files not discovered")
				}
				return
			}
		case <-timer.C:
			t.Fatal("ignore edit was not reloaded")
		}
	}
}

func TestIgnoreRemovalRescans(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".guardignore"), []byte("a.go\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "a.go"), []byte("package main"), 0600); err != nil {
		t.Fatal(err)
	}
	w, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := w.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := os.Remove(filepath.Join(root, ".guardignore")); err != nil {
		t.Fatal(err)
	}
	update, err := w.reconcile(map[string]fsnotify.Op{".guardignore": fsnotify.Remove})
	if err != nil {
		t.Fatal(err)
	}
	if update.Entries == nil {
		t.Fatal("ignore deletion must publish full snapshot")
	}
	if _, found := w.index["a.go"]; !found {
		t.Fatal("previously excluded file not restored")
	}
}
