package billing

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetrapp/rest-api/pkg/cache"
	"github.com/monetrapp/rest-api/pkg/internal/stripe_helper"
	"github.com/monetrapp/rest-api/pkg/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v72"
	"strings"
	"time"
)

var (
	ErrUserAlreadyHasCustomer = errors.New("user already has customer")
)

type Billing interface {
	CreateCustomer(ctx context.Context, login models.Login, user *models.User) (*stripe.Customer, error)

	GetIsTrial(ctx context.Context, accountId uint64) (bool, error)
	GetTrialEndTime(ctx context.Context, accountId uint64) (time.Time, error)

	GetPaymentMethods(ctx context.Context, accountId uint64) ([]models.PaymentMethod, error)
	AttachPaymentMethod(ctx context.Context, accountId uint64, method *models.PaymentMethod) error

	GetInvoices(ctx context.Context, accountId uint64) ([]models.Invoice, error)
	CreateInvoice(ctx context.Context, accountId uint64, invoice *models.Invoice) error

	CreateSubscription(ctx context.Context, accountId uint64) error
	GetActiveSubscription(ctx context.Context, accountId uint64) (*models.Subscription, error)
	CancelSubscription(ctx context.Context, accountId uint64) error

	Close() error
}

var (
	_ Billing = &billingBase{}
)

type billingBase struct {
	log    *logrus.Entry
	memory cache.Cache
	stripe stripe_helper.Stripe
	db     pg.DBI
}

func (b *billingBase) CreateCustomer(ctx context.Context, login models.Login, user *models.User) (*stripe.Customer, error) {
	span := sentry.StartSpan(ctx, "CreateCustomer")
	defer span.Finish()

	log := b.log.WithFields(logrus.Fields{
		"accountId": user.AccountId,
		"loginId":   user.LoginId,
		"userId":    user.UserId,
	})

	if user.StripeCustomerId != nil {
		log.Warn("user already has a stripe_customer_id")
		return nil, errors.WithStack(ErrUserAlreadyHasCustomer)
	}

	name := stripe.String(strings.TrimSpace(login.FirstName + " " + login.LastName))
	stripeCustomerParams := stripe.CustomerParams{
		Description: name,
		Email:       &login.Email,
		Name:        name,
		Phone:       nil,
	}

	log.Trace("creating stripe customer for user/login")
	customer, err := b.stripe.CreateCustomer(span.Context(), stripeCustomerParams)
	if err != nil {
		log.WithError(err).Error("failed to create stripe customer")
		return nil, err
	}

	log.Trace("updating user record with new stripe_customer_id")
	_, err = b.db.ModelContext(span.Context(), &user).
		Set(`"stripe_customer_id" = ?`, customer.ID).
		WherePK().
		Update(&user)
	if err != nil {
		log.WithError(err).Error("failed to update user record with new stripe_customer_id")
		return nil, err
	}

	return customer, nil
}

func (b *billingBase) GetIsTrial(ctx context.Context, accountId uint64) (bool, error) {
	panic("implement me")
}

func (b *billingBase) GetTrialEndTime(ctx context.Context, accountId uint64) (time.Time, error) {
	panic("implement me")
}

func (b *billingBase) GetPaymentMethods(ctx context.Context, accountId uint64) ([]models.PaymentMethod, error) {
	panic("implement me")
}

func (b *billingBase) AttachPaymentMethod(ctx context.Context, accountId uint64, method *models.PaymentMethod) error {
	panic("implement me")
}

func (b *billingBase) GetInvoices(ctx context.Context, accountId uint64) ([]models.Invoice, error) {
	panic("implement me")
}

func (b *billingBase) CreateInvoice(ctx context.Context, accountId uint64, invoice *models.Invoice) error {
	panic("implement me")
}

func (b *billingBase) CreateSubscription(ctx context.Context, accountId uint64) error {
	panic("implement me")
}

func (b *billingBase) GetActiveSubscription(ctx context.Context, accountId uint64) (*models.Subscription, error) {
	panic("implement me")
}

func (b *billingBase) CancelSubscription(ctx context.Context, accountId uint64) error {
	panic("implement me")
}

func (b *billingBase) Close() error {
	panic("implement me")
}
