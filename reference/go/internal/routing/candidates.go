package routing

import (
	"fmt"

	"example.local/ptpg-spec-reference/internal/domain"
)

// BuildCandidateSet demonstrates the static part of candidate resolution.
// Production code additionally invokes Template hooks outside the scheduling
// critical section and replaces the default decrease with Template output.
func BuildCandidateSet(
	group domain.CredentialGroup,
	instances []domain.CredentialInstance,
	protocol domain.ProtocolID,
	model string,
) ([]domain.CandidateSnapshot, error) {
	allowed := make(map[string]struct{}, len(group.TemplateIDs))
	for _, id := range group.TemplateIDs {
		allowed[id] = struct{}{}
	}

	out := make([]domain.CandidateSnapshot, 0, len(instances))
	for _, inst := range instances {
		if _, ok := allowed[inst.TemplateID]; !ok {
			continue
		}

		upstream, ok := matchModel(inst.Models, protocol, model)
		if !ok {
			continue
		}

		out = append(out, domain.CandidateSnapshot{
			InstanceID:    inst.ID,
			TemplateID:    inst.TemplateID,
			Availability:  inst.Availability,
			Decrease:      1,
			ClientModel:   model,
			UpstreamModel: upstream,
		})
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("no candidate for protocol=%s model=%s", protocol, model)
	}
	return out, nil
}

func matchModel(models []domain.ModelDefinition, protocol domain.ProtocolID, requested string) (string, bool) {
	for _, m := range models {
		if m.ID != requested {
			continue
		}
		if len(m.Protocols) > 0 {
			matched := false
			for _, p := range m.Protocols {
				if p == protocol {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		if m.Upstream != "" {
			return m.Upstream, true
		}
		return m.ID, true
	}
	return "", false
}
