package fixtures

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGivenIHaveAPlaidLink(t *testing.T) {
	user, _ := GivenIHaveABasicAccount(t)

	link := GivenIHaveAPlaidLink(t, user)
	assert.NotZero(t, link.LinkId, "link must have been created")
	assert.Equal(t, user.UserId, link.CreatedByUserId, "link must have been created by the provided user")
	assert.Equal(t, user.UserId, *link.UpdatedByUserId, "link must have been updated by the provided user")
	assert.NotNil(t, link.CreatedByUser, "user object should be included for created by")
	assert.NotNil(t, link.UpdatedByUser, "user object should be included for updated by")
	assert.NotNil(t, link.PlaidLinkId, "plaid link should have been created")
	assert.NotNil(t, link.PlaidLink, "plaid link object should be included")
}
