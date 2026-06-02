package mcp

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/manuelibar/workbench/internal/mcp/tools"
)

const (
	methodListTools     = "tools/list"
	methodListResources = "resources/list"
	methodListPrompts   = "prompts/list"
)

type syncTracker struct {
	generation int64
	required   map[string]bool
	observed   map[string]bool
	done       chan struct{}
}

type capabilitySync struct {
	mu         sync.Mutex
	generation int64
	current    *syncTracker
	timeout    time.Duration
}

func newCapabilitySync(timeout time.Duration) *capabilitySync {
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return &capabilitySync{timeout: timeout}
}

func (s *capabilitySync) SetTimeout(timeout time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.timeout = timeout
}

func (s *capabilitySync) Begin(categories []string) *syncTracker {
	methods := methodsForCategories(categories)
	if len(methods) == 0 {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current != nil {
		close(s.current.done)
	}
	s.generation++
	tracker := &syncTracker{
		generation: s.generation,
		required:   map[string]bool{},
		observed:   map[string]bool{},
		done:       make(chan struct{}),
	}
	for _, method := range methods {
		tracker.required[method] = true
	}
	s.current = tracker
	return tracker
}

func (s *capabilitySync) MarkObserved(method string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tracker := s.current
	if tracker == nil || !tracker.required[method] || tracker.observed[method] {
		return
	}
	tracker.observed[method] = true
	if len(tracker.observed) == len(tracker.required) {
		close(tracker.done)
		s.current = nil
	}
}

func (s *capabilitySync) Wait(ctx context.Context, tracker *syncTracker) tools.CapabilitySyncStatus {
	if tracker == nil {
		return tools.CapabilitySyncStatus{Status: "unchanged"}
	}
	timeout := s.getTimeout()
	if timeout < 0 {
		timeout = 0
	}
	if timeout == 0 {
		return s.status(tracker, true)
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-tracker.done:
		return s.status(tracker, false)
	case <-timer.C:
		return s.status(tracker, true)
	case <-ctx.Done():
		return s.status(tracker, true)
	}
}

func (s *capabilitySync) getTimeout() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.timeout
}

func (s *capabilitySync) status(tracker *syncTracker, timedOut bool) tools.CapabilitySyncStatus {
	s.mu.Lock()
	if timedOut && s.current == tracker {
		s.current = nil
	}
	required := make([]string, 0, len(tracker.required))
	for method := range tracker.required {
		required = append(required, method)
	}
	observed := make([]string, 0, len(tracker.observed))
	for method := range tracker.observed {
		observed = append(observed, method)
	}
	s.mu.Unlock()
	sort.Strings(required)
	sort.Strings(observed)
	status := "synced"
	if timedOut {
		status = "timeout_fallback"
	}
	return tools.CapabilitySyncStatus{
		Generation: tracker.generation,
		Status:     status,
		Required:   required,
		Observed:   observed,
		TimedOut:   timedOut,
	}
}

func methodsForCategories(categories []string) []string {
	seen := map[string]bool{}
	for _, category := range categories {
		switch category {
		case "tools":
			seen[methodListTools] = true
		case "resources":
			seen[methodListResources] = true
		case "prompts":
			seen[methodListPrompts] = true
		}
	}
	methods := make([]string, 0, len(seen))
	for method := range seen {
		methods = append(methods, method)
	}
	sort.Strings(methods)
	return methods
}
