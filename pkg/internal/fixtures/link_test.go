package fixtures

import (
	"context"
	"testing"

	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
)

func TestGivenIHaveAPlaidLink(t *testing.T) {
	testutils.ForEachDatabase(t, func(ctx context.Context, t *testing.T, db *bun.DB) {
		user, _ := GivenIHaveABasicAccount(t)

		link := GivenIHaveAPlaidLink(t, user)
		assert.NotZero(t, link.LinkId, "link must have been created")
		assert.Equal(t, user.UserId, link.CreatedByUserId, "link must have been created by the provided user")
		assert.NotNil(t, link.CreatedByUser, "user object should be included for created by")
		assert.NotNil(t, link.PlaidLinkId, "plaid link should have been created")
		assert.NotNil(t, link.PlaidLink, "plaid link object should be included")
	})
}
