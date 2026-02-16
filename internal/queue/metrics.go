package queue

import "sync/atomic"

type Snapshot struct {
	Started   int64 `json:"started"`
	Completed int64 `json:"completed"`
	Failed    int64 `json:"failed"`
	InFlight  int64 `json:"in_flight"`
}

type Metrics struct {
	started   atomic.Int64
	completed atomic.Int64
	failed    atomic.Int64
	inFlight  atomic.Int64
}

func NewMetrics() *Metrics {
	return &Metrics{}
}

func (m *Metrics) MarkStart() {
	m.started.Add(1)
	m.inFlight.Add(1)
}

func (m *Metrics) MarkSuccess() {
	m.completed.Add(1)
	decrementSafe(&m.inFlight)
}

func (m *Metrics) MarkFailure() {
	m.failed.Add(1)
	decrementSafe(&m.inFlight)
}

func (m *Metrics) Snapshot() Snapshot {
	return Snapshot{
		Started:   m.started.Load(),
		Completed: m.completed.Load(),
		Failed:    m.failed.Load(),
		InFlight:  m.inFlight.Load(),
	}
}

func decrementSafe(value *atomic.Int64) {
	for {
		current := value.Load()
		if current <= 0 {
			return
		}
		if value.CompareAndSwap(current, current-1) {
			return
		}
	}
}
