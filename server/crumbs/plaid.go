package crumbs

import (
	"github.com/getsentry/sentry-go"
)

const (
	PlaidItemIDTag = "plaid.item_id"
)

func IncludePlaidItemIDTag(span *sentry.Span, itemId string) {
	if span != nil {
		span.SetTag(PlaidItemIDTag, itemId)
	}
}
