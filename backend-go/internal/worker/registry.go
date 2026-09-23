package worker

import (
	"sync"
	"time"
)

type WorkerState struct {
	Name           string    `json:"name"`
	Category       string    `json:"category"`
	Interval       string    `json:"interval"`
	Status         string    `json:"status"` // "RUNNING", "IDLE", "ERROR"
	LastRun        time.Time `json:"last_run"`
	RunCount       int64     `json:"run_count"`
	ProcessedItems int64     `json:"processed_items"`
	LastError      string    `json:"last_error,omitempty"`
	Description    string    `json:"description"`
}

type WorkerRegistry struct {
	mu      sync.RWMutex
	workers map[string]*WorkerState
}

var GlobalRegistry = &WorkerRegistry{
	workers: make(map[string]*WorkerState),
}

func (r *WorkerRegistry) RegisterWorker(name, category, interval, description string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.workers[name] = &WorkerState{
		Name:        name,
		Category:    category,
		Interval:    interval,
		Status:      "RUNNING",
		LastRun:     time.Now().UTC(),
		Description: description,
	}
}

func (r *WorkerRegistry) RecordRun(name string, processed int64, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	state, exists := r.workers[name]
	if !exists {
		state = &WorkerState{
			Name:   name,
			Status: "RUNNING",
		}
		r.workers[name] = state
	}

	state.LastRun = time.Now().UTC()
	state.RunCount++
	state.ProcessedItems += processed
	if err != nil {
		state.Status = "ERROR"
		state.LastError = err.Error()
	} else {
		state.Status = "RUNNING"
		state.LastError = ""
	}
}

func (r *WorkerRegistry) GetSnapshots() []WorkerState {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]WorkerState, 0, len(r.workers))
	for _, w := range r.workers {
		list = append(list, *w)
	}
	return list
}
