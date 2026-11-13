package screens

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
	"github.com/MoeidHeidari/guardrail/internal/watcher"
	"github.com/grindlemire/go-tui"
)

func previewText(lines [][]tui.TextSpan) string {
	var out []string
	for _, line := range lines {
		var text strings.Builder
		for _, span := range line {
			text.WriteString(span.Text)
		}
		out = append(out, text.String())
	}
	return strings.Join(out, "\n")
}

func TestHighlightLanguages(t *testing.T) {
	for _, test := range []struct {
		path, code string
	}{
		{"main.go", "package main\n\nfunc main() { println(\"hello\") }\n"},
		{"app.ts", "const value: string = \"hello\";\n"},
		{"app.js", "const value = \"hello\";\n"},
		{"app.py", "def greet():\n    return \"hello\"\n"},
		{"config.json", "{\"enabled\": true, \"value\": 42}\n"},
		{"config.yaml", "enabled: true\nvalue: 42\n"},
		{"main.rs", "fn main() { println!(\"hello\"); }\n"},
		{"run.sh", "#!/bin/sh\necho \"hello\"\n"},
		{"Dockerfile", "FROM golang:1.26\nRUN echo hello\n"},
	} {
		t.Run(test.path, func(t *testing.T) {
			lines, language, err := highlight(test.path, test.code)
			if err != nil {
				t.Fatal(err)
			}
			if language == "" || language == "plaintext" {
				t.Fatalf("no language detected for %s", test.path)
			}
			if got := previewText(lines); got != test.code {
				t.Fatalf("highlighter changed code:\ngot %q\nwant %q", got, test.code)
			}
			colors := make(map[string]bool)
			for _, line := range lines {
				for _, span := range line {
					if !span.Style.Fg.IsDefault() {
						colors[fmt.Sprint(span.Style.Fg)] = true
					}
				}
			}
			if len(colors) < 2 {
				t.Fatalf("expected distinct syntax colors, got %d", len(colors))
			}
		})
	}
}

func TestHighlightPlainTextAndControls(t *testing.T) {
	for _, code := range []string{"", "hello", "hello\n", "a\n\nb", "你好\ncafé"} {
		lines, _, err := highlight("unknown.guardrail-unknown-extension", code)
		if err != nil || previewText(lines) != code {
			t.Fatalf("plain text %q became %q: %v", code, previewText(lines), err)
		}
	}
	if got := safeCode("\t\x1b[2J\x07\r"); got != `    \x1b[2J\a\r` {
		t.Fatalf("unexpected control rendering: %q", got)
	}
}

func TestReadPreviewLimitsAndInvalidFiles(t *testing.T) {
	root := t.TempDir()
	for _, test := range []struct {
		name string
		data []byte
		want string
	}{
		{"text", []byte("actual content"), ""},
		{"empty", nil, ""},
		{"exact-limit", []byte(strings.Repeat("x", previewLimit)), ""},
		{"oversize", []byte(strings.Repeat("x", previewLimit+1)), "1 MiB"},
		{"binary", []byte{'a', 0, 'b'}, "binary"},
		{"invalid-utf8", []byte{0xff, 0xfe}, "non-UTF-8"},
		{"too-many-lines", []byte(strings.Repeat("\n", previewLineLimit)), "line preview limit"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := os.WriteFile(filepath.Join(root, test.name), test.data, 0o600); err != nil {
				t.Fatal(err)
			}
			got, err := readPreview(root, test.name)
			if test.want != "" {
				if err == nil || !strings.Contains(err.Error(), test.want) {
					t.Fatalf("expected %q error, got %v", test.want, err)
				}
			} else if err != nil || string(got) != string(test.data) {
				t.Fatalf("unexpected preview: %v", err)
			}
		})
	}
	if _, err := readPreview(root, "missing"); err == nil {
		t.Fatal("missing files must report an error")
	}
	if _, err := readPreview(root, "."); err == nil {
		t.Fatal("directories must not be read as files")
	}
	outside := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(outside, []byte("outside"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	if _, err := readPreview(root, "link"); err == nil {
		t.Fatal("symlinks must not be previewed")
	}
	if _, err := readPreview(root, filepath.Join("..", filepath.Base(outside))); err == nil {
		t.Fatal("preview must not escape the context")
	}
}

func fileScreen(t *testing.T, code string) *Main {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte(code), 0o600); err != nil {
		t.Fatal(err)
	}
	w, err := watcher.New(root)
	if err != nil {
		t.Fatal(err)
	}
	m := NewMain(styles.Default(), w.Root, w.Initial)
	m.collapsed["."] = false
	m.rebuildRows()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	m.selected = 1
	m.navigationFocus = false
	m.contentOpened = true
	m.contentTabs, m.contentActive = []string{"main.go"}, "main.go"
	return m
}

func pressRune(t *testing.T, m *Main, char rune) {
	t.Helper()
	for _, binding := range m.KeyMap() {
		if binding.Pattern.Rune == char && binding.Pattern.Mod == tui.ModNone {
			binding.Handler(tui.KeyEvent{Key: tui.KeyRune, Rune: char})
			return
		}
	}
	t.Fatalf("missing binding for %q", char)
}

