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
	GetInstitution(ctx context.Context, institutionId string) (PlaidInstitution, error)
}

var (
	_ PlaidInstitutions = &plaidInstitutionsBase{}
)

type PlaidInstitution struct {
	InstitutionId string                  `json:"institutionId"`
	Name          string                  `json:"name"`
	Products      []plaid.Products        `json:"products"`
	CountryCodes  []plaid.CountryCode     `json:"countryCodes"`
	URL           string                  `json:"url,omitempty"`
	PrimaryColor  string                  `json:"primaryColor,omitempty"`
	Logo          string                  `json:"logo,omitempty"`
	Status        plaid.InstitutionStatus `json:"status"`
}

func NewPlaidInstitution(input plaid.Institution) PlaidInstitution {
	return PlaidInstitution{
		InstitutionId: input.GetInstitutionId(),
		Name:          input.GetName(),
		Products:      input.GetProducts(),
		CountryCodes:  input.GetCountryCodes(),
		URL:           input.GetUrl(),
		PrimaryColor:  input.GetPrimaryColor(),
		Logo:          input.GetLogo(),
		Status:        input.GetStatus(),
	}
}

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

func (p *plaidInstitutionsBase) GetInstitution(ctx context.Context, institutionId string) (PlaidInstitution, error) {
	span := sentry.StartSpan(ctx, "GetInstitution")
	defer span.Finish()

	span.Data = map[string]interface{}{
		"institutionId": institutionId,
	}

	var institution PlaidInstitution
	{ // Check to see if the institution is in the cache.
		if err := p.caching.GetEz(span.Context(), p.cacheKey(institutionId), &institution); err == nil && institution.InstitutionId != "" {
			return institution, nil
		}
	}

	result, err := p.platypus.GetInstitution(span.Context(), institutionId)
	if err != nil {
		return institution, err
	}

	institution = NewPlaidInstitution(*result)

	if err = p.caching.SetEzTTL(span.Context(), p.cacheKey(institutionId), institution, 30*time.Minute); err != nil {
		p.log.WithField("institutionId", institutionId).WithError(err).Warn("failed to cache institution details")
	}

	return institution, nil
}

func (p *plaidInstitutionsBase) cacheKey(institutionId string) string {
	return fmt.Sprintf("plaid:institutions:%s", institutionId)
}
