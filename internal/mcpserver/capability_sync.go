package mcpserver

import (
	"context"
	"sort"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const defaultRefreshListWait = 3 * time.Second

type capabilitySyncTracker struct {
	generation int64
	required   map[string]bool
	observed   map[string]bool
	done       chan struct{}
}

func newCapabilitySyncTracker(generation int64, required []string) *capabilitySyncTracker {
	req := make(map[string]bool, len(required))
	for _, method := range required {
		req[method] = true
	}
	return &capabilitySyncTracker{generation: generation, required: req, observed: map[string]bool{}, done: make(chan struct{})}
}

func (s *Server) SetRefreshListWait(d time.Duration) {
	s.syncMu.Lock()
	s.refreshListWait = d
	s.syncMu.Unlock()
}

func (s *Server) installCapabilityListMiddleware() {
	s.sdkServer.AddReceivingMiddleware(func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			result, err := next(ctx, method, req)
			if err == nil {
				s.markCapabilityListObserved(method)
			}
			return result, err
		}
	})
}

func (s *Server) startCapabilitySync(required []string) *capabilitySyncTracker {
	s.syncMu.Lock()
	defer s.syncMu.Unlock()
	s.capabilityGeneration++
	tracker := newCapabilitySyncTracker(s.capabilityGeneration, required)
	s.capabilitySync = tracker
	return tracker
}

func (s *Server) markCapabilityListObserved(method string) {
	s.syncMu.Lock()
	defer s.syncMu.Unlock()
	tracker := s.capabilitySync
	if tracker == nil || !tracker.required[method] || tracker.observed[method] {
		return
	}
	tracker.observed[method] = true
	if len(tracker.observed) == len(tracker.required) {
		close(tracker.done)
		s.capabilitySync = nil
	}
}

func (s *Server) waitForCapabilityRelist(ctx context.Context, tracker *capabilitySyncTracker, index CapabilityIndexWire) CapabilitySyncWire {
	wait := s.getRefreshListWait()
	if wait <= 0 {
		return s.capabilitySyncWire(tracker, true, index)
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-tracker.done:
		return s.capabilitySyncWire(tracker, false, CapabilityIndexWire{})
	case <-timer.C:
		return s.capabilitySyncWire(tracker, true, index)
	case <-ctx.Done():
		return s.capabilitySyncWire(tracker, true, index)
	}
}

func (s *Server) getRefreshListWait() time.Duration {
	s.syncMu.Lock()
	defer s.syncMu.Unlock()
	return s.refreshListWait
}

func (s *Server) capabilitySyncWire(tracker *capabilitySyncTracker, timedOut bool, index CapabilityIndexWire) CapabilitySyncWire {
	s.syncMu.Lock()
	observed := make([]string, 0, len(tracker.observed))
	for method := range tracker.observed {
		observed = append(observed, method)
	}
	required := make([]string, 0, len(tracker.required))
	for method := range tracker.required {
		required = append(required, method)
	}
	if timedOut && s.capabilitySync == tracker {
		s.capabilitySync = nil
	}
	s.syncMu.Unlock()
	sort.Strings(observed)
	sort.Strings(required)
	status := "client_relisted"
	var indexPtr *CapabilityIndexWire
	if timedOut {
		status = "fallback_index"
		indexPtr = &index
	}
	return CapabilitySyncWire{Generation: tracker.generation, Status: status, Required: required, Observed: observed, TimedOut: timedOut, Index: indexPtr}
}
