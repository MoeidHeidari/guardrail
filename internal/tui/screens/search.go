package screens

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/MoeidHeidari/guardrail/internal/tui/components"
	"github.com/grindlemire/go-tui"
)

func (m *Main) openSearch() {
	m.searchOpen, m.animationOnly = true, false
	m.searchQuery, m.searchSelection = nil, 0
	m.searchPreviewFocus = false
	m.dragPane = nil
}

func (m *Main) searchMatches() []string {
	query := strings.ToLower(string(m.searchQuery))
	var paths []string
	for _, entry := range m.entries {
		if !entry.IsDir() && strings.Contains(strings.ToLower(filepath.ToSlash(entry.Path)), query) {
			paths = append(paths, entry.Path)
		}
	}
	sort.Strings(paths)
	return paths
}

func (m *Main) moveSearch(delta int) {
	m.searchSelection = max(0, min(m.searchSelection+delta, len(m.searchMatches())-1))
	m.animationOnly = false
}

func (m *Main) openSearchResult() {
	paths := m.searchMatches()
	if len(paths) == 0 {
		return
	}
	path := paths[min(m.searchSelection, len(paths)-1)]
	for parent := filepath.Dir(path); ; parent = filepath.Dir(parent) {
		m.collapsed[parent] = false
		if parent == "." {
			break
		}
	}
	m.flattenRows()
	m.selected = m.rowPositions[path]
	m.showSearchPreview(path)
	m.searchPreviewFocus, m.animationOnly = true, false
	m.reveal()
}

func (m *Main) searchKeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnStop(tui.KeyCtrlC, func(event tui.KeyEvent) { event.App().Stop() }),
		tui.OnStop(tui.KeyEscape, func(tui.KeyEvent) {
			if m.searchPreviewFocus {
				m.searchPreviewFocus = false
			} else {
				m.searchOpen = false
			}
			m.animationOnly = false
		}),
		tui.OnStop(tui.KeyLeft, func(tui.KeyEvent) { m.searchPreviewFocus = false }),
		tui.OnStop(tui.KeyUp, func(tui.KeyEvent) { m.moveSearchOrPreview(-1) }),
		tui.OnStop(tui.KeyDown, func(tui.KeyEvent) { m.moveSearchOrPreview(1) }),
		tui.OnStop(tui.KeyHome, func(tui.KeyEvent) { m.searchEdge(false) }),
		tui.OnStop(tui.KeyEnd, func(tui.KeyEvent) { m.searchEdge(true) }),
		tui.OnStop(tui.KeyPageUp, func(tui.KeyEvent) { m.moveSearchOrPreview(-max(1, m.renderHeight-12)) }),
		tui.OnStop(tui.KeyPageDown, func(tui.KeyEvent) { m.moveSearchOrPreview(max(1, m.renderHeight-12)) }),
		tui.OnStop(tui.KeyEnter, func(tui.KeyEvent) { m.openSearchResult() }),
		tui.OnStop(tui.KeyRight, func(tui.KeyEvent) { m.openSearchResult() }),
		tui.OnStop(tui.Rune('['), func(tui.KeyEvent) { m.searchHorizontal('[', -4) }),
		tui.OnStop(tui.Rune(']'), func(tui.KeyEvent) { m.searchHorizontal(']', 4) }),
		tui.OnStop(tui.KeyBackspace, func(tui.KeyEvent) {
			m.searchPreviewFocus = false
			if len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				m.searchSelection = 0
			}
		}),
		tui.OnStop(tui.AnyRune, func(event tui.KeyEvent) {
			m.searchQuery = append(m.searchQuery, event.Rune)
			m.searchPreviewFocus = false
			m.searchSelection, m.animationOnly = 0, false
		}),
		tui.OnStop(tui.AnyKey, func(tui.KeyEvent) {}),
	}
}

func (m *Main) renderSearch() *tui.Element {
	paths := m.searchMatches()
	m.searchSelection = max(0, min(m.searchSelection, len(paths)-1))
	title := m.label(fmt.Sprintf("Files / %d matches", len(paths)))
	m.highlightFocus(title, true)
	content := components.NewSection(m.theme, "", title)
	content.Apply(tui.WithPaddingTRBL(1, 2, 1, 2))
	results := components.NewContainer(m.theme, tui.WithWidthPercent(40), tui.WithHeightPercent(100),
		tui.WithMinWidth(0), tui.WithFlexShrink(1))
	body := components.NewContainer(m.theme, tui.WithDirection(tui.Row), tui.WithGap(2),
		tui.WithHeight(0), tui.WithFlexGrow(1))
	body.AddChild(results, m.renderSearchPreview())
	content.AddChild(body)
	m.searchRows, m.searchPaths = nil, nil
	limit := max(1, m.renderHeight-12)
	start := max(0, m.searchSelection-limit+1)
	for i := start; i < min(len(paths), start+limit); i++ {
		row := m.label(display(filepath.ToSlash(paths[i])))
		m.highlightFocus(row, i == m.searchSelection)
		results.AddChild(row)
		m.searchRows = append(m.searchRows, row)
		m.searchPaths = append(m.searchPaths, paths[i])
	}
	if len(paths) == 0 {
		results.AddChild(m.label("No matching files."))
	}
	return content
}

func (m *Main) handleSearchMouse(event tui.MouseEvent) bool {
	if event.Action == tui.MousePress && event.Button == tui.MouseLeft &&
		m.searchButton != nil && m.searchButton.Rect().Contains(event.X, event.Y) {
		if !m.searchOpen {
			m.openSearch()
		}
		return true
	}
	if !m.searchOpen {
		return false
	}
	if event.Action == tui.MousePress {
		if m.searchViewer != nil && m.searchViewer.feed != nil &&
			m.searchViewer.feed.Rect().Contains(event.X, event.Y) {
			m.searchPreviewFocus = true
			switch event.Button {
			case tui.MouseWheelUp:
				m.moveSearchOrPreview(-3)
			case tui.MouseWheelDown:
				m.moveSearchOrPreview(3)
			}
			return true
		}
		switch event.Button {
		case tui.MouseWheelUp:
			m.searchPreviewFocus = false
			m.moveSearch(-3)
		case tui.MouseWheelDown:
			m.searchPreviewFocus = false
			m.moveSearch(3)
		case tui.MouseLeft:
			for i, row := range m.searchRows {
				if row.Rect().Contains(event.X, event.Y) {
					paths := m.searchMatches()
					for j, path := range paths {
						if path == m.searchPaths[i] {
							m.searchSelection = j
							m.openSearchResult()
							return true
						}
					}
				}
			}
		}
	}
	return true
}
