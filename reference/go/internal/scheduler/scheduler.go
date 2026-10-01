package scheduler

import (
	"errors"
	"sort"
	"sync"

	"example.local/ptpg-spec-reference/internal/domain"
)

var ErrNoCandidate = errors.New("no eligible credential")

type AvailabilityRoundRobin struct {
	mu      sync.Mutex
	cursors map[string]int
}

func NewAvailabilityRoundRobin() *AvailabilityRoundRobin {
	return &AvailabilityRoundRobin{
		cursors: map[string]int{},
	}
}

func (s *AvailabilityRoundRobin) Select(
	req domain.RoutingRequest,
	candidates []domain.CandidateRuntime,
) (domain.CandidateRuntime, error) {
	eligible := make([]domain.CandidateRuntime, 0, len(candidates))
	for _, c := range candidates {
		if c.HardMaxConcurrency != nil &&
			c.ActiveRequests >= *c.HardMaxConcurrency {
			continue
		}
		eligible = append(eligible, c)
	}

	if len(eligible) == 0 {
		return domain.CandidateRuntime{}, ErrNoCandidate
	}

	max := eligible[0].EffectiveAvailability
	for _, c := range eligible[1:] {
		if c.EffectiveAvailability > max {
			max = c.EffectiveAvailability
		}
	}

	top := make([]domain.CandidateRuntime, 0, len(eligible))
	for _, c := range eligible {
		if c.EffectiveAvailability == max {
			top = append(top, c)
		}
	}

	sort.Slice(top, func(i, j int) bool {
		return top[i].InstanceID < top[j].InstanceID
	})

	bucket := req.GroupID + "|" +
		string(req.Protocol) + "|" +
		req.RequestedModel

	s.mu.Lock()
	defer s.mu.Unlock()

	i := s.cursors[bucket] % len(top)
	out := top[i]
	s.cursors[bucket] = (i + 1) % len(top)
	return out, nil
}

func RuntimeCandidate(
	c domain.CandidateSnapshot,
	active int,
) domain.CandidateRuntime {
	return domain.CandidateRuntime{
		CandidateSnapshot: c,
		ActiveRequests:    active,
		EffectiveAvailability: c.Availability -
			float64(active)*c.Decrease,
	}
}
