package screens

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/watcher"
	"github.com/grindlemire/go-tui"
)

func TestLocateLiveChange(t *testing.T) {
	for _, test := range []struct {
		name, before, after string
		line, column        int
	}{
		{"append", "hello", "hello world", 0, 11},
		{"newline", "first\n", "first\nsecond\n", 1, 6},
		{"middle", "hello world", "hello nice world", 0, 11},
		{"delete", "hello world", "hello", 0, 5},
		{"empty", "hello", "", 0, 0},
		{"unicode", "a", "a你好", 0, 3},
		{"tabs", "a", "a\tb", 0, 6},
		{"multiline", "one\ntwo\nthree", "one\nchanged\nthree", 1, 7},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := locateChange(test.before, test.after)
			if got == nil || got.line != test.line || got.column != test.column {
				t.Fatalf("got %+v, want line %d column %d", got, test.line, test.column)
			}
		})
	}
	if locateChange("same", "same") != nil {
		t.Fatal("metadata-only updates must not invent a cursor")
	}
}

func TestLiveCursorFollowsAndHighlights(t *testing.T) {
	before := strings.Repeat("// before\n", 200)
	m := fileScreen(t, before)
	m.contentTab = true
	render(m, 120, 24)
	if m.preview.cursor != nil {
		t.Fatal("initial preview must not invent an edit")
	}
	after := before + "// latest live edit\n"
	if err := os.WriteFile(filepath.Join(m.root, "main.go"), []byte(after), 0o600); err != nil {
		t.Fatal(err)
	}
	m.Apply(watcher.Update{Entries: m.entries, Changes: []watcher.Change{{Path: "main.go", Kind: "modified"}}})
	frame := render(m, 120, 24)
	if m.contentOffset == 0 || !strings.Contains(frame, "latest live edit") || !strings.Contains(frame, "201>") {
		t.Fatalf("live cursor must follow latest changed line:\n%s", frame)
	}
	found := false
	for _, el := range m.feed.Children() {
		for _, span := range el.RichText() {
			if span.Style.Bg.Equal(m.theme.Activity) {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("changed position must have a contrasting cursor block")
	}
	// A later metadata-only update must not pull the view back to the edit.
	m.contentOffset = 0
	m.previewDirty = true
	render(m, 120, 24)
	if m.contentOffset != 0 {
		t.Fatal("unchanged content must not force scroll")
	}
}

func TestCursorSpansPreserveHighlightedText(t *testing.T) {
	m := testScreen()
	line := []tui.TextSpan{{Text: "hello "}, {Text: "你好"}}
	for _, column := range []int{0, 6, 8} {
		spans := m.cursorSpans(line, column)
		text := ""
		for _, span := range spans {
			text += span.Text
		}
		want := "hello 你好"
		if column == 8 {
			want += " "
		}
		if text != want {
			t.Fatalf("cursor corrupted text: %q, want %q", text, want)
		}
	}
}
