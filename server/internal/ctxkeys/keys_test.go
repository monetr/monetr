package ctxkeys

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlogAttrsFromContext(t *testing.T) {
	t.Run("extracts known keys from context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), AccountID, uint64(5678))

		attrs := SlogAttrsFromContext(ctx)
		assert.Len(t, attrs, 1, "should produce one attr")
		assert.Equal(t, slog.Any("accountId", uint64(5678)), attrs[0])
	})

	t.Run("extracts multiple keys", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), AccountID, uint64(5678))
		ctx = context.WithValue(ctx, UserID, uint64(9876))

		attrs := SlogAttrsFromContext(ctx)
		assert.Len(t, attrs, 2)

		keys := map[string]bool{}
		for _, a := range attrs {
			keys[a.Key] = true
		}
		assert.True(t, keys["accountId"])
		assert.True(t, keys["userId"])
	})

	t.Run("empty context produces no attrs", func(t *testing.T) {
		attrs := SlogAttrsFromContext(context.Background())
		assert.Empty(t, attrs)
	})
}
