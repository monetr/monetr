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
	span := sentry.StartSpan(request.Context(), "http.client")
	defer span.Finish()
	span.Description = fmt.Sprintf("%s %s", strings.ToUpper(request.Method), request.URL.String())
	span.SetTag("http.url", request.URL.String())
	span.SetTag("http.query", "?"+request.URL.RawQuery)
	span.SetTag("http.request.method", request.Method)
	span.SetTag("server.address", request.URL.Hostname())
	span.SetTag("url.full", request.URL.String())
	span.SetTag("net.peer.name", request.URL.Host)

	response, err := o.inner.RoundTrip(request)
	span.SetTag("http.response.status_code", fmt.Sprint(response.StatusCode))
	// Both of these are the same, one is for opentelemetry, one is for sentry.
	span.SetTag("http.response.body.size", fmt.Sprint(response.ContentLength))
	span.SetTag("http.response_content_length", fmt.Sprint(response.ContentLength))

	span.Status = sentry.HTTPtoSpanStatus(response.StatusCode)

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
