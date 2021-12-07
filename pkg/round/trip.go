package round

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/getsentry/sentry-go"
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
	span := sentry.StartSpan(request.Context(), fmt.Sprintf("%s %s", strings.ToUpper(request.Method), request.URL.Path))
	defer span.Finish()
	response, err := o.inner.RoundTrip(request)

	o.handler(request.Context(), request, response, err)

	if err != nil || response.StatusCode > http.StatusPermanentRedirect {
		span.Status = sentry.SpanStatusInternalError
	} else {
		span.Status = sentry.SpanStatusOK
	}

	return response, err
}

func NewObservabilityRoundTripper(inner http.RoundTripper, handler Handler) http.RoundTripper {
	return &ObservabilityRoundTripper{
		handler: handler,
		inner:   inner,
	}
}
