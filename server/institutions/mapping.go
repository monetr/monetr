package institutions

type Mapping struct {
	Id                  string  `json:"id"`
	Name                string  `json:"name"`
	Website             *string `json:"website"`
	Logo                *string `json:"logo"`
	PlaidInstitutionId  *string `json:"plaidInstitutionId"`
	TellerInstitutionId *string `json:"tellerInstitutionId"`
}
