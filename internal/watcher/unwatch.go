package watcher

import (
	"errors"
	"runtime"
	"syscall"

	"github.com/fsnotify/fsnotify"
)

func removedWatch(err error) bool {
	// Linux invalidates inotify watches on deletion before fsnotify consumes IN_IGNORED.
	return errors.Is(err, fsnotify.ErrNonExistentWatch) ||
		(runtime.GOOS == "linux" && errors.Is(err, syscall.EINVAL))
}
