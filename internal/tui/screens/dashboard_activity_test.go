package screens

import (
	"fmt"
	"strings"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/analysis"
	"github.com/MoeidHeidari/guardrail/internal/config"
	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
	"github.com/MoeidHeidari/guardrail/internal/watcher"
	"github.com/grindlemire/go-tui"
)

func TestSidebarActivityKeepsContentAndScrolls(t *testing.T) {
	m := fileScreen(t, "package main\n")
	m.contentTab = true
	for i := range 60 {
		m.Apply(watcher.Update{Changes: []watcher.Change{{Path: fmt.Sprintf("file%02d", i), Kind: "modified"}}})
	}
	render(m, 120, 32)
	pressRune(t, m, 'w')
	press(t, m, tui.KeyEnd)
	render(m, 120, 32)
	if !m.contentTab || !m.watchFocus || m.changeOffset != 60-viewport(m.watchPane) {
		t.Fatal("sidebar activity must scroll independently without replacing Content")
	}
	press(t, m, tui.KeyHome)
	render(m, 120, 32)
	rect := m.watchPane.ContentRect()
	m.HandleMouse(tui.MouseEvent{Button: tui.MouseWheelDown, X: rect.X + 1, Y: rect.Y})
	if m.changeOffset != 3 || !m.contentTab {
		t.Fatal("wheel must scroll only sidebar activity")
	}
	render(m, 120, 32)
	rect = m.watchPane.ContentRect()
	m.HandleMouse(tui.MouseEvent{Button: tui.MouseLeft, X: rect.Right() - 1, Y: rect.Y})
	render(m, 120, 32)
	m.HandleMouse(tui.MouseEvent{Button: tui.MouseLeft, Action: tui.MouseDrag, X: rect.Right() + 10, Y: rect.Bottom() + 10})
	render(m, 120, 32)
	if m.changeOffset != 60-viewport(m.watchPane) || m.contentOffset != 0 {
		t.Fatal("activity scrollbar drag must remain captured without scrolling Content")
	}
	m.HandleMouse(tui.MouseEvent{Action: tui.MouseRelease})
}

func TestFindingTableAttributionAndVirtualization(t *testing.T) {
	for _, theme := range []styles.Theme{styles.Default(), styles.Light()} {
		m := testScreen()
		m.theme = theme
		m.BeginAnalysis(1)
		result := analysis.LiveResult{Generation: 1, Rules: []config.Rule{{
			ID: "length", Type: "javascript", File: ".guard/rules/max-source-lines.js",
		}}}
		file := analysis.Result{Path: "src/main.go"}
		for i := range 1000 {
			file.Findings = append(file.Findings, analysis.Finding{
				Path: file.Path, Rule: "length", Kind: "violation", Line: i + 1, Column: 1,
				Message: fmt.Sprintf("violation-%04d", i),
			})
		}
		result.Report.Files = []analysis.Result{file}
		m.ApplyAnalysis(result)
		m.openWorkspace(violationsPage)
		frame := render(m, 160, 32)
		for _, text := range []string{"src/main.go:1:1", "Rule set: General / Rule: length",
			".guard/rules.yml -> .guard/rules/max-source-lines.js", "violation-0000"} {
			if !strings.Contains(frame, text) {
				t.Fatalf("missing detailed finding content %q:\n%s", text, frame)
			}
		}
		if len(m.feed.Children()) > 40 {
			t.Fatal("finding table must render a bounded viewport")
		}
		m.changesFocus = true
		press(t, m, tui.KeyEnd)
		frame = render(m, 160, 32)
		if m.findingIndex != 999 || !strings.Contains(frame, "violation-0999") {
			t.Fatalf("last finding must remain reachable:\n%s", frame)
		}
		for _, size := range [][2]int{{120, 24}, {80, 24}, {160, 40}} {
			render(m, size[0], size[1])
			if m.feed.Rect().Height < 1 {
				t.Fatal("table must retain a visible viewport")
			}
			for _, row := range m.feed.Children() {
				if row.BorderTitle() != "" {
					t.Fatal("table must not use border-embedded titles")
				}
			}
			if m.findingIndex != 999 {
				t.Fatal("resize must preserve selected finding")
			}
		}
	}
}
