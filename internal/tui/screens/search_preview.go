package screens

import (
	"github.com/MoeidHeidari/guardrail/internal/tui/components"
	"github.com/MoeidHeidari/guardrail/internal/watcher"
	"github.com/grindlemire/go-tui"
)

func (m *Main) showSearchPreview(path string) {
	if m.searchViewer == nil {
		m.searchViewer = &Main{theme: m.theme, root: m.root}
	}
	viewer := m.searchViewer
	if viewer.preview.path != path {
		viewer.contentOffset, viewer.contentX = 0, 0
	}
	entry := m.entries[m.entryPositions[path]]
	viewer.rows = []treeRow{{entry: entry}}
	viewer.previewDirty = true
}

func (m *Main) moveSearchOrPreview(delta int) {
	if m.searchPreviewFocus && m.searchViewer != nil {
		viewer := m.searchViewer
		viewer.contentOffset = max(0, min(viewer.contentOffset+delta,
			max(0, len(viewer.preview.lines)-viewport(viewer.feed))))
		return
	}

	m.moveSearch(delta)
}

func (m *Main) searchHorizontal(char rune, delta int) {
	if m.searchPreviewFocus && m.searchViewer != nil {
		m.searchViewer.scrollContentX(delta)
		return
	}
	m.searchQuery = append(m.searchQuery, char)
	m.searchSelection = 0
}

func (m *Main) searchEdge(end bool) {
	distance := len(m.entries)
	if m.searchPreviewFocus && m.searchViewer != nil {
		distance = len(m.searchViewer.preview.lines)
	}
	if !end {
		distance = -distance
	}
	m.moveSearchOrPreview(distance)
}

func (m *Main) renderSearchPreview() *tui.Element {
	pane := components.NewContainer(m.theme, tui.WithWidth(0), tui.WithFlexGrow(1),
		tui.WithMinWidth(0), tui.WithHeightPercent(100))
	if m.searchViewer == nil {
		pane.AddChild(m.label("Enter or Right to preview a file"))
		return pane
	}
	viewer := m.searchViewer
	path := viewer.rows[0].entry.Path
	if i, exists := m.entryPositions[path]; exists {
		entry := m.entries[i]
		if entry != viewer.rows[0].entry {
			viewer.previewDirty = true
			viewer.rows[0].entry = entry
		}
	} else {
		viewer.rows[0].entry = watcher.Entry{Path: path}
		viewer.previewDirty = true
	}
	title := m.label(display(path))
	m.highlightFocus(title, m.searchPreviewFocus)
	viewer.renderHeight = m.renderHeight
	viewer.feed = components.NewContainer(m.theme, tui.WithHeight(0), tui.WithFlexGrow(1),
		tui.WithScrollable(tui.ScrollBoth), tui.WithScrollOffset(viewer.contentX, viewer.contentOffset))
	viewer.loadPreview()
	viewer.renderPreview()
	pane.AddChild(title, viewer.feed)
	return pane
}
