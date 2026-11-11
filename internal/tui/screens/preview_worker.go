package screens

import (
	"context"
	"time"

	"github.com/MoeidHeidari/guardrail/internal/tui/styles"
	"github.com/MoeidHeidari/guardrail/internal/watcher"
)

type previewRequest struct {
	theme    styles.Theme
	id       uint64
	root     string
	entry    watcher.Entry
	previous filePreview
	x, y     int
}

type PreviewResult struct {
	id      uint64
	preview filePreview
}

// StartPreviewWorker bounds work to one running and one latest pending preview.
// Call stop before disposing the screen; result application belongs to the UI loop.
func (m *Main) StartPreviewWorker(parent context.Context) (<-chan PreviewResult, func()) {
	ctx, cancel := context.WithCancel(parent)
	m.previewRequests = make(chan previewRequest, 1)
	results := make(chan PreviewResult, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				return
			case request := <-m.previewRequests:
				// Briefly coalesce rapid saves before starting expensive lexing.
				timer := time.NewTimer(40 * time.Millisecond)
				select {
				case <-ctx.Done():
					timer.Stop()
					return
				case <-timer.C:
				}
				select {
				case latest := <-m.previewRequests:
					request = latest
				default:
				}
				worker := &Main{
					theme: request.theme,
					root:  request.root, rows: []treeRow{{entry: request.entry}},
					preview: request.previous, previewDirty: true,
					contentX: request.x, contentOffset: request.y,
				}
				worker.loadPreview()
				result := PreviewResult{id: request.id, preview: worker.preview}
				select {
				case <-results:
				default:
				}
				select {
				case results <- result:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return results, func() { cancel(); <-done }
}

func (m *Main) requestPreview() {
	if len(m.rows) == 0 {
		return
	}
	entry := m.rows[m.selected].entry
	if entry.Path == m.requestedPath && !m.previewDirty {
		return
	}
	m.previewID++
	request := previewRequest{
		theme: m.theme,
		id:    m.previewID, root: m.root, entry: entry, previous: m.preview,
		x: m.contentX, y: m.contentOffset,
	}
	m.requestedPath, m.previewDirty = entry.Path, false
	if entry.Path != m.preview.path {
		m.preview = filePreview{path: entry.Path, message: "Loading preview..."}
		m.contentX, m.contentOffset = 0, 0
	}
	select {
	case <-m.previewRequests:
	default:
	}
	m.previewRequests <- request
}

func (m *Main) ApplyPreview(result PreviewResult) bool {
	if result.id != m.previewID || m.previewDirty || len(m.rows) == 0 ||
		m.rows[m.selected].entry.Path != result.preview.path {
		return false
	}
	m.animationOnly = false
	changed := m.preview.code != result.preview.code
	m.preview = result.preview
	m.contentOffset = min(m.contentOffset, max(0, len(m.preview.lines)-viewport(m.feed)))
	m.scrollContentX(0)
	if changed {
		m.followCursor()
	}
	return true
}
