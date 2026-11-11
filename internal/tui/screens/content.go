package screens

import (
	"fmt"
	"github.com/grindlemire/go-tui"
	"strconv"
	"strings"
)

func (m *Main) loadPreview() {
	if len(m.rows) == 0 {
		m.preview = filePreview{message: "Select a file in the context tree."}
		return
	}
	entry := m.rows[m.selected].entry
	if entry.Path == m.preview.path && !m.previewDirty {
		return
	}
	if entry.Path != m.preview.path {
		m.contentOffset = 0
		m.contentX = 0
	}
	m.previewDirty = false
	previous := m.preview
	m.preview = filePreview{path: entry.Path}
	if entry.IsDir() {
		m.preview.message = "Select a file in the context tree to view its content."
		return
	}
	content, err := readPreview(m.root, entry.Path)
	if err != nil {
		m.preview.message = fmt.Sprintf("Cannot preview %s: %s", display(entry.Path), display(err.Error()))
		return
	}
	code := strings.ReplaceAll(string(content), "\r\n", "\n")
	m.preview.code = code
	if previous.path == entry.Path && previous.message == "" {
		m.preview.cursor = locateChange(previous.code, code)
		if m.preview.cursor == nil {
			m.preview.cursor = previous.cursor
		}
	}
	m.preview.lines, m.preview.language, err = highlightWithStyle(entry.Path, code, m.theme.Syntax)
	if err != nil {
		m.preview.message = display(err.Error())
	}
	m.preview.gutter = max(4, len(strconv.Itoa(len(m.preview.lines))))
	for _, line := range m.preview.lines {
		width := m.preview.gutter + 2
		for _, span := range line {
			width += tui.StringWidth(span.Text)
		}
		width++
		m.preview.width = max(m.preview.width, width)
	}
	m.contentOffset = min(m.contentOffset, max(0, len(m.preview.lines)-viewport(m.feed)))
	m.scrollContentX(0)
	if cursor := m.preview.cursor; cursor != nil && previous.code != code && err == nil {
		m.followCursor()
	}
}

func (m *Main) followCursor() {
	cursor := m.preview.cursor
	if cursor == nil || m.preview.message != "" {
		return
	}
	height := viewport(m.feed)
	m.contentOffset = max(0, min(cursor.line-height/2, len(m.preview.lines)-height))
	line := strings.Split(m.preview.code, "\n")[cursor.line]
	visible := []rune(safeCode(line))
	column := m.preview.gutter + 2 + tui.StringWidth(string(visible[:min(cursor.column, len(visible))]))
	width := 80
	if m.feed != nil {
		width, _ = m.feed.ViewportSize()
	}
	m.contentX = max(0, column-max(1, width)+3)
}

func (m *Main) cursorSpans(line []tui.TextSpan, column int) []tui.TextSpan {
	var out []tui.TextSpan
	remaining := column
	cursorStyle := tui.NewStyle().Foreground(m.theme.Background).Background(m.theme.Activity)
	for _, span := range line {
		runes := []rune(span.Text)
		if remaining < 0 || remaining >= len(runes) {
			out = append(out, span)
			if remaining >= 0 {
				remaining -= len(runes)
			}
			continue
		}
		if remaining > 0 {
			out = append(out, tui.TextSpan{Text: string(runes[:remaining]), Style: span.Style})
		}
		out = append(out, tui.TextSpan{Text: string(runes[remaining : remaining+1]), Style: cursorStyle})
		if remaining+1 < len(runes) {
			out = append(out, tui.TextSpan{Text: string(runes[remaining+1:]), Style: span.Style})
		}
		remaining = -1
	}
	if remaining >= 0 {
		out = append(out, tui.TextSpan{Text: " ", Style: cursorStyle})
	}
	return out
}

func (m *Main) renderPreview() {
	if m.preview.message != "" {
		m.feed.AddChild(m.label(m.preview.message))
		return
	}
	if len(m.preview.lines) == 1 && len(m.preview.lines[0]) == 0 && m.preview.cursor == nil {
		m.feed.AddChild(m.label("(empty file)"))
		return
	}
	start := min(m.contentOffset, len(m.preview.lines))
	end := min(len(m.preview.lines), start+max(100, m.renderHeight))
	spacer := func(height int) {
		if height > 0 {
			m.feed.AddChild(tui.New(tui.WithHeight(height), tui.WithWidth(m.preview.width), tui.WithFlexShrink(0)))
		}
	}
	spacer(start)
	for i := start; i < end; i++ {
		line := m.preview.lines[i]
		active := m.preview.cursor != nil && m.preview.cursor.line == i
		gutter := fmt.Sprintf("%*d  ", m.preview.gutter, i+1)
		if active {
			gutter = fmt.Sprintf("%*d> ", m.preview.gutter, i+1)
			line = m.cursorSpans(line, m.preview.cursor.column)
		}
		spans := []tui.TextSpan{{
			Text: gutter, Style: tui.NewStyle().Foreground(m.theme.Muted),
		}}
		spans = append(spans, line...)
		width := 0
		for _, span := range spans {
			width += tui.StringWidth(span.Text)
		}
		el := tui.New(tui.WithRichText(spans...), tui.WithHeight(1),
			tui.WithWidth(width), tui.WithFlexShrink(0), tui.WithWrap(false))
		if active {
			el.Apply(tui.WithBackground(tui.Style{Bg: m.theme.ActiveSelection}))
		}
		m.feed.AddChild(el)
	}
	spacer(len(m.preview.lines) - end)
}

func (m *Main) scrollContentX(delta int) {
	width := 80
	if m.feed != nil {
		width, _ = m.feed.ViewportSize()
	}
	m.contentX = max(0, min(m.contentX+delta, m.preview.width-max(1, width)))
}
