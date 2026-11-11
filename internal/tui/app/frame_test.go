package app

import (
	"testing"
	"time"
)

func TestFrameQueueCoalescesWithoutPostponing(t *testing.T) {
	q := newFrameQueue()
	defer q.timer.Stop()
	if q.tick != nil {
		t.Fatal("idle queue must have no active tick")
	}
	q.request(true)
	tick := q.tick
	for range 10000 {
		q.request(false)
	}
	if q.tick != tick || !q.full {
		t.Fatal("animation requests must not replace a pending full frame")
	}
	select {
	case <-q.tick:
	case <-time.After(time.Second):
		t.Fatal("continuous events must not starve a scheduled frame")
	}
	if !q.take() || q.tick != nil || q.full {
		t.Fatal("render must consume the pending full frame and return to idle")
	}
	q.request(false)
	if q.take() {
		t.Fatal("animation-only frames must remain lightweight")
	}
}
