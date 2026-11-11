package screens

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/grindlemire/go-tui"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	previewLimit     = 1 << 20
	previewLineLimit = 20000
)

type filePreview struct {
	path     string
	language string
	lines    [][]tui.TextSpan
	message  string
	width    int
	gutter   int
	code     string
	cursor   *changeCursor
}

type changeCursor struct {
	line   int
	column int
}

func locateChange(before, after string) *changeCursor {
	if before == after {
		return nil
	}
	old, current := []rune(before), []rune(after)
	start := 0
	for start < len(old) && start < len(current) && old[start] == current[start] {
		start++
	}
	oldEnd, end := len(old), len(current)
	for oldEnd > start && end > start && old[oldEnd-1] == current[end-1] {
		oldEnd--
		end--
	}
	position := start
	if end > start {
		position = end - 1
		// A line-ending write belongs to the line just written, not the next blank row.
		if current[position] != '\n' {
			position++
		}
	}
	line, column := 0, 0
	for _, r := range current[:position] {
		if r == '\n' {
			line, column = line+1, 0
		} else {
			column += len([]rune(safeCode(string(r))))
		}
	}
	return &changeCursor{line: line, column: column}
}

func readPreview(rootPath, path string) (content []byte, err error) {
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, root.Close()) }()
	info, err := root.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("only regular files can be previewed (got %s)", info.Mode().Type())
	}
	if info.Size() > previewLimit {
		return nil, fmt.Errorf("file exceeds the 1 MiB preview limit")
	}
	file, err := root.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	info, err = file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("only regular files can be previewed (got %s)", info.Mode().Type())
	}
	content, err = io.ReadAll(io.LimitReader(file, previewLimit+1))
	if err != nil {
		return nil, err
	}
	if len(content) > previewLimit {
		return nil, fmt.Errorf("file exceeds the 1 MiB preview limit")
	}
	if bytes.IndexByte(content, 0) >= 0 || !utf8.Valid(content) {
		return nil, fmt.Errorf("binary or non-UTF-8 file; no text preview available")
	}
	if bytes.Count(content, []byte{'\n'})+1 > previewLineLimit {
		return nil, fmt.Errorf("file exceeds the %d line preview limit", previewLineLimit)
	}
	return content, nil
}

func safeCode(text string) string {
	var out strings.Builder
	for _, r := range text {
		switch {
		case r == '\t':
			out.WriteString("    ")
		case unicode.IsControl(r):
			quoted := strconv.QuoteRune(r)
			out.WriteString(quoted[1 : len(quoted)-1])
		default:
			out.WriteRune(r)
		}
	}
	return out.String()
}

func highlight(path, code string) (lines [][]tui.TextSpan, language string, err error) {
	return highlightWithStyle(path, code, "onedark")
}

func highlightWithStyle(path, code, syntax string) (lines [][]tui.TextSpan, language string, err error) {
	// Chroma iterators can panic on lexer errors; show these as preview errors.
	defer func() {
		if failure := recover(); failure != nil {
			lines = nil
			err = fmt.Errorf("syntax highlighting failed: %v", failure)
		}
	}()
	lexer := lexers.Match(path)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	language = lexer.Config().Name
	iterator, err := chroma.Coalesce(lexer).Tokenise(&chroma.TokeniseOptions{State: "root"}, code)
	if err != nil {
		return nil, language, fmt.Errorf("syntax highlighting failed: %w", err)
	}
	if syntax == "" {
		syntax = "onedark"
	}
	palette := styles.Get(syntax)
	lines = make([][]tui.TextSpan, 1)
	for token := iterator(); token != chroma.EOF; token = iterator() {
		entry := palette.Get(token.Type)
		style := tui.NewStyle()
		if entry.Colour.IsSet() {
			style = style.Foreground(tui.RGBColor(entry.Colour.Red(), entry.Colour.Green(), entry.Colour.Blue()))
		}
		if entry.Bold == chroma.Yes {
			style = style.Bold()
		}
		if entry.Italic == chroma.Yes {
			style = style.Italic()
		}
		parts := strings.Split(token.Value, "\n")
		for i, part := range parts {
			if i > 0 {
				lines = append(lines, nil)
			}
			if part != "" {
				index := len(lines) - 1
				lines[index] = append(lines[index], tui.TextSpan{Text: safeCode(part), Style: style})
			}
		}
	}
	return lines, language, nil
}