func TestContentTabsAndLiveReload(t *testing.T) {
	m := fileScreen(t, "package main\nvar before = 1\n")
	render(m, 120, 24)
	press(t, m, tui.KeyEnter)
	frame := render(m, 120, 24)
	if !m.contentTab || !strings.Contains(frame, "var before = 1") || !strings.Contains(frame, "[Go]") {
		t.Fatalf("enter did not open highlighted content:\n%s", frame)
	}
	if err := os.WriteFile(filepath.Join(m.root, "main.go"), []byte("package main\nvar after = 2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	m.Apply(watcher.Update{Entries: m.entries, Changes: []watcher.Change{{Path: "main.go", Kind: "modified"}}})
	frame = render(m, 120, 24)
	if !strings.Contains(frame, "var after = 2") || strings.Contains(frame, "var before") {
		t.Fatalf("selected preview was not refreshed:\n%s", frame)
	}
	pressRune(t, m, 'w')
	frame = render(m, 120, 24)
	if !m.contentTab || !m.watchFocus || !strings.Contains(frame, "M main.go") {
		t.Fatal("1 must focus sidebar Watch without replacing Content")
	}
	rect := m.contentTabButton.Rect()
	if !m.HandleMouse(tui.MouseEvent{Button: tui.MouseLeft, X: rect.X, Y: rect.Y}) || !m.contentTab {
		t.Fatal("content tab must be clickable")
	}
	render(m, 120, 24)
	rect = m.watchTabButton.Rect()
	m.HandleMouse(tui.MouseEvent{Button: tui.MouseLeft, X: rect.X, Y: rect.Y})
	if !m.contentTab || !m.watchFocus {
		t.Fatal("Watch must focus sidebar activity without replacing Content")
	}
	rect = m.tree.ContentRect()
	m.HandleMouse(tui.MouseEvent{Button: tui.MouseLeft, X: rect.X, Y: rect.Y + treeRowHeight})
	if !m.contentTab {
		t.Fatal("clicking a file must open Content")
	}
	if err := os.Remove(filepath.Join(m.root, "main.go")); err != nil {
		t.Fatal(err)
	}
	m.Apply(watcher.Update{Entries: m.entries[:1], Changes: []watcher.Change{{Path: "main.go", Kind: "deleted"}}})
	frame = render(m, 120, 24)
	if strings.Contains(frame, "var after") || !strings.Contains(frame, "Select a file") {
		t.Fatal("deleted content must not remain visible")
	}
}

func TestContentNavigationAndReadErrors(t *testing.T) {
	m := fileScreen(t, strings.Repeat("// "+strings.Repeat("long", 50)+"\n", 60))
	pressRune(t, m, '1')
	render(m, 120, 24)
	press(t, m, tui.KeyTab)
	press(t, m, tui.KeyPageDown)
	if m.contentOffset <= 0 {
		t.Fatal("page down must scroll content")
	}
	press(t, m, tui.KeyEnd)
	if m.contentOffset != len(m.preview.lines)-viewport(m.feed) {
		t.Fatal("end must scroll to last line")
	}
	press(t, m, tui.KeyHome)
	if m.contentOffset != 0 {
		t.Fatal("home must scroll to first line")
	}
	pressRune(t, m, ']')
	if m.contentX != 4 {
		t.Fatal("] in Content must scroll horizontally")
	}
	pressRune(t, m, '[')
	if m.contentX != 0 {
		t.Fatal("left in Content must scroll horizontally")
	}
	rect := m.feed.Rect()
	m.HandleMouse(tui.MouseEvent{Button: tui.MouseWheelDown, X: rect.X, Y: rect.Y})
	if m.contentOffset != 3 {
		t.Fatal("wheel must scroll Content, not Watch")
	}
	if err := os.Remove(filepath.Join(m.root, "main.go")); err != nil {
		t.Fatal(err)
	}
	pressRune(t, m, 'r')
	if frame := render(m, 120, 24); !strings.Contains(frame, "Cannot preview") {
		t.Fatalf("read errors must be visible:\n%s", frame)
	}
}

func TestContentLongLinesAndSelection(t *testing.T) {
	m := fileScreen(t, "// "+strings.Repeat("x", 200)+"TAIL\n")
	pressRune(t, m, '1')
	frame := render(m, 120, 24)
	if strings.Contains(frame, "TAIL") {
		t.Fatal("long source line should initially be clipped to the viewport")
	}
	press(t, m, tui.KeyTab)
	for range 100 {
		pressRune(t, m, ']')
	}
	frame = render(m, 120, 24)
	if !strings.Contains(frame, "TAIL") {
		t.Fatalf("horizontal scrolling must reveal end of source:\n%s", frame)
	}
	press(t, m, tui.KeyTab)
	press(t, m, tui.KeyTab)
	press(t, m, tui.KeyHome)
	frame = render(m, 120, 24)
	if strings.Contains(frame, "Select a file") {
		t.Fatal("tree selection must not replace the open file tab")
	}
}
