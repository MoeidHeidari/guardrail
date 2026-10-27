package screens_test

import (
	"strings"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/tui/screens"
	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
	"github.com/grindlemire/go-tui"
)

func TestMainScreen(t *testing.T) {
	screen := screens.NewMain(styles.Default())
	root := screen.Render(nil)
	buffer := tui.NewBuffer(80, 24)
	root.RenderTo(buffer, 80, 24)
	frame := buffer.StringTrimmed()
	for _, text := range []string{"GUARDRAIL", "Workspace", "Overview", "q / ctrl+c quit"} {
		if !strings.Contains(frame, text) {
			t.Errorf("missing %q:\n%s", text, frame)
		}
	}
	body := root.Children()[1]
	content := body.Children()[1]
	if content.Text() != "" || len(content.Children()) != 1 || content.Children()[0].Text() != "Overview" {
		t.Fatal("main content should be empty and ready for composition")
	}
	if strings.ContainsAny(frame, "╭╮╰╯│─") {
		t.Fatal("main screen should not render borders")
	}
	header := root.Children()[0].Rect()
	area := content.Rect()
	if area.X+area.Width != header.X+header.Width {
		t.Fatal("content should extend to the header's right edge")
	}
}

func TestMainQuitBindings(t *testing.T) {
	bindings := screens.NewMain(styles.Default()).KeyMap()
	if len(bindings) != 2 {
		t.Fatalf("got %d key bindings, want 2", len(bindings))
	}
	if bindings[0].Pattern.Rune != 'q' || bindings[0].Handler == nil {
		t.Fatal("missing q quit binding")
	}
	if bindings[1].Pattern.Rune != 'c' || bindings[1].Pattern.Mod != tui.ModCtrl || bindings[1].Handler == nil {
		t.Fatal("missing ctrl+c quit binding")
	}
}
