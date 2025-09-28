package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/cache"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func buildAccountCacheKey(accountId ID[Account]) string {
	return fmt.Sprintf("accounts:%s", accountId)
}

// AccountsRepository is a more global repository for accessing and modifying
// data regarding account objects. It is not scoped to a single user and can
// access any account. It also does implement some caching to make account reads
// less expensive. Writes via this interface will also propagate to the cache
// where as writes via other methods will not.
type AccountsRepository interface {
	GetAccount(ctx context.Context, accountId ID[Account]) (*Account, error)
	GetAccountByCustomerId(ctx context.Context, stripeCustomerId string) (*Account, error)
	UpdateAccount(ctx context.Context, account *Account) error
}

var (
	_ AccountsRepository = &accountsRepositoryBase{}
)

type accountsRepositoryBase struct {
	log   *logrus.Entry
	cache cache.Cache
	db    pg.DBI
}

func NewAccountRepository(
	log *logrus.Entry,
	cacheClient cache.Cache,
	db pg.DBI,
) AccountsRepository {
	return &accountsRepositoryBase{
		log:   log,
		cache: cacheClient,
		db:    db,
	}
}

func (p *accountsRepositoryBase) GetAccount(
	ctx context.Context,
	accountId ID[Account],
) (*Account, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Status = sentry.SpanStatusOK

	log := p.log.WithContext(span.Context()).WithField("accountId", accountId)

	var account Account
	if err := p.cache.GetEz(
		span.Context(),
		buildAccountCacheKey(accountId),
		&account,
	); err != nil {
		log.WithError(err).Errorf("failed to retrieve account data from cache")
	}

	if !account.AccountId.IsZero() {
		log.Trace("returning account from cache")
		return &account, nil
	}

	if err := p.db.ModelContext(span.Context(), &account).
		Where(`"account"."account_id" = ?`, accountId).
		Limit(1).
		Select(&account); err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve account by Id")
	}

	if err := p.cache.SetEzTTL(
		span.Context(),
		buildAccountCacheKey(accountId),
		account,
		30*time.Minute,
	); err != nil {
		log.WithError(err).Warn("failed to store account in cache")
	}

	return &account, nil
}

func (p *accountsRepositoryBase) GetAccountByCustomerId(
	ctx context.Context,
	stripeCustomerId string,
) (*Account, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var account Account
	if err := p.db.ModelContext(span.Context(), &account).
		Where(`"account"."stripe_customer_id" = ?`, stripeCustomerId).
		Limit(1).
		Select(&account); err != nil {

		span.Status = sentry.SpanStatusInternalError
		span.SetData("stripeCustomerId", stripeCustomerId)

		return nil, errors.Wrap(err, "failed to retrieve account by customer Id")
	}

	return &account, nil
}

func (p *accountsRepositoryBase) UpdateAccount(ctx context.Context, account *Account) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	log := p.log.
		WithContext(span.Context()).
		WithField("accountId", account.AccountId)

	log.Debug("updating account")

	_, err := p.db.ModelContext(span.Context(), account).
		Where(`"account"."account_id" = ?`, account.AccountId).
		Update(account)
	if err != nil {
		log.WithError(err).Errorf("failed to update account")
		return errors.Wrap(err, "failed to update account")
	}

	if err = p.cache.SetEzTTL(
		span.Context(),
		buildAccountCacheKey(account.AccountId),
		account,
		30*time.Minute,
	); err != nil {
		log.WithError(err).Warn("failed to store account in cache")
	}

	return nil
}

func (r *repositoryBase) GetAccount(ctx context.Context) (*Account, error) {
	if r.account != nil {
		return r.account, nil
	}

	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var account Account
	err := r.txn.ModelContext(span.Context(), &account).
		Where(`"account"."account_id" = ?`, r.AccountId()).
		Limit(1).
		Select(&account)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve account")
	}

	span.Status = sentry.SpanStatusOK

	r.account = &account

	return r.account, nil
}
