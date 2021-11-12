package logging

import (
	"github.com/monetr/monetr/pkg/internal/ctxkeys"
	"github.com/sirupsen/logrus"
)

var (
	_ logrus.Formatter = &contextFormatter{}
)

type contextFormatter struct {
	inner logrus.Formatter
}

func NewContextFormatterWrapper(inner logrus.Formatter) logrus.Formatter {
	return &contextFormatter{
		inner: inner,
	}
}

func (c *contextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// If there isn't even a context on the provided entry, then there is no work for this wrapper to do. We can simply
	// call the inner formatter.
	if entry.Context == nil {
		return c.inner.Format(entry)
	}

	// We cannot safely modify anything on the entry object being passed to this method. So if we want to add fields
	// to the entry right before it actually gets logged we need to make a duplicate. The WithFields method is
	// thread-safe and allows us to add the fields we want.
	duplicate := duplicateEntry(entry, ctxkeys.LogrusFieldsFromContext(entry.Context, entry.Data))

	// Now that we have our new entry with (potentially) some additional helpful fields from the context, we can send
	// this off to the inner formatter.
	return c.inner.Format(duplicate)
}
