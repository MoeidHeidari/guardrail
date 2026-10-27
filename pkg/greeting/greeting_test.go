package greeting_test

import (
	"strings"
	"testing"

	"github.com/MoeidHeidari/guardrail/pkg/greeting"
)

func TestBuildDefault(t *testing.T) {
	got := greeting.Build("Alice", false)
	want := "Hello, Alice! 👋"
	if got != want {
		t.Errorf("Build() = %q, want %q", got, want)
	}
}

func TestBuildShout(t *testing.T) {
	got := greeting.Build("Alice", true)
	if got != strings.ToUpper(got) {
		t.Errorf("Build(shout=true) should be uppercase, got %q", got)
	}
	if !strings.Contains(got, "ALICE") {
		t.Errorf("expected 'ALICE' in shouted output, got %q", got)
	}
}

func TestBuildWorld(t *testing.T) {
	got := greeting.Build("World", false)
	if !strings.Contains(got, "World") {
		t.Errorf("expected 'World' in greeting, got %q", got)
	}
}
