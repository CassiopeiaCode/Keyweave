package domain

type ProtocolID string

type ModelDefinition struct {
	ID                   string         `json:"id"`
	Upstream             string         `json:"upstream,omitempty"`
	DisplayName          string         `json:"displayName,omitempty"`
	Protocols            []ProtocolID   `json:"protocols,omitempty"`
	ForceResponseMapping bool           `json:"forceResponseMapping,omitempty"`
	Metadata             map[string]any `json:"metadata,omitempty"`
}

type CredentialInstance struct {
	ID           string
	TemplateID   string
	Availability float64
	Proxy        string
	Models       []ModelDefinition
	CustomFields map[string]any
	State        map[string]any
	Version      int64
}

type CredentialGroup struct {
	ID          string
	TemplateIDs []string
}

type CandidateSnapshot struct {
	InstanceID string
	TemplateID string

	Availability float64
	Decrease     float64

	ClientModel   string
	UpstreamModel string

	HardMaxConcurrency *int
	Attributes         map[string]any
}

type CandidateRuntime struct {
	CandidateSnapshot
	ActiveRequests        int
	EffectiveAvailability float64
}

type RoutingRequest struct {
	RequestID      string
	GroupID        string
	Protocol       ProtocolID
	RequestedModel string
}
