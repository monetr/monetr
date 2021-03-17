package fixtures

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"testing"
)

func ManualLink(t *testing.T) *models.Link {
	return &models.Link{
		LinkType:              models.ManualLinkType,
		InstitutionName:       fmt.Sprintf("%s Bank", gofakeit.Company()),
		CustomInstitutionName: "Personal Bank",
	}
}

func PlaidLink(t *testing.T) (*models.Link, *models.PlaidLink) {
	return &models.Link{
			LinkType:              models.ManualLinkType,
			InstitutionName:       fmt.Sprintf("%s Bank", gofakeit.Company()),
			CustomInstitutionName: "Personal Bank",
		}, &models.PlaidLink{
			ItemId:      gofakeit.UUID(),
			AccessToken: gofakeit.UUID(),
			Products: []string{
				"transactions",
			},
			WebhookUrl:      "",
			InstitutionId:   "1234",
			InstitutionName: gofakeit.Company(),
		}
}
