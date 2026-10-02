package routing

import (
	"errors"

	"github.com/CassiopeiaCode/Keyweave/go/internal/domain"
)

var ErrNoCandidate = errors.New("no routable candidate")

func BuildCandidateSet(
	group domain.CredentialGroup,
	templates map[string]domain.CredentialTemplate,
	instances []domain.CredentialInstance,
	protocol domain.ProtocolID,
	model string,
) ([]domain.CandidateSnapshot, error) {
	allowedTemplates := make(map[string]struct{}, len(group.TemplateIDs))
	for _, id := range group.TemplateIDs {
		allowedTemplates[id] = struct{}{}
	}

	out := make([]domain.CandidateSnapshot, 0, len(instances))
	for _, inst := range instances {
		if _, ok := allowedTemplates[inst.TemplateID]; !ok {
			continue
		}

		tpl, ok := templates[inst.TemplateID]
		if !ok || !supportsProtocol(tpl.Protocols, protocol) {
			continue
		}

		upstreamModel, ok := matchModel(inst.Models, protocol, model)
		if !ok {
			continue
		}

		out = append(out, domain.CandidateSnapshot{
			InstanceID:    inst.ID,
			TemplateID:    inst.TemplateID,
			Availability:  inst.Availability,
			Decrease:      1,
			CanSchedule:   true,
			ClientModel:   model,
			UpstreamModel: upstreamModel,
			Attributes:    map[string]interface{}{},
		})
	}

	if len(out) == 0 {
		return nil, ErrNoCandidate
	}
	return out, nil
}

func supportsProtocol(protocols []domain.ProtocolID, requested domain.ProtocolID) bool {
	for _, p := range protocols {
		if p == requested {
			return true
		}
	}
	return false
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
