package watcher

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/fsnotify/fsnotify"
)

func TestWriteBatchDeliversOnlyChangedEntries(t *testing.T) {
	root := t.TempDir()
	for i := range 100 {
		write(t, filepath.Join(root, fmt.Sprintf("file%d", i)), "before")
	}
	w, updates, _ := start(t, root)
	write(t, filepath.Join(root, "file1"), "after")
	update := await(t, updates, map[string]string{"file1": "modified"})
	if update.Entries != nil || len(update.Upsert) != 1 || len(update.Removed) != 0 {
		t.Fatalf("single write must deliver a delta, not a full snapshot: %+v", update)
	}
	if len(w.Initial.Entries) != 101 {
		t.Fatal("initial snapshot must remain complete")
	}
}

func TestSubtreeReconciliationDoesNotInspectUnrelatedPaths(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "untouched"), "before")
	w, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	original := w.index["untouched"]
	// Deliberately omit the unrelated write from this batch. If reconciliation
	// scans the whole context it would incorrectly include it in this delta.
	write(t, filepath.Join(root, "untouched"), "unrelated change")
	if err := os.Mkdir(filepath.Join(root, "new"), 0o700); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(root, "new", "file"), "new child")
	update, err := w.reconcile(map[string]fsnotify.Op{"new": fsnotify.Create})
	if err != nil {
		t.Fatal(err)
	}
	if len(update.Upsert) != 2 || update.Entries != nil || w.index["untouched"] != original {
		t.Fatalf("subtree creation must not inspect unrelated files: %+v", update)
	}
	if err := os.RemoveAll(filepath.Join(root, "new")); err != nil {
		t.Fatal(err)
	}
	update, err = w.reconcile(map[string]fsnotify.Op{"new": fsnotify.Remove})
	if err != nil || len(update.Removed) != 2 {
		t.Fatalf("directory removal must delete its indexed subtree: %+v, %v", update, err)
	}
	if _, exists := w.index[filepath.Join("new", "file")]; exists {
		t.Fatal("deleted child retained in index")
	}
}

func BenchmarkSingleWriteBatch(b *testing.B) {
	for _, count := range []int{1000, 100000} {
		b.Run(fmt.Sprint(count), func(b *testing.B) {
			root := b.TempDir()
			if err := os.WriteFile(filepath.Join(root, "file0"), []byte("current"), 0o600); err != nil {
				b.Fatal(err)
			}
			entries := make([]Entry, count)
			w := &Watcher{Root: root, index: make(map[string]Entry), children: make(map[string]map[string]bool)}
			for i := range count {
				entry := Entry{Path: fmt.Sprintf("file%d", i)}
				entries[i] = entry
				w.put(entry)
			}
			pending := map[string]fsnotify.Op{"file0": fsnotify.Write}
			b.Run("indexed", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					if _, err := w.reconcile(pending); err != nil {
						b.Fatal(err)
					}
				}
			})
			b.Run("previous-full-copy-diff", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					snapshot := append([]Entry(nil), entries...)
					info, err := os.Lstat(filepath.Join(root, "file0"))
					if err != nil {
						b.Fatal(err)
					}
					snapshot[0] = Entry{Path: "file0", Mode: info.Mode(), Size: info.Size(), ModTime: info.ModTime()}
					diff(entries, snapshot, map[string]bool{"file0": true})
				}
			})
		})
	}
}
