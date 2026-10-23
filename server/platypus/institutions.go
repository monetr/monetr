package platypus

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/cache"
	plaid "github.com/plaid/plaid-go/v49"
)

type PlaidInstitutions interface {
	GetInstitution(ctx context.Context, institutionId string) (PlaidInstitution, error)
}

var (
	_ PlaidInstitutions = &plaidInstitutionsBase{}
)

type PlaidInstitution struct {
	InstitutionId string              `json:"institutionId"`
	Name          string              `json:"name"`
	Products      []plaid.Products    `json:"products"`
	CountryCodes  []plaid.CountryCode `json:"countryCodes"`
	URL           string              `json:"url,omitempty"`
	PrimaryColor  string              `json:"primaryColor,omitempty"`
	Logo          string              `json:"logo,omitempty"`
	Status        Status              `json:"status"`
}

type Status struct {
	ItemLogins          *plaid.ProductStatus `json:"item_logins,omitempty"`
	TransactionsUpdates *plaid.ProductStatus `json:"transactions_updates,omitempty"`
	Auth                *plaid.ProductStatus `json:"auth,omitempty"`
	Identity            *plaid.ProductStatus `json:"identity,omitempty"`
	InvestmentsUpdates  *plaid.ProductStatus `json:"investments_updates,omitempty"`
	LiabilitiesUpdates  *plaid.ProductStatus `json:"liabilities_updates,omitempty"`
	Liabilities         *plaid.ProductStatus `json:"liabilities,omitempty"`
	Investments         *plaid.ProductStatus `json:"investments,omitempty"`
}

func NewPlaidInstitution(input *plaid.Institution) PlaidInstitution {
	return PlaidInstitution{
		InstitutionId: input.GetInstitutionID(),
		Name:          input.GetName(),
		Products:      input.GetProducts(),
		CountryCodes:  input.GetCountryCodes(),
		URL:           deref(input.GetURL()),
		PrimaryColor:  deref(input.GetPrimaryColor()),
		Logo:          deref(input.GetLogo()),
		Status: Status{
			ItemLogins:          input.GetStatus().GetItemLogins(),
			TransactionsUpdates: input.GetStatus().GetTransactionsUpdates(),
			Auth:                input.GetStatus().GetAuth(),
			Identity:            input.GetStatus().GetIdentity(),
			InvestmentsUpdates:  input.GetStatus().GetInvestmentsUpdates(),
			LiabilitiesUpdates:  input.GetStatus().GetLiabilitiesUpdates(),
			Liabilities:         input.GetStatus().GetLiabilities(),
			Investments:         input.GetStatus().GetInvestments(),
		},
	}
}

type plaidInstitutionsBase struct {
	log      *slog.Logger
	platypus Platypus
	caching  cache.Cache
}

func NewPlaidInstitutionWrapper(log *slog.Logger, platypus Platypus, caching cache.Cache) PlaidInstitutions {
	return &plaidInstitutionsBase{
		log:      log,
		platypus: platypus,
		caching:  caching,
	}
}

func (p *plaidInstitutionsBase) GetInstitution(ctx context.Context, institutionId string) (PlaidInstitution, error) {
	span := sentry.StartSpan(ctx, "GetInstitution")
	defer span.Finish()

	span.Data = map[string]any{
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

	institution = NewPlaidInstitution(result)

	if err = p.caching.SetEzTTL(span.Context(), p.cacheKey(institutionId), institution, 30*time.Minute); err != nil {
		p.log.WarnContext(span.Context(), "failed to cache institution details", "institutionId", institutionId, "err", err)
	}

	return institution, nil
}

func (*plaidInstitutionsBase) cacheKey(institutionId string) string {
	return fmt.Sprintf("plaid:institutions:%s", institutionId)
}
