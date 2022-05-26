package crumbs

import (
	"context"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
)

func WrapError(ctx context.Context, err error, message string) error {
	Error(ctx, message, "error", nil)
	return errors.Wrap(err, message)
}

func Error(ctx context.Context, message, category string, data map[string]interface{}) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.AddBreadcrumb(&sentry.Breadcrumb{
			Type:      "error",
			Category:  category,
			Message:   message,
			Data:      data,
			Level:     sentry.LevelError,
			Timestamp: time.Now(),
		}, nil)
	}
}

func ReportError(ctx context.Context, err error, message, category string, data map[string]interface{}) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.CaptureException(err)
		hub.AddBreadcrumb(&sentry.Breadcrumb{
			Type:      "error",
			Category:  category,
			Message:   message,
			Data:      data,
			Level:     sentry.LevelError,
			Timestamp: time.Now(),
		}, nil)
	}
}

func Warn(ctx context.Context, message, category string, data map[string]interface{}) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.AddBreadcrumb(&sentry.Breadcrumb{
			Type:      "warning",
			Category:  category,
			Message:   message,
			Data:      data,
			Level:     sentry.LevelWarning,
			Timestamp: time.Now(),
		}, nil)
	}
}

func Debug(ctx context.Context, message string, data map[string]interface{}) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.AddBreadcrumb(&sentry.Breadcrumb{
			Type:      "debug",
			Category:  "debug",
			Message:   message,
			Data:      data,
			Level:     sentry.LevelDebug,
			Timestamp: time.Now(),
		}, nil)
	}
}

func HTTP(ctx context.Context, message, category, url, method string, statusCode int, data map[string]interface{}) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		if data == nil {
			data = map[string]interface{}{}
		}

		data["url"] = url
		data["method"] = method
		data["status_code"] = statusCode
		data["reason"] = http.StatusText(statusCode)

		level := sentry.LevelInfo
		if statusCode >= 400 {
			level = sentry.LevelError
		}

		hub.AddBreadcrumb(&sentry.Breadcrumb{
			Type:      "http",
			Category:  category,
			Message:   message,
			Data:      data,
			Level:     level,
			Timestamp: time.Now(),
		}, nil)
	}
}

func AddTag(ctx context.Context, name, value string) {
	sentry.GetHubFromContext(ctx).ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag(name, value)
	})
}
