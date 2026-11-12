package screens

import (
	"fmt"
	"path/filepath"

	"github.com/MoeidHeidari/guardrail/internal/analysis"
	"github.com/MoeidHeidari/guardrail/internal/config"
	"github.com/grindlemire/go-tui"
)

const findingHeight = 1

func (m *Main) BeginAnalysis(generation uint64) {
	m.analysisGeneration, m.analysisPending = generation, true
	m.animationOnly = false
}

func (m *Main) ApplyAnalysis(result analysis.LiveResult) bool {
	if result.Generation != m.analysisGeneration {
		return false
	}
	m.analysisPending = false
	m.analysisResult = result
	m.findings = result.Report.Findings()
	if result.Error != "" {
		m.findings = append(m.findings, analysis.ErrorFinding(config.FileName, "configuration", fmt.Errorf("%s", result.Error)))
	}
	m.findingIndex = min(m.findingIndex, max(0, len(m.findings)-1))
	if m.findingFilter != "" {
		indices := m.visibleFindings()
		valid := false
		for _, index := range indices {
			valid = valid || index == m.findingIndex
		}
		if !valid && len(indices) > 0 {
			m.findingIndex = indices[0]
		}
	}
	if m.overviewPath != "" {
		if _, exists := m.entryPositions[m.overviewPath]; !exists {
			m.overviewPath = ""
			m.previewDirty = true
		}
	}
	pane := m.feed
	if m.overviewVisible() {
		pane = m.overviewList
		m.selectOverviewFinding()
	}
	m.findingOffset = min(m.findingOffset, max(0, len(m.visibleFindings())*findingHeight-viewport(pane)))
	m.animationOnly = false
	return true
}

func (m *Main) analysisStatus() string {
	if m.analysisPending {
		return "Analyzing changes... previous results may be outdated."
	}
	if m.analysisResult.Error != "" {
		return "Analysis unavailable: " + display(m.analysisResult.Error)
	}
	if m.analysisGeneration == 0 {
		return "Add .guard/rules.yml to configure engineering analysis."
	}
	violations, failures := 0, 0
	for _, finding := range m.findings {
		if finding.Kind == "error" {
			failures++
		} else {
			violations++
		}
	}
	return fmt.Sprintf("%d checked files / %d violations / %d execution errors",
		m.analysisResult.Report.Checked, violations, failures)
}

func (m *Main) findingsVisible() bool {
	return m.dashboardVisible() && m.workspacePage == violationsPage
}

func (m *Main) renderFindings() {
	indices := m.visibleFindings()
	if len(indices) == 0 {
		if m.findingFilter != "" {
			m.feed.AddChild(m.label("No findings match this filter."))
			return
		}
		if !m.analysisPending && m.analysisGeneration > 0 && m.analysisResult.Error == "" {
			m.feed.AddChild(m.label("No violations found in the analyzed scope."))
		}
		return
	}
	start := min(max(0, m.findingOffset/findingHeight), len(indices))
	end := min(len(indices), start+max(20, m.renderHeight/findingHeight+2))
	spacer := func(height int) {
		if height > 0 {
			m.feed.AddChild(tui.New(tui.WithHeight(height), tui.WithFlexShrink(0)))
		}
	}
	spacer(start * findingHeight)
	for i := start; i < end; i++ {
		index := indices[i]
		finding := m.findings[index]
		set, _ := m.findingSource(finding)
		m.feed.AddChild(m.findingTableRow([]string{finding.Kind, findingLocation(finding),
			set, finding.Rule, finding.Message}, index == m.findingIndex, false))
	}
	spacer((len(indices) - end) * findingHeight)
}

func (m *Main) findingSource(finding analysis.Finding) (string, string) {
	id := finding.Rule
	for _, rule := range m.analysisResult.Rules {
		if rule.ID == id && (finding.Set == "" || finding.Set == rule.Set) {
			source := rule.Source
			if source == "" {
				source = config.FileName
			}
			if rule.File != "" {
				source += " -> " + rule.File
			}
			set := rule.Set
			if set == "" {
				set = "General"
			}
			return set, source
		}
	}
	if id == "configuration" {
		return "System", config.FileName
	}
	if id == "parser" {
		return "System", "Tree-sitter"
	}
	return "System", id
}

func (m *Main) openFinding() {
	if m.findingsVisible() && len(m.visibleFindings()) == 0 {
		return
	}
	if len(m.findings) == 0 {
		return
	}
	finding := m.findings[m.findingIndex]
	path := filepath.FromSlash(finding.Path)
	if _, exists := m.entryPositions[path]; !exists {
		return
	}
	for parent := filepath.Dir(path); ; parent = filepath.Dir(parent) {
		m.collapsed[parent] = false
		if parent == "." {
			break
		}
	}
	m.flattenRows()
	m.selected = m.rowPositions[path]
	m.contentTab, m.contentOpened, m.changesFocus, m.navigationFocus = true, true, true, false
	m.previewDirty = true
	m.jumpLine, m.jumpColumn = max(1, finding.Line), max(1, finding.Column)
	m.reveal()
}

func (m *Main) applyFindingJump() {
	if m.jumpLine == 0 || m.preview.message != "" {
		return
	}
	m.preview.cursor = &changeCursor{line: min(m.jumpLine-1, max(0, len(m.preview.lines)-1)), column: m.jumpColumn - 1}
	m.followCursor()
	m.jumpLine, m.jumpColumn = 0, 0
}
