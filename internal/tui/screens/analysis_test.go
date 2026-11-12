package screens

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MoeidHeidari/guardrail/internal/analysis"
	"github.com/MoeidHeidari/guardrail/internal/config"
	"github.com/grindlemire/go-tui"
)

func TestAnalysisStaleResultsAndPages(t *testing.T) {
	m := testScreen()
	m.BeginAnalysis(2)
	result := analysis.LiveResult{Generation: 1, Rules: []config.Rule{{ID: "todo", Type: "regex"}},
		Report: analysis.Report{Checked: 1, Files: []analysis.Result{{Path: "src/main.go",
			Findings: []analysis.Finding{{Path: "src/main.go", Rule: "todo", Kind: "violation", Line: 2, Column: 3, Message: "remove TODO"}}}}}}
	if m.ApplyAnalysis(result) || !m.analysisPending {
		t.Fatal("stale report must be rejected")
	}
	result.Generation = 2
	if !m.ApplyAnalysis(result) || m.analysisPending {
		t.Fatal("current report must be accepted")
	}
	m.openWorkspace(violationsPage)
	frame := render(m, 120, 32)
	if !strings.Contains(frame, "remove TODO") || !strings.Contains(frame, "1 violations") {
		t.Fatalf("findings were not rendered:\n%s", frame)
	}
	m.openWorkspace(rulesPage)
	render(m, 120, 40)
	m.changesFocus = true
	m.move(1)
	if !strings.Contains(render(m, 120, 40), "todo: No description provided") {
		t.Fatal("configured rule was not rendered")
	}
	m.BeginAnalysis(3)
	m.ApplyAnalysis(analysis.LiveResult{Generation: 3, Error: "invalid configuration"})
	m.openWorkspace(violationsPage)
	if !strings.Contains(render(m, 120, 32), "[error]") || len(m.findings) != 1 {
		t.Fatal("configuration errors must appear as report errors")
	}
}

func TestFindingKeyboardAndMouseOpenLocation(t *testing.T) {
	for _, mouse := range []bool{false, true} {
		m := fileScreen(t, "package main\n// TODO\n")
		m.BeginAnalysis(1)
		path := m.rows[m.selected].entry.Path
		m.ApplyAnalysis(analysis.LiveResult{Generation: 1, Report: analysis.Report{Files: []analysis.Result{{
			Path: path, Findings: []analysis.Finding{{Path: path, Rule: "todo", Kind: "violation", Line: 2, Column: 4}},
		}}}})
		m.openWorkspace(violationsPage)
		render(m, 120, 32)
		m.changesFocus = true
		if mouse {
			rect := m.feed.ContentRect()
			m.HandleMouse(tui.MouseEvent{Action: tui.MousePress, Button: tui.MouseLeft, X: rect.X + 1, Y: rect.Y})
		} else {
			press(t, m, tui.KeyEnter)
		}
		render(m, 120, 32)
		if !m.contentTab || m.preview.cursor == nil || m.preview.cursor.line != 1 || m.preview.cursor.column != 3 {
			t.Fatalf("finding did not open exact location: %+v", m.preview.cursor)
		}
	}
}

func TestResolvedReportClearsFindings(t *testing.T) {
	m := testScreen()
	m.BeginAnalysis(1)
	m.ApplyAnalysis(analysis.LiveResult{Generation: 1, Error: "missing config"})
	m.BeginAnalysis(2)
	m.ApplyAnalysis(analysis.LiveResult{Generation: 2, Report: analysis.Report{Checked: 2}})
	if len(m.findings) != 0 || !strings.Contains(m.analysisStatus(), "0 violations") {
		t.Fatal("resolved findings must be removed")
	}
}

func TestAnalysisJumpAsyncPreview(t *testing.T) {
	m := fileScreen(t, "package main\n// TODO\n")
	if err := os.WriteFile(filepath.Join(m.root, "main.go"), []byte("package main\n// TODO\n"), 0600); err != nil {
		t.Fatal(err)
	}
	m.jumpLine, m.jumpColumn = 2, 3
	m.previewID = 1
	m.ApplyPreview(PreviewResult{id: 1, preview: filePreview{path: "main.go", code: "package main\n// TODO\n",
		lines: make([][]tui.TextSpan, 3)}})
	if m.jumpLine != 0 || m.preview.cursor == nil || m.preview.cursor.line != 1 {
		t.Fatal("async preview did not preserve finding jump")
	}
}
