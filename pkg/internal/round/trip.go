package round

import (
	"context"
	"net/http"
)

var (
	_ http.RoundTripper = &ObservabilityRoundTripper{}
)

type Handler func(ctx context.Context, request *http.Request, response *http.Response, err error)

type ObservabilityRoundTripper struct {
	handler Handler
	inner   http.RoundTripper
}

func (o *ObservabilityRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := o.inner.RoundTrip(request)

	o.handler(request.Context(), request, response, err)

	return response, err
}

func NewObservabilityRoundTripper(inner http.RoundTripper, handler Handler) http.RoundTripper {
	return &ObservabilityRoundTripper{
		handler: handler,
		inner:   inner,
	}
}
