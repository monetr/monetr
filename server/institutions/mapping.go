package institutions

type Mapping struct {
	Id                  string  `json:"id"`
	Name                string  `json:"name"`
	Website             *string `json:"website"`
	PlaidInstitutionId  *string `json:"plaidInstitutionId"`
	TellerInstitutionId *string `json:"tellerInstitutionId"`
	PrimaryColor        *string `json:"primaryColor"`
	MappedBy            string  `json:"mappedBy"`
}
