package scheduler

import (
	"testing"

	"example.local/ptpg-spec-reference/internal/domain"
)

func TestDecreaseSteersToIdleCredential(t *testing.T) {
	s := NewAvailabilityRoundRobin()
	req := domain.RoutingRequest{
		RequestID:      "r1",
		GroupID:        "g1",
		Protocol:       "openai.chat",
		RequestedModel: "fast",
	}

	a := domain.CandidateSnapshot{
		InstanceID:   "a",
		Availability: 100,
		Decrease:     10,
	}
	b := domain.CandidateSnapshot{
		InstanceID:   "b",
		Availability: 100,
		Decrease:     10,
	}

	got, err := s.Select(req, []domain.CandidateRuntime{
		RuntimeCandidate(a, 1),
		RuntimeCandidate(b, 0),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.InstanceID != "b" {
		t.Fatalf("expected b, got %s", got.InstanceID)
	}
}

func TestHardConcurrency(t *testing.T) {
	s := NewAvailabilityRoundRobin()
	one := 1

	req := domain.RoutingRequest{
		RequestID:      "r2",
		GroupID:        "g1",
		Protocol:       "openai.responses",
		RequestedModel: "x",
	}

	a := domain.CandidateSnapshot{
		InstanceID:         "a",
		Availability:       1000,
		HardMaxConcurrency: &one,
	}
	b := domain.CandidateSnapshot{
		InstanceID:   "b",
		Availability: 100,
	}

	got, err := s.Select(req, []domain.CandidateRuntime{
		RuntimeCandidate(a, 1),
		RuntimeCandidate(b, 0),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.InstanceID != "b" {
		t.Fatalf("expected b, got %s", got.InstanceID)
	}
}
