package teller

type InstitutionCapability string

const (
	DetailInstitutionCapability      InstitutionCapability = "detail"
	BalanceInstitutionCapability     InstitutionCapability = "balance"
	TransactionInstitutionCapability InstitutionCapability = "transaction"
	IdentityInstitutionCapability    InstitutionCapability = "identity"
)

type Institution struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	Capabilities []InstitutionCapability
}
