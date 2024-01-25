package teller

import (
	"context"
	"testing"

	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestGetHealth(t *testing.T) {
	log := testutils.GetLog(t)
	client, err := NewClient(log, config.Teller{
		Enabled:       true,
		ApplicationId: "app_abc123",
	})
	assert.NoError(t, err, "must not have an error creating a client without a certificate")
	assert.NoError(t, client.GetHealth(context.Background()), "must pass health check")
}
