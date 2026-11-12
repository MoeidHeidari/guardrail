package watcher

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func write(t *testing.T, path, text string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestResolveRoot(t *testing.T) {
	root := t.TempDir()
	want, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{root, filepath.Join(root, ".")} {
		got, err := ResolveRoot(path)
		if err != nil || got != want {
			t.Fatalf("ResolveRoot(%q) = %q, %v; want %q", path, got, err, want)
		}
	}
	t.Chdir(root)
	if got, err := ResolveRoot(""); err != nil || got != want {
		t.Fatalf("default context = %q, %v", got, err)
	}
	file := filepath.Join(root, "file")
	write(t, file, "")
	for _, path := range []string{file, filepath.Join(root, "missing")} {
		if _, err := ResolveRoot(path); err == nil {
			t.Fatalf("accepted invalid context %q", path)
		}
	}
	link := filepath.Join(t.TempDir(), "context")
	if err := os.Symlink(root, link); err != nil {
		t.Fatal(err)
	}
	if got, err := ResolveRoot(link); err != nil || got != want {
		t.Fatalf("symlink context = %q, %v", got, err)
	}
}

func start(t *testing.T, root string) (*Watcher, <-chan Update, <-chan error) {
	t.Helper()
	w, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	updates := make(chan Update, 100)
	done := make(chan error, 1)
	go func() { done <- w.Run(ctx, updates) }()
	t.Cleanup(func() {
		cancel()
		select {
		case err := <-done:
			if err != nil {
				t.Error(err)
			}
		case <-time.After(5 * time.Second):
			t.Error("watcher did not stop")
		}
		if err := w.Close(); err != nil {
			t.Error(err)
		}
	})
	return w, updates, done
}

func await(t *testing.T, updates <-chan Update, expected map[string]string) Update {
	t.Helper()
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	var last Update
	for len(expected) > 0 {
		select {
		case last = <-updates:
			for _, change := range last.Changes {
				if expected[change.Path] == change.Kind {
					delete(expected, change.Path)
				}
			}
		case <-timer.C:
			t.Fatalf("timed out waiting for changes: %v; last update: %+v", expected, last)
		}
	}
	return last
}

func TestRecursiveChanges(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "existing"), 0o700); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(root, "existing", "file.txt")
	write(t, file, "before")
	write(t, filepath.Join(root, ".hidden"), "hidden")
	w, updates, _ := start(t, root)
	if len(w.Initial.Entries) != 4 {
		t.Fatalf("initial discovery missed files: %+v", w.Initial)
	}
	write(t, file, "after")
	await(t, updates, map[string]string{filepath.Join("existing", "file.txt"): "modified"})
	nested := filepath.Join(root, "new", "nested")
	if err := os.MkdirAll(nested, 0o700); err != nil {
		t.Fatal(err)
	}
	child := filepath.Join(nested, "child.txt")
	write(t, child, "created before directory watch")
	await(t, updates, map[string]string{
		"new": "created", filepath.Join("new", "nested"): "created",
		filepath.Join("new", "nested", "child.txt"): "created",
	})
	write(t, child, "updated")
	await(t, updates, map[string]string{filepath.Join("new", "nested", "child.txt"): "modified"})
	if err := os.Rename(filepath.Join(root, "new"), filepath.Join(root, "renamed")); err != nil {
		t.Fatal(err)
	}
	await(t, updates, map[string]string{
		"new": "deleted", "renamed": "created",
		filepath.Join("new", "nested", "child.txt"):     "deleted",
		filepath.Join("renamed", "nested", "child.txt"): "created",
	})
	child = filepath.Join(root, "renamed", "nested", "child.txt")
	write(t, child, "after rename")
	await(t, updates, map[string]string{filepath.Join("renamed", "nested", "child.txt"): "modified"})
	if err := os.RemoveAll(filepath.Join(root, "renamed")); err != nil {
		t.Fatal(err)
	}
	await(t, updates, map[string]string{
		"renamed": "deleted", filepath.Join("renamed", "nested"): "deleted",
		filepath.Join("renamed", "nested", "child.txt"): "deleted",
	})
	temporary := filepath.Join(root, "replacement")
	write(t, temporary, "atomic save")
	if err := os.Rename(temporary, file); err != nil {
		t.Fatal(err)
	}
	await(t, updates, map[string]string{filepath.Join("existing", "file.txt"): "modified"})
	write(t, file, "after atomic save")
	await(t, updates, map[string]string{filepath.Join("existing", "file.txt"): "modified"})
}

func TestSymlinksAreNotTraversed(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	write(t, filepath.Join(outside, "outside.txt"), "outside")
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(root, filepath.Join(root, "loop")); err != nil {
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
	if len(w.Initial.Entries) != 3 {
		t.Fatalf("symlinks must be listed but not followed: %+v", w.Initial)
	}
}

func TestDirectoryRecreationAndPermissions(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "dir")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	_, updates, _ := start(t, root)
	if err := os.Remove(dir); err != nil {
		t.Fatal(err)
	}
	await(t, updates, map[string]string{"dir": "deleted"})
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	await(t, updates, map[string]string{"dir": "created"})
	child := filepath.Join(dir, "file")
	write(t, child, "one")
	await(t, updates, map[string]string{filepath.Join("dir", "file"): "created"})
	if err := os.Chmod(child, 0o400); err != nil {
		t.Fatal(err)
	}
	await(t, updates, map[string]string{filepath.Join("dir", "file"): "changed"})
}

func TestDiffDetectsWritesWithoutMetadataChanges(t *testing.T) {
	entry := Entry{Path: "file", Size: 3, ModTime: time.Now()}
	changes := diff([]Entry{entry}, []Entry{entry}, map[string]bool{"file": true})
	if len(changes) != 1 || changes[0].Kind != "modified" {
		t.Fatalf("write must be visible even when size and timestamp match: %+v", changes)
	}
	entry.Mode = os.ModeDir
	if changes := diff([]Entry{entry}, []Entry{entry}, map[string]bool{"file": true}); len(changes) != 0 {
		t.Fatalf("directory writes should not duplicate child events: %+v", changes)
	}
}

func TestCancellationDuringPendingUpdate(t *testing.T) {
	root := t.TempDir()
	w, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := w.Close(); err != nil {
			t.Error(err)
		}
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- w.Run(ctx, make(chan Update)) }()
	write(t, filepath.Join(root, "file"), "pending")
	<-time.After(150 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("cancellation must not block on delivering updates")
	}
}

func TestRootRemovalIsReported(t *testing.T) {
	root := filepath.Join(t.TempDir(), "context")
	if err := os.Mkdir(root, 0o700); err != nil {
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- w.Run(ctx, make(chan Update)) }()
	if err := os.Remove(root); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err == nil {
		t.Fatal("removing context must report error")
	}
}
