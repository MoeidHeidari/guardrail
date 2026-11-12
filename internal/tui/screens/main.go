package screens

import (
	"github.com/MoeidHeidari/guardrail/internal/analysis"
	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
	"github.com/MoeidHeidari/guardrail/internal/watcher"
	"github.com/grindlemire/go-tui"
	"path/filepath"
	"strings"
	"time"
)

const historyLimit = 200

type treeRow struct {
	entry watcher.Entry
	depth int
}

type Main struct {
	overviewPreviewOpened, rulesPreviewOpened bool
	overviewLine, overviewColumn              int
	collapsedRuleSets                         map[string]bool
	selectedRuleSet                           string
	rulesList                                 *tui.Element
	rulesFiles                                *tui.Element
	rulesPreview                              *tui.Element
	rulesSelection                            int
	rulesFileSelection                        int
	rulesFocus                                int
	rulesItems                                []ruleKey
	rulesMatches                              []int
	rulesPath                                 string
	findingX                                  int
	findingFilter                             string
	filterButtons                             []*tui.Element
	tableWidths                               []int
	overviewList                              *tui.Element
	overviewPreview                           *tui.Element
	overviewPath                              string
	overviewListX                             int
	overviewPreviewFocus                      bool
	renderWidth                               int
	watchPane                                 *tui.Element
	watchFocus                                bool
	theme                                     styles.Theme
	root                                      string
	entries                                   []watcher.Entry
	rows                                      []treeRow
	collapsed                                 map[string]bool
	changes                                   []watcher.Change
	selected                                  int
	treeOffset                                int
	changeOffset                              int
	changesFocus                              bool
	tree                                      *tui.Element
	feed                                      *tui.Element
	status                                    map[string]string
	contentTab                                bool
	contentOpened                             bool
	dashboardTab                              bool
	dashboardButton                           *tui.Element
	workspacePage                             workspacePage
	navigationFocus                           bool
	navigationIndex                           int
	navigationItems                           []*tui.Element
	contentOffset                             int
	contentX                                  int
	preview                                   filePreview
	previewDirty                              bool
	watchTabButton                            *tui.Element
	contentTabButton                          *tui.Element
	animationFrame                            int
	activeUntil                               map[string]time.Time
	animationRows                             map[string]*tui.Element
	animationText                             map[string]string
	cachedRoot                                *tui.Element
	animationOnly                             bool
	dragPane                                  *tui.Element
	dragGrab                                  int
	renderHeight                              int
	animationIndent                           map[string]string
	entryPositions                            map[string]int
	rowPositions                              map[string]int
	changedFiles                              int
	previewRequests                           chan previewRequest
	previewID                                 uint64
	requestedPath                             string
	treeChildren                              map[string][]string
	analysisGeneration                        uint64
	analysisPending                           bool
	analysisResult                            analysis.LiveResult
	findingIndex                              int
	findingOffset                             int
	findings                                  []analysis.Finding
	jumpLine                                  int
	jumpColumn                                int
}

func NewMain(theme styles.Theme, root string, initial watcher.Update) *Main {
	m := &Main{theme: theme, root: root, dashboardTab: true, navigationFocus: true, collapsed: map[string]bool{".": false}, status: make(map[string]string),
		activeUntil: make(map[string]time.Time)}
	m.Apply(initial)
	return m
}

func (m *Main) InvalidateRender() {
	m.animationOnly = false
}

