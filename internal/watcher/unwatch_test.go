package watcher

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"testing"

	"github.com/fsnotify/fsnotify"
)

func TestRemovedWatchErrors(t *testing.T) {
	if removedWatch(fmt.Errorf("watch: %w", os.ErrNotExist)) != (runtime.GOOS == "windows") {
		t.Fatal("Windows removal of deleted watch paths must tolerate missing files")
	}
	if !removedWatch(fmt.Errorf("watch: %w", fsnotify.ErrNonExistentWatch)) {
		t.Fatal("already removed watches must be tolerated")
	}
	if removedWatch(fmt.Errorf("watch: %w", syscall.EINVAL)) != (runtime.GOOS == "linux") {
		t.Fatal("only Linux invalidated watch descriptors should tolerate EINVAL")
	}
	for _, err := range []error{syscall.EBADF, syscall.EACCES, fmt.Errorf("backend failure")} {
		if removedWatch(err) {
			t.Fatalf("unexpected backend errors must remain visible: %v", err)
		}
	}
}
