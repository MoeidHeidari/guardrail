package ignore

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

const FileName = ".guardignore"

type Matcher struct{ matcher gitignore.Matcher }

func Load(root string) (matcher *Matcher, err error) {
	dir, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, dir.Close()) }()
	file, err := dir.Open(FileName)
	if errors.Is(err, os.ErrNotExist) {
		return &Matcher{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", FileName, err)
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	data, err := io.ReadAll(io.LimitReader(file, (1<<20)+1))
	if err != nil {
		return nil, err
	}
	if len(data) > 1<<20 {
		return nil, fmt.Errorf("%s exceeds 1 MiB", FileName)
	}
	var patterns []gitignore.Pattern
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	scanner.Buffer(make([]byte, 4096), 1<<20)
	for scanner.Scan() {
		line := strings.TrimSuffix(scanner.Text(), "\r")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, gitignore.ParsePattern(line, nil))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return &Matcher{matcher: gitignore.NewMatcher(patterns)}, nil
}

func (matcher *Matcher) Matches(path string, directory bool) bool {
	path = filepath.ToSlash(path)
	if matcher == nil || matcher.matcher == nil || path == "." || path == FileName {
		return false
	}
	parts := strings.Split(path, "/")
	// Git does not re-include files whose parent directory is excluded.
	for i := 1; i < len(parts); i++ {
		if matcher.matcher.Match(parts[:i], true) {
			return true
		}
	}
	return matcher.matcher.Match(parts, directory)
}