func (m *Main) Apply(update watcher.Update) {
	m.animationOnly = false
	selectedPath := "."
	if len(m.rows) > 0 {
		selectedPath = m.rows[m.selected].entry.Path
	}
	structural := update.Entries != nil
	if structural {
		m.entries = append([]watcher.Entry(nil), update.Entries...)
		m.entryPositions = make(map[string]int, len(m.entries))
		for i, entry := range m.entries {
			m.entryPositions[entry.Path] = i
		}
	} else {
		for _, path := range update.Removed {
			if i, exists := m.entryPositions[path]; exists {
				last := len(m.entries) - 1
				m.entries[i] = m.entries[last]
				m.entryPositions[m.entries[i].Path] = i
				m.entries[last] = watcher.Entry{}
				m.entries = m.entries[:last]
				delete(m.entryPositions, path)
				delete(m.status, path)
				delete(m.collapsed, path)
				structural = true
			}
		}
		for _, entry := range update.Upsert {
			if i, exists := m.entryPositions[entry.Path]; exists {
				structural = structural || m.entries[i].IsDir() != entry.IsDir()
				m.entries[i] = entry
				if row, visible := m.rowPositions[entry.Path]; visible {
					m.rows[row].entry = entry
				}
			} else {
				m.entryPositions[entry.Path] = len(m.entries)
				m.entries = append(m.entries, entry)
				structural = true
			}
		}
	}
	directories := update.Upsert
	if update.Entries != nil {
		directories = update.Entries
	}
	for _, entry := range directories {
		if entry.IsDir() {
			if _, known := m.collapsed[entry.Path]; !known {
				m.collapsed[entry.Path] = true
			}
		}
	}
	for _, change := range update.Changes {
		if !structural && m.status[change.Path] == "" && change.Kind != "deleted" {
			if i, exists := m.entryPositions[change.Path]; exists && !m.entries[i].IsDir() {
				m.changedFiles++
			}
		}
		if change.Kind != "deleted" {
			m.activeUntil[change.Path] = time.Now().Add(2 * time.Second)
		} else {
			delete(m.activeUntil, change.Path)
		}
		switch change.Kind {
		case "created":
			m.status[change.Path] = "A"
		case "modified", "changed":
			m.status[change.Path] = "M"
		}
		for parent := filepath.Dir(change.Path); change.Path != "."; parent = filepath.Dir(parent) {
			if m.status[parent] == "" {
				m.status[parent] = "*"
			}
			if parent == "." {
				break
			}
		}
		if change.Path == m.preview.path || change.Path == "." ||
			strings.HasPrefix(m.preview.path, change.Path+string(filepath.Separator)) {
			m.previewDirty = true
		}
		if change.Kind == "deleted" {
			delete(m.collapsed, change.Path)
			delete(m.status, change.Path)
		}
	}
	if structural {
		m.changedFiles = 0
		for path := range m.status {
			if i, exists := m.entryPositions[path]; !exists {
				delete(m.status, path)
			} else if !m.entries[i].IsDir() {
				m.changedFiles++
			}
		}
	}
	if m.changeOffset > 0 {
		m.changeOffset += len(update.Changes)
	}
	incoming := update.Changes
	if len(incoming) >= historyLimit {
		m.changes = append(m.changes[:0], incoming[len(incoming)-historyLimit:]...)
	} else {
		if excess := len(m.changes) + len(incoming) - historyLimit; excess > 0 {
			copy(m.changes, m.changes[excess:])
			clear(m.changes[len(m.changes)-excess:])
			m.changes = m.changes[:len(m.changes)-excess]
		}
		m.changes = append(m.changes, incoming...)
	}
	if structural {
		if update.Entries != nil {
			m.rebuildRows()
		} else {
			m.updateHierarchy(update)
			m.flattenRows()
		}
	} else {
		m.changeOffset = min(m.changeOffset, max(0, len(m.changes)-viewport(m.watchPane)))
		return
	}
	for {
		if i, found := m.rowPositions[selectedPath]; found {
			m.selected = i
			break
		}
		if selectedPath == "." {
			break
		}
		selectedPath = filepath.Dir(selectedPath)
	}
	m.selected = min(m.selected, max(0, len(m.rows)-1))
	m.treeOffset = min(m.treeOffset, max(0, len(m.rows)*treeRowHeight-viewport(m.tree)))
	m.changeOffset = min(m.changeOffset, max(0, len(m.changes)-viewport(m.watchPane)))
}
