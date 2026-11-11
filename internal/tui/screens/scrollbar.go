package screens

import "github.com/grindlemire/go-tui"

func (m *Main) styleScrollbar(el *tui.Element) {
	el.Apply(
		tui.WithScrollbarStyle(tui.Style{Fg: m.theme.Separator, Bg: m.theme.Background}),
		tui.WithScrollbarThumbStyle(tui.Style{Fg: m.theme.Accent, Bg: m.theme.Background}),
	)
}

func scrollThumb(el *tui.Element) (top, size, travel, maximum int) {
	height := viewport(el)
	_, content := el.ContentSize()
	_, maximum = el.MaxScroll()
	size = min(height, max(1, height*height/max(1, content)))
	travel = height - size
	_, offset := el.ScrollOffset()
	if maximum > 0 {
		top = offset * travel / maximum
	}
	return
}

func (m *Main) handleScrollbar(event tui.MouseEvent) bool {
	if m.dragPane != nil {
		if event.Action == tui.MouseRelease {
			m.dragPane = nil
			return true
		}
		if event.Action == tui.MouseDrag {
			el := m.dragPane
			_, _, travel, maximum := scrollThumb(el)
			position := max(0, min(travel, event.Y-el.ContentRect().Y-m.dragGrab))
			offset := 0
			if travel > 0 {
				offset = (position*maximum + travel/2) / travel
			}
			m.setScroll(el, offset)
			return true
		}
	}
	if event.Action != tui.MousePress || event.Button != tui.MouseLeft {
		return false
	}
	for _, el := range []*tui.Element{m.tree, m.feed} {
		if el == nil {
			continue
		}
		rect := el.ContentRect()
		top, size, _, maximum := scrollThumb(el)
		if maximum == 0 || event.X != rect.X+rect.Width-1 || !rect.Contains(event.X, event.Y) {
			continue
		}
		m.changesFocus = el == m.feed
		y := event.Y - rect.Y
		if y >= top && y < top+size {
			m.dragPane, m.dragGrab = el, y-top
		} else {
			_, offset := el.ScrollOffset()
			delta := viewport(el)
			if y < top {
				delta = -delta
			}
			m.setScroll(el, max(0, min(maximum, offset+delta)))
		}
		return true
	}
	return false
}

func (m *Main) setScroll(el *tui.Element, offset int) {
	if el == m.tree {
		m.treeOffset = offset
	} else if m.contentTab {
		m.contentOffset = offset
	} else {
		m.changeOffset = offset
	}
	// Render replaces elements, so continue dragging against the current pane.
	if m.dragPane != nil {
		m.dragPane = el
	}
}
