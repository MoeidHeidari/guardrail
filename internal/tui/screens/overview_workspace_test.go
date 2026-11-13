package screens

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/MoeidHeidari/guardrail/internal/analysis"
	"github.com/MoeidHeidari/guardrail/internal/config"
	"github.com/MoeidHeidari/guardrail/internal/watcher"
	"github.com/grindlemire/go-tui"
)

func overviewFixture(t *testing.T) *Main {
	t.Helper()
	m := fileScreen(t, "package main\nvar first = 1\nvar second = 2\n")
	m.BeginAnalysis(1)
	m.ApplyAnalysis(analysis.LiveResult{Generation: 1, Rules: []config.Rule{{ID: "test", Set: "quality"}},
		Report: analysis.Report{Checked: 1, Files: []analysis.Result{{Path: "main.go", Findings: []analysis.Finding{
			{Path: "main.go", Rule: "test", Line: 2, Column: 1, Kind: "violation"},
			{Path: "main.go", Rule: "test", Line: 3, Column: 1, Kind: "error"},
		}}}}})
	m.openWorkspace(overviewPage)
	return m
}

func TestOverviewPreviewNavigationAndLayout(t *testing.T) {
	m := overviewFixture(t)
	m.overviewPreviewOpened = true
	frame := render(m, 160, 40)
	for _, text := range []string{"Violations", "Errors", "Affected files", "Rules", "Rule sets",
		"quality / 1 rules", "Live content / main.go", "var first = 1"} {
		if !strings.Contains(frame, text) {
			t.Fatalf("missing dashboard content %q:\n%s", text, frame)
		}
	}
	if m.overviewPreview.Rect().X <= m.overviewList.Rect().Right() {
		t.Fatal("wide dashboard must show preview beside findings")
	}
	spans := m.overviewList.Children()[0].RichText()
	if len(spans) != 3 || spans[1].Text != "[quality]" ||
		!spans[1].Style.Bg.Equal(m.theme.Selection) || !spans[1].Style.Fg.Equal(m.theme.SelectionText) {
		t.Fatal("affected-file rule sets must render as accent tags")
	}
	m.changesFocus = true
	press(t, m, tui.KeyDown)
	render(m, 160, 40)
	if m.contentTab || m.findingIndex != 1 || m.preview.cursor == nil || m.preview.cursor.line != 2 {
		t.Fatal("arrow selection must preview finding location without leaving Overview")
	}
	rect := m.overviewList.ContentRect()
	m.HandleMouse(tui.MouseEvent{Button: tui.MouseLeft, X: rect.X + 1, Y: rect.Y})
	render(m, 160, 40)
	if m.findingIndex != 0 || m.contentTab || m.preview.cursor.line != 1 {
		t.Fatal("click must update embedded preview")
	}
	render(m, 100, 40)
	if m.overviewPreview.Rect().Y <= m.overviewList.Rect().Bottom() {
		t.Fatal("narrow dashboard must stack preview below findings")
	}
}

func TestOverviewAsyncPreviewUpdatesAndDeletion(t *testing.T) {
	m := overviewFixture(t)
	m.overviewPreviewOpened = true
	results, stop := m.StartPreviewWorker(context.Background())
	defer stop()
	render(m, 160, 40)
	accept := func() {
		t.Helper()
		select {
		case result := <-results:
			if !m.ApplyPreview(result) {
				t.Fatal("current embedded preview rejected")
			}
		case <-time.After(3 * time.Second):
			t.Fatal("embedded preview did not finish")
		}
	}
	accept()
	if !strings.Contains(render(m, 160, 40), "var first = 1") {
		t.Fatal("async code not displayed")
	}
	if err := os.WriteFile(filepath.Join(m.root, "main.go"), []byte("package main\nvar updated = 3\n"), 0600); err != nil {
		t.Fatal(err)
	}
	m.Apply(watcher.Update{Changes: []watcher.Change{{Path: "main.go", Kind: "modified"}}})
	render(m, 160, 40)
	accept()
	if !strings.Contains(render(m, 160, 40), "var updated = 3") {
		t.Fatal("live save did not refresh embedded content")
	}
	m.Apply(watcher.Update{Removed: []string{"main.go"}})
	m.BeginAnalysis(2)
	m.ApplyAnalysis(analysis.LiveResult{Generation: 2})
	frame := render(m, 160, 40)
	if strings.Contains(frame, "var updated = 3") || m.overviewPath != "" {
		t.Fatal("deleted file must not retain stale content")
	}
}
