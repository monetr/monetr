package models

type InstitutionStatus string

const (
	Healthy  InstitutionStatus = "HEALTHY"
	Degraded InstitutionStatus = "DEGRADED"
	Down     InstitutionStatus = "DOWN"
)

type Institution struct {
	tableName string `pg:"institutions"`

	InstitutionId      uint64   `json:"institutionId" pg:"institution_id,notnull,pk,type:'bigserial'"`
	Name               string   `json:"name" pg:"name,notnull"`
	PlaidInstitutionId *string  `json:"-" pg:"plaid_institution_id,unique"`
	PlaidProducts      []string `json:"-" pg:"plaid_products,type:'text[]'"`
	URL                *string  `json:"url" pg:"url"`
	PrimaryColor       *string  `json:"primaryColor" pg:"primary_color"`
	Logo               *string  `json:"logo" pg:"logo"`
}
