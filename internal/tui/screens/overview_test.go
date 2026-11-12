package screens

import (
	"strings"
	"testing"
	"time"

	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
	"github.com/MoeidHeidari/guardrail/internal/watcher"
	"github.com/grindlemire/go-tui"
)

func TestReferenceOverviewPanels(t *testing.T) {
	for _, theme := range []styles.Theme{styles.Default(), styles.Light()} {
		m := NewMain(theme, "/workspace", watcher.Update{Entries: testScreen().entries})
		m.Apply(watcher.Update{Changes: []watcher.Change{
			{Path: "example.go", Kind: "modified", Time: time.Now()},
		}})
		frame := render(m, 120, 32)
		for _, text := range []string{"Workspace", "Engineering status", "Affected files", "Errors",
			"example.go", "Add .guard/rules.yml", "Watching"} {
			if !strings.Contains(frame, text) {
				t.Fatalf("missing reference panel content %q:\n%s", text, frame)
			}
		}
		if len(m.feed.Children()) != 3 {
			t.Fatal("Overview must compose three distinct panels")
		}
		for _, card := range m.feed.Children()[1:2] {
			if card.BorderTitle() != "" || card.Children()[0].TextAlign() != tui.TextAlignLeft {
				t.Fatal("panel headings must be left-aligned inside uninterrupted borders")
			}
			if !card.Border().All(tui.BorderRounded) ||
				!card.Background().Bg.Equal(theme.Background) {
				t.Fatal("overview panels must use thin theme-matched outlines")
			}
		}
	}
}

func TestOverviewFitsNormalTerminal(t *testing.T) {
	m := NewMain(styles.Default(), "/workspace", watcher.Update{Entries: testScreen().entries})
	for range 20 {
		m.Apply(watcher.Update{Changes: []watcher.Change{{Path: "latest.go", Kind: "modified"}}})
	}
	frame := render(m, 120, 24)
	if !strings.Contains(frame, "latest.go") {
		t.Fatal("sidebar activity must remain visible at standard terminal height")
	}
}
