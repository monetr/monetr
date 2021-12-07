package platypus

import (
	"context"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/pkg/cache"
	"github.com/plaid/plaid-go/plaid"
	"github.com/sirupsen/logrus"
)

type PlaidInstitutions interface {
	GetInstitution(ctx context.Context, institutionId string) (*plaid.Institution, error)
}

var (
	_ PlaidInstitutions = &plaidInstitutionsBase{}
)

type plaidInstitutionsBase struct {
	log      *logrus.Entry
	platypus Platypus
	caching  cache.Cache
}

func NewPlaidInstitutionWrapper(log *logrus.Entry, platypus Platypus, caching cache.Cache) PlaidInstitutions {
	return &plaidInstitutionsBase{
		log:      log,
		platypus: platypus,
		caching:  caching,
	}
}

func (p *plaidInstitutionsBase) GetInstitution(ctx context.Context, institutionId string) (*plaid.Institution, error) {
	span := sentry.StartSpan(ctx, "GetInstitution")
	defer span.Finish()

	span.Data = map[string]interface{}{
		"institutionId": institutionId,
	}

	{ // Check to see if the institution is in the cache.
		var institution plaid.Institution
		if err := p.caching.GetEz(span.Context(), p.cacheKey(institutionId), institution); err == nil && institution.InstitutionId != "" {
			return &institution, nil
		}
	}

	result, err := p.platypus.GetInstitution(span.Context(), institutionId)
	if err != nil {
		return nil, err
	}

	if err = p.caching.SetEzTTL(span.Context(), p.cacheKey(institutionId), result, 30*time.Minute); err != nil {
		p.log.WithField("institutionId", institutionId).WithError(err).Warn("failed to cache institution details")
	}

	return result, nil
}

func (p *plaidInstitutionsBase) cacheKey(institutionId string) string {
	return fmt.Sprintf("plaid:institutions:%s", institutionId)
}
