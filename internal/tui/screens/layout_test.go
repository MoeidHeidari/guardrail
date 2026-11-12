package screens

import (
	"fmt"
	"strings"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/watcher"
	"github.com/grindlemire/go-tui"
)

func TestEditorWorkspaceGeometry(t *testing.T) {
	m := testScreen()
	for _, size := range [][2]int{{240, 60}, {160, 48}, {120, 40}, {80, 24}, {60, 16}, {40, 12}, {20, 8}, {80, 24}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			root := m.Render(nil)
			buffer := tui.NewBuffer(size[0], size[1])
			root.RenderTo(buffer, size[0], size[1])
			header, body, footer := root.Children()[0], root.Children()[1], root.Children()[2]
			if header.Rect().Height != 3 || footer.Rect().Height != 3 {
				t.Fatal("header and footer must include top and bottom padding")
			}
			if footer.Rect().Y != size[1]-3 || body.Rect().Y != 3 || body.Rect().Height != size[1]-6 {
				t.Fatalf("chrome and workspace misaligned: header=%+v body=%+v footer=%+v",
					header.Rect(), body.Rect(), footer.Rect())
			}
			if body.Children()[0].Rect().Width > 38 {
				t.Fatal("sidebar must not grow beyond 38 columns")
			}
			content := body.Children()[1].Rect()
			if content.X+content.Width != size[0] {
				t.Fatalf("unused space after content: right edge %d, terminal width %d", content.X+content.Width, size[0])
			}
			feed := m.feed.Rect()
			if feed.X+feed.Width != size[0]-2 {
				t.Fatal("main viewport must reach the right panel's internal padding")
			}
			bounds := tui.Rect{Width: size[0], Height: size[1]}
			for _, el := range []*tui.Element{header, body, footer, body.Children()[0], body.Children()[1], m.tree, m.feed} {
				if !bounds.ContainsRect(el.Rect()) {
					t.Fatalf("element outside terminal: %+v in %+v", el.Rect(), bounds)
				}
			}
			if size[0] >= 80 {
				if viewport(m.feed) < size[1]-13 || viewport(m.tree) < max(1, size[1]-25) {
					t.Fatalf("too much vertical space consumed: tree=%d feed=%d", viewport(m.tree), viewport(m.feed))
				}
				for _, text := range []string{"GUARDRAIL", "1 Watch", "2 Content", "Watching", "Context", "Tab focus"} {
					if !strings.Contains(buffer.StringTrimmed(), text) {
						t.Errorf("missing %q:\n%s", text, buffer.StringTrimmed())
					}
				}
			}
		})
	}
}

func TestWorkspaceFocusAndTabStyling(t *testing.T) {
	m := testScreen()
	frame := render(m, 120, 24)
	if !strings.Contains(frame, "/   Context") {
		t.Fatal("status bar must identify tree focus")
	}
	if !m.dashboardButton.TextStyle().Fg.Equal(m.theme.Accent) {
		t.Fatal("active workspace tab must have an accent underline")
	}
	pressRune(t, m, '1')
	frame = render(m, 120, 24)
	if !strings.Contains(frame, "/   Watch") {
		t.Fatal("status bar must follow pane focus")
	}
	selected := m.tree.Children()[m.selected]
	if !selected.Background().Bg.Equal(m.theme.Chrome) {
		t.Fatal("tree selection must become subtle when the other pane has focus")
	}
	pressRune(t, m, '2')
	m.changesFocus = true
	frame = render(m, 120, 24)
	if !strings.Contains(frame, "/   Content") || !m.contentTabButton.TextStyle().Fg.Equal(m.theme.Accent) {
		t.Fatal("tab switch must update active tab and status bar")
	}
}

func TestWatchChangeColorsAndSummary(t *testing.T) {
	m := testScreen()
	m.Apply(watcher.Update{Entries: m.entries, Changes: []watcher.Change{
		{Path: "src/main.go", Kind: "modified"},
		{Path: "README.md", Kind: "created"},
		{Path: "deleted.txt", Kind: "deleted"},
	}})
	frame := render(m, 120, 24)
	if !strings.Contains(frame, "2 changed") {
		t.Fatalf("status bar must summarize changed files:\n%s", frame)
	}
	for i, color := range []tui.Color{m.theme.Danger, m.theme.Added, m.theme.Modified} {
		spans := m.watchPane.Children()[i].RichText()
		if len(spans) != 3 || !spans[1].Style.Fg.Equal(color) {
			t.Fatal("feed must distinguish event kinds with separate colors")
		}
	}
}
