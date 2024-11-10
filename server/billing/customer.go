package billing

import (
	"context"

	"github.com/monetr/monetr/server/build"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"github.com/stripe/stripe-go/v81"
)

// CreateCustomer takes an account object and will create a stripe customer for
// that account if one does not already exist. It will then store this in the
// database and update any cache. As well as update it on the provided object.
// If the account provided already has a Stripe customer associated with it then
// this function will do nothing and return nil.
func (b *baseBilling) CreateCustomer(
	ctx context.Context,
	owner Login,
	account *Account,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	if account.StripeCustomerId != nil {
		return errors.New("account already has a stripe customer")
	}

	crumbs.Debug(span.Context(), "Account does not have a Stripe customer, a new one will be created.", nil)

	name := owner.Name()
	customer, err := b.stripe.CreateCustomer(span.Context(), stripe.CustomerParams{
		Email: &owner.Email,
		Name:  &name,
		Metadata: map[string]string{
			"accountId":   account.AccountId.String(),
			"environment": b.config.Environment,
			"loginId":     owner.LoginId.String(),
			"release":     build.Release,
			"revision":    build.Revision,
		},
	})
	if err != nil {
		return err
	}

	// Stash the newly created customer ID on the account.
	account.StripeCustomerId = &customer.ID

	// Then store those details for later.
	return b.accounts.UpdateAccount(span.Context(), account)
}
