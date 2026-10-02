package routing

import (
	"errors"
	"testing"

	"github.com/CassiopeiaCode/Keyweave/go/internal/domain"
)

func TestBuildCandidateSetEnforcesGroupScopeProtocolAndModel(t *testing.T) {
	group := domain.CredentialGroup{
		ID:          "g-1",
		TemplateIDs: []string{"t-1"},
	}
	templates := map[string]domain.CredentialTemplate{
		"t-1": {ID: "t-1", Protocols: []domain.ProtocolID{"openai.chat"}},
		"t-2": {ID: "t-2", Protocols: []domain.ProtocolID{"openai.chat"}},
	}
	instances := []domain.CredentialInstance{
		{
			ID:         "i-1",
			TemplateID: "t-1",
			CredentialOfficialFields: domain.CredentialOfficialFields{
				Availability: 100,
				Models: []domain.ModelDefinition{{
					ID:       "fast",
					Upstream: "provider/model-a",
					Protocols: []domain.ProtocolID{
						"openai.chat",
					},
				}},
			},
		},
		{
			ID:         "i-2",
			TemplateID: "t-2",
			CredentialOfficialFields: domain.CredentialOfficialFields{
				Availability: 100,
				Models:       []domain.ModelDefinition{{ID: "fast", Upstream: "provider/model-b"}},
			},
		},
		{
			ID:         "i-3",
			TemplateID: "t-1",
			CredentialOfficialFields: domain.CredentialOfficialFields{
				Availability: 100,
				Models:       []domain.ModelDefinition{{ID: "slow", Upstream: "provider/model-slow"}},
			},
		},
	}

	got, err := BuildCandidateSet(group, templates, instances, domain.ProtocolID("openai.chat"), "fast")
	if err != nil {
		t.Fatalf("build candidates: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("unexpected candidate count: %d", len(got))
	}
	if got[0].InstanceID != "i-1" {
		t.Fatalf("unexpected selected candidate: %#v", got[0])
	}
	if got[0].UpstreamModel != "provider/model-a" {
		t.Fatalf("unexpected upstream model: %s", got[0].UpstreamModel)
	}
}

func TestBuildCandidateSetReturnsErrNoCandidate(t *testing.T) {
	group := domain.CredentialGroup{
		ID:          "g-1",
		TemplateIDs: []string{"t-1"},
	}
	templates := map[string]domain.CredentialTemplate{
		"t-1": {ID: "t-1", Protocols: []domain.ProtocolID{"anthropic.messages"}},
	}
	instances := []domain.CredentialInstance{
		{
			ID:         "i-1",
			TemplateID: "t-1",
			CredentialOfficialFields: domain.CredentialOfficialFields{
				Availability: 100,
				Models:       []domain.ModelDefinition{{ID: "fast"}},
			},
		},
	}

	_, err := BuildCandidateSet(group, templates, instances, domain.ProtocolID("openai.chat"), "fast")
	if !errors.Is(err, ErrNoCandidate) {
		t.Fatalf("expected ErrNoCandidate, got %v", err)
	}
}
