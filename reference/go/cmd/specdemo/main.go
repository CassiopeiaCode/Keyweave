package main

import (
	"fmt"

	"example.local/ptpg-spec-reference/internal/domain"
	"example.local/ptpg-spec-reference/internal/scheduler"
)

func main() {
	s := scheduler.NewAvailabilityRoundRobin()

	req := domain.RoutingRequest{
		RequestID:      "r1",
		GroupID:        "g1",
		Protocol:       "openai.chat",
		RequestedModel: "fast",
	}

	base := []domain.CandidateSnapshot{
		{
			InstanceID: "a",
			Availability: 100,
			Decrease: 10,
		},
		{
			InstanceID: "b",
			Availability: 100,
			Decrease: 10,
		},
	}

	candidates := []domain.CandidateRuntime{
		scheduler.RuntimeCandidate(base[0], 1),
		scheduler.RuntimeCandidate(base[1], 0),
	}

	got, err := s.Select(req, candidates)
	if err != nil {
		panic(err)
	}

	fmt.Println(got.InstanceID) // b
}
