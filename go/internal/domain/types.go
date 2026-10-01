package domain

import "time"

type ProtocolID string

type ModelDefinition struct {
	ID                   string                 `json:"id"`
	Upstream             string                 `json:"upstream,omitempty"`
	DisplayName          string                 `json:"displayName,omitempty"`
	Protocols            []ProtocolID           `json:"protocols,omitempty"`
	ForceResponseMapping bool                   `json:"forceResponseMapping,omitempty"`
	InputModalities      []string               `json:"inputModalities,omitempty"`
	OutputModalities     []string               `json:"outputModalities,omitempty"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
}

type CredentialOfficialFields struct {
	Address      *string           `json:"address,omitempty"`
	BaseURL      *string           `json:"baseUrl,omitempty"`
	Models       []ModelDefinition `json:"models,omitempty"`
	Availability float64           `json:"availability"`
	Proxy        *string           `json:"proxy,omitempty"`
}

type CredentialTemplate struct {
	ID               string
	Name             string
	Description      *string
	Protocols        []ProtocolID
	OfficialDefaults map[string]interface{}
	CustomFields     map[string]interface{}
	State            map[string]interface{}
	JSSource         string
	Version          int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type CredentialInstance struct {
	CredentialOfficialFields
	ID            string
	TemplateID    string
	Name          *string
	KeyCiphertext []byte
	KeyNonce      []byte
	KeyKID        *string
	CustomFields  map[string]interface{}
	State         map[string]interface{}
	Version       int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CredentialGroup struct {
	ID            string
	Name          string
	TemplateIDs   []string
	SchedulerType string
	SchedulerCode *string
	Config        map[string]interface{}
	State         map[string]interface{}
	Version       int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type AccessKey struct {
	ID           string
	Name         string
	SecretHash   []byte
	SecretPrefix *string
	GroupID      string
	CreatedAt    time.Time
	RevokedAt    *time.Time
}

type CredentialSource struct {
	ID         string
	Name       string
	Type       string
	TemplateID *string
	Config     map[string]interface{}
	Code       *string
	State      map[string]interface{}
	Version    int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type SecretItem struct {
	ID         string
	OwnerType  string
	OwnerID    string
	Name       string
	Ciphertext []byte
	Nonce      []byte
	KeyID      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
