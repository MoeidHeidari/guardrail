package app

import "time"

const frameInterval = 16 * time.Millisecond

type frameQueue struct {
	timer *time.Timer
	tick  <-chan time.Time
	full  bool
}

func newFrameQueue() *frameQueue {
	timer := time.NewTimer(time.Hour)
	timer.Stop()
	return &frameQueue{timer: timer}
}

func (q *frameQueue) request(full bool) {
	q.full = q.full || full
	if q.tick == nil {
		q.timer.Reset(frameInterval)
		q.tick = q.timer.C
	}
}

func (q *frameQueue) take() bool {
	full := q.full
	q.tick, q.full = nil, false
	return full
}
