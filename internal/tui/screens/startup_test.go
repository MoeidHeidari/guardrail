package screens

import (
	"errors"
	"strings"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
	"github.com/grindlemire/go-tui"
)

func renderStartup(component tui.Component) string {
	buffer := tui.NewBuffer(80, 24)
	component.Render(nil).RenderTo(buffer, 80, 24)
	return buffer.StringTrimmed()
}

func TestPathPickerShowsCurrentDirectoryAndValidatesSelection(t *testing.T) {
	picker := NewPathPicker(styles.Default(), "/workspace", func(path string) (string, error) {
		if path != "/workspace" {
			return "", errors.New("context must be a directory")
		}
		return path, nil
	})
	frame := renderStartup(picker)
	for _, text := range []string{"GUARDRAIL", "Choose a context", "PROJECT DIRECTORY", "/workspace", "[Enter] Start watching"} {
		if !strings.Contains(frame, text) {
			t.Errorf("path picker missing %q:\n%s", text, frame)
		}
	}
	picker.submit(" ")
	if picker.err == "" || picker.selected != "" {
		t.Fatal("empty context must not be accepted")
	}
	picker.submit("/invalid")
	if picker.err == "" || picker.selected != "" {
		t.Fatal("invalid context must remain in the picker")
	}
	picker.submit("/workspace")
	if picker.err != "" || picker.Selected() != "/workspace" {
		t.Fatalf("valid context was not selected: %q, %q", picker.Selected(), picker.err)
	}
}

func TestPathPickerEscapeCancelsBeforeInputHandling(t *testing.T) {
	for _, binding := range NewPathPicker(styles.Default(), "/", func(path string) (string, error) {
		return path, nil
	}).KeyMap() {
		if binding.Pattern.Key == tui.KeyEscape && binding.Preempt {
			return
		}
	}
	t.Fatal("Escape must cancel the picker before the focused input handles it")
}

func TestSetupDialogExplainsConfigurationAndDocumentation(t *testing.T) {
	dialog := NewSetupDialog(styles.Default(), "/workspace", errors.New("no YAML rule sets found in .guard"))
	frame := renderStartup(dialog)
	for _, text := range []string{
		"Rules are not ready", "/workspace", "no YAML rule sets found",
		"CONFIGURATION ERROR", "Fix the rule configuration", "docs/GETTING_STARTED.md", "docs/RULES.md",
		"NEXT STEPS", "[q] Quit",
	} {
		if !strings.Contains(frame, text) {
			t.Errorf("setup dialog missing %q:\n%s", text, frame)
		}
	}
}
