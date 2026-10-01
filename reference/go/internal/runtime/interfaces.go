package runtime

import (
	"context"
	"io"
	"net/http"
)

type RequestContext struct {
	RequestID      string
	GroupID        string
	ClientKeyID    string
	Protocol       string
	RequestedModel string
	UpstreamModel  string
	Attempt        int
}

type Secret interface {
	Reveal() string
}

type Secrets interface {
	Get(
		ctx context.Context,
		name string,
	) (Secret, bool, error)
}

type HostHTTP interface {
	Do(
		ctx context.Context,
		req *http.Request,
	) (*http.Response, error)
}

type InstanceMutator interface {
	PatchAvailability(
		ctx context.Context,
		value float64,
	) error

	PatchCustomFields(
		ctx context.Context,
		mergePatch map[string]any,
	) error

	AtomicIncrement(
		ctx context.Context,
		path string,
		delta float64,
	) (float64, error)
}

type TemplateContext struct {
	Request  RequestContext
	Fields   map[string]any
	Secrets  Secrets
	HTTP     HostHTTP
	Instance InstanceMutator
}

// TemplateProgram is a host-side facade. A JS runtime adapts compiled
// JS exports to this interface.
type TemplateProgram interface {
	Supports(
		ctx context.Context,
		in TemplateContext,
	) (bool, error)

	GetAvailability(
		ctx context.Context,
		in TemplateContext,
	) (float64, error)

	GetDecrease(
		ctx context.Context,
		in TemplateContext,
	) (float64, error)

	Execute(
		ctx context.Context,
		in TemplateContext,
		body io.Reader,
	) (*http.Response, error)

	OnError(
		ctx context.Context,
		in TemplateContext,
		cause error,
	) (ErrorAction, error)
}

type ErrorAction struct {
	Action string // stop | retry
	Target string // same | another
}
