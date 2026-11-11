package screens

import (
	"fmt"
	"github.com/MoeidHeidari/guardrail/internal/tui/components"
	"github.com/grindlemire/go-tui"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func display(text string) string {
	quoted := strconv.Quote(text)
	return quoted[1 : len(quoted)-1]
}

func (m *Main) label(text string) *tui.Element {
	return tui.New(tui.WithText(text), tui.WithHeight(1), tui.WithFlexShrink(0),
		tui.WithWrap(false), tui.WithTruncate(true))
}

func (m *Main) AdvanceAnimation() bool {
	return m.advanceAnimation(time.Now())
}

func (m *Main) HasActivity() bool {
	return len(m.activeUntil) > 0
}

func (m *Main) advanceAnimation(now time.Time) bool {
	m.animationFrame = (m.animationFrame + 1) % 4
	changed := false
	for path, until := range m.activeUntil {
		el := m.animationRows[path]
		if !now.Before(until) {
			delete(m.activeUntil, path)
			if el != nil {
				el.SetText(m.animationIndent[path] + "    " + m.animationText[path])
				changed = true
			}
		} else if el != nil {
			el.SetText(m.animationIndent[path] + dotFrames[m.animationFrame] + m.animationText[path])
			changed = true
		}
	}
	m.animationOnly = changed
	return changed
}

var dotFrames = [...]string{".   ", "..  ", "... ", " .. "}

func (m *Main) Render(app *tui.App) *tui.Element {
	if m.animationOnly && m.cachedRoot != nil {
		m.animationOnly = false
		return m.cachedRoot
	}
	m.animationRows = make(map[string]*tui.Element)
	m.animationText = make(map[string]string)
	m.animationIndent = make(map[string]string)
	if app != nil {
		_, m.renderHeight = app.Size()
	}
	dragTree := m.dragPane != nil && m.dragPane == m.tree
	dragFeed := m.dragPane != nil && m.dragPane == m.feed
	header := components.NewHeader(m.theme, "GUARDRAIL", display(filepath.Base(m.root))+"  /  workspace")
	m.tree = components.NewContainer(m.theme, tui.WithFlexGrow(1),
		tui.WithHeight(0), tui.WithScrollable(tui.ScrollVertical),
		tui.WithScrollOffset(0, m.treeOffset))
	m.styleScrollbar(m.tree)
	start := min(m.treeOffset, len(m.rows))
	end := min(len(m.rows), start+max(100, m.renderHeight))
	if start > 0 {
		m.tree.AddChild(tui.New(tui.WithHeight(start), tui.WithFlexShrink(0)))
	}
	for i := start; i < end; i++ {
		row := m.rows[i]
		name := filepath.Base(row.entry.Path)
		if row.entry.Path == "." {
			name = filepath.Base(m.root)
		}
		prefix := "    "
		if row.entry.IsDir() {
			prefix = "[-] "
			if m.collapsed[row.entry.Path] {
				prefix = "[+] "
			}
			name += "/"
		}
		changing := !row.entry.IsDir() && m.status[row.entry.Path] != ""
		_, active := m.activeUntil[row.entry.Path]
		if changing && active {
			prefix = dotFrames[m.animationFrame]
		}
		el := m.label(strings.Repeat("  ", row.depth) + prefix + display(name))
		if status := m.status[row.entry.Path]; status != "" {
			el = m.label(strings.Repeat("  ", row.depth) + prefix + "[" + status + "] " + display(name))
			color := m.theme.Modified
			if status == "A" {
				color = m.theme.Added
			}
			if changing {
				color = m.theme.Activity
			}
			el.Apply(tui.WithTextStyle(tui.Style{Fg: color}))
		}
		if i == m.selected {
			background := m.theme.Selection
			foreground := m.theme.Foreground
			if changing {
				background = m.theme.ActiveSelection
				foreground = m.theme.Activity
			}
			if m.changesFocus {
				background = m.theme.Chrome
				foreground = m.theme.Accent
				if changing {
					foreground = m.theme.Activity
				}
			}
			el.Apply(tui.WithTextStyle(tui.Style{Fg: foreground, Bg: background}),
				tui.WithBackground(tui.Style{Bg: background}))
		}
		m.tree.AddChild(el)
		if changing && active {
			m.animationRows[row.entry.Path] = el
			m.animationText[row.entry.Path] = strings.TrimPrefix(el.Text(), strings.Repeat("  ", row.depth)+prefix)
			m.animationIndent[row.entry.Path] = strings.Repeat("  ", row.depth)
		}
	}
	if end < len(m.rows) {
		m.tree.AddChild(tui.New(tui.WithHeight(len(m.rows)-end), tui.WithFlexShrink(0)))
	}
	treeTitle := "Context"
	treeHeading := m.label(treeTitle)
	treeHeading.Apply(tui.WithTextStyle(tui.NewStyle().Foreground(m.theme.Muted).Bold()))
	if !m.changesFocus {
		treeHeading.Apply(tui.WithTextStyle(tui.NewStyle().Foreground(m.theme.Accent).Bold()))
	}
	sidebar := components.NewSection(m.theme, "", treeHeading, m.tree)
	sidebar.Apply(tui.WithPaddingTRBL(1, 2, 1, 2))
	content := m.renderPanel()
	if dragTree {
		m.dragPane = m.tree
	} else if dragFeed {
		m.dragPane = m.feed
	}
	focus := "Context"
	if m.changesFocus {
		focus = "Watch"
		if m.contentTab {
			focus = "Content"
		} else if m.dashboardVisible() {
			focus = "Dashboard"
		}
	}
	status := m.label(fmt.Sprintf("Watching   %d files & folders   %d changed   /   %s", len(m.entries), m.changedFiles, focus))
	status.Apply(tui.WithMinWidth(0), tui.WithFlexShrink(1))
	status.Apply(tui.WithBackground(tui.Style{Bg: m.theme.Chrome}),
		tui.WithTextStyle(tui.Style{Fg: m.theme.Muted}))
	hint := "0/1/2 tabs   Tab focus   Enter open   q quit"
	if app != nil {
		width, _ := app.Size()
		if width < 90 {
			hint = "Tab focus   q quit"
		}
		if width < 55 {
			hint = "q quit"
		}
	}
	help := m.label(hint)
	help.Apply(tui.WithTextStyle(tui.Style{Fg: m.theme.Muted}))
	help.Apply(tui.WithFlexShrink(0))
	footer := components.NewContainer(m.theme, tui.WithHeight(3), tui.WithFlexShrink(0),
		tui.WithDirection(tui.Row), tui.WithJustify(tui.JustifySpaceBetween),
		tui.WithPaddingTRBL(1, 2, 1, 2),
		tui.WithGap(2),
		tui.WithBackground(tui.Style{Bg: m.theme.Chrome}))
	footer.AddChild(status, help)
	m.cachedRoot = components.NewLayout(m.theme, header, sidebar, content, footer)
	return m.cachedRoot
}
