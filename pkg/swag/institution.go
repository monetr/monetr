package swag

import (
	"time"

	"github.com/plaid/plaid-go/plaid"
)

type InstitutionStatus uint8

const (
	UnknownInstitutionStatus InstitutionStatus = 0
	HealthyInstitutionStatus InstitutionStatus = iota
	DegradedInstitutionStatus
	DownInstitutionStatus
)

type InstitutionResponse struct {
	Name         string                    `json:"name"`
	URL          *string                   `json:"url,omitempty"`
	PrimaryColor *string                   `json:"primaryColor,omitempty"`
	Logo         *string                   `json:"logo,omitempty"`
	Status       InstitutionStatusResponse `json:"status"`
}

type InstitutionStatusResponse struct {
	Login          bool                       `json:"login"`
	Transactions   bool                       `json:"transactions"`
	PlaidIncidents []InstitutionPlaidIncident `json:"plaidIncidents"`
}

type InstitutionPlaidIncident struct {
	Start time.Time  `json:"start"`
	End   *time.Time `json:"end"`
	Title string     `json:"title"`
}

func NewInstitutionResponse(plaidInstitution *plaid.Institution) InstitutionResponse {
	ins := InstitutionResponse{
		Name:         plaidInstitution.Name,
		URL:          plaidInstitution.Url.Get(),
		PrimaryColor: plaidInstitution.PrimaryColor.Get(),
		Logo:         plaidInstitution.Logo.Get(),
		Status: InstitutionStatusResponse{
			Login:          plaidInstitution.Status != nil && plaidInstitution.Status.ItemLogins.Status == "HEALTHY",
			Transactions:   plaidInstitution.Status != nil && plaidInstitution.Status.TransactionsUpdates.Status == "HEALTHY",
			PlaidIncidents: nil,
		},
	}

	if plaidInstitution.Status != nil {
		ins.Status.PlaidIncidents = make([]InstitutionPlaidIncident, len(plaidInstitution.Status.HealthIncidents))
		for i, plaidIncident := range plaidInstitution.Status.HealthIncidents {
			ins.Status.PlaidIncidents[i] = InstitutionPlaidIncident{
				Start: plaidIncident.StartDate,
				End:   plaidIncident.EndDate,
				Title: plaidIncident.Title,
			}
		}
	}

	return ins
}
