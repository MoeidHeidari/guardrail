package watcher

import (
	"errors"
	"os"
	"runtime"
	"syscall"

	"github.com/fsnotify/fsnotify"
)

func removedWatch(err error) bool {
	return errors.Is(err, fsnotify.ErrNonExistentWatch) ||
		(runtime.GOOS == "windows" && errors.Is(err, os.ErrNotExist)) ||
		(runtime.GOOS == "linux" && errors.Is(err, syscall.EINVAL))
}
