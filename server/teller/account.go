package teller

type AccountType string

const (
	AccountTypeDepository AccountType = "depository"
	AccountTypeCredit     AccountType = "credit"
)

type AccountSubType string

const (
	// Depository
	AccountSubTypeChecking             AccountSubType = "checking"
	AccountSubTypeSavings              AccountSubType = "savings"
	AccountSubTypeMoneyMarket          AccountSubType = "money_market"
	AccountSubTypeCertificateOfDeposit AccountSubType = "certificate_of_deposit"
	AccountSubTypeTreasury             AccountSubType = "treasury"
	AccountSubTypeSweep                AccountSubType = "sweep"

	// Credit
	AccountSubTypeCreditCard AccountSubType = "credit_card"
)

type AccountStatus string

const (
	AccountStatusOpen   AccountStatus = "open"
	AccountStatusClosed AccountStatus = "closed"
)

type Account struct {
	Id           string `json:"id"`
	Currency     string `json:"currency"`
	EnrollmentId string `json:"enrollment_id"`
	Institution  struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	} `json:"institution"`
	Mask    string            `json:"last_four"`
	Links   map[string]string `json:"links"`
	Name    string            `json:"name"`
	Type    AccountType       `json:"type"`
	SubType AccountSubType    `json:"subtype"`
	Status  AccountStatus     `json:"status"`
}
