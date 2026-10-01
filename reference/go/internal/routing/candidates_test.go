package routing

import (
	"testing"

	"example.local/ptpg-spec-reference/internal/domain"
)

func TestBuildCandidateSetEnforcesGroupScopeAndModel(t *testing.T) {
	group := domain.CredentialGroup{ID: "g", TemplateIDs: []string{"t1"}}
	instances := []domain.CredentialInstance{
		{ID: "a", TemplateID: "t1", Availability: 100, Models: []domain.ModelDefinition{{ID: "client", Upstream: "up-a", Protocols: []domain.ProtocolID{"openai.chat"}}}},
		{ID: "b", TemplateID: "t2", Availability: 100, Models: []domain.ModelDefinition{{ID: "client", Upstream: "up-b"}}},
		{ID: "c", TemplateID: "t1", Availability: 100, Models: []domain.ModelDefinition{{ID: "other"}}},
	}
	got, err := BuildCandidateSet(group, instances, domain.ProtocolID("openai.chat"), "client")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].InstanceID != "a" || got[0].UpstreamModel != "up-a" {
		t.Fatalf("unexpected candidates: %#v", got)
	}
}
