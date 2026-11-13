package screens

import (
	"strings"
	"testing"

	"github.com/grindlemire/go-tui"
)

func TestFooterFollowsFocus(t *testing.T) {
	m := testScreen()
	for _, test := range []struct {
		name string
		set  func()
		want string
	}{
		{"context", m.focusContext, "Enter/Right open"},
		{"menu", func() { m.navigationFocus = true }, "Enter/Right open"},
		{"violations", func() { m.openWorkspaceFocused(violationsPage) }, "Left/Right filter"},
		{"rules", func() { m.openWorkspaceFocused(rulesPage) }, "Space fold"},
		{"content", func() { m.openTab("README.md"); m.focusMain() }, "< > tabs"},
		{"watch", func() { m.openNavigation(len(workspaceLabels)) }, "Enter/Right open"},
	} {
		t.Run(test.name, func(t *testing.T) {
			test.set()
			if hint := m.footerHint(120); !strings.Contains(hint, test.want) {
				t.Fatalf("missing %q in footer: %s", test.want, hint)
			}
			if frame := render(m, 120, 40); !strings.Contains(frame, test.want) {
				t.Fatalf("shortcut %q is clipped in rendered footer:\n%s", test.want, frame)
			}
		})
	}
}

func TestFooterFitsTerminal(t *testing.T) {
	m := testScreen()
	for _, width := range []int{20, 40, 54, 55, 80, 99, 100, 120, 160, 240} {
		frame := render(m, width, 40)
		hint := m.footerHint(width)
		if tui.StringWidth(hint) > width-4 {
			t.Errorf("footer exceeds width %d: %s", width, hint)
		}
		if !strings.Contains(frame, "q quit") {
			t.Errorf("quit guide is not visible at width %d", width)
		}
	}
}
