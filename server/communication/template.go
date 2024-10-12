package communication

type Email interface {
	EmailAddress() string
	Name() (firstName, lastName string)
	Template() string
	Subject() string
}

var (
	_ Email = VerifyEmailParams{}
)

type VerifyEmailParams struct {
	BaseURL      string
	Email        string
	FirstName    string
	LastName     string
	SupportEmail string
	VerifyURL    string
}

func (p VerifyEmailParams) EmailAddress() string {
	return p.Email
}

func (p VerifyEmailParams) Name() (firstName, lastName string) {
	return p.FirstName, p.LastName
}

func (VerifyEmailParams) Template() string {
	return "VerifyEmailAddress"
}

func (VerifyEmailParams) Subject() string {
	return "Verify Your Email Address"
}

type PasswordResetParams struct {
	BaseURL      string
	Email        string
	FirstName    string
	LastName     string
	SupportEmail string
	ResetURL     string
}

func (p PasswordResetParams) EmailAddress() string {
	return p.Email
}

func (p PasswordResetParams) Name() (firstName, lastName string) {
	return p.FirstName, p.LastName
}

func (PasswordResetParams) Template() string {
	return "ForgotPassword"
}

func (PasswordResetParams) Subject() string {
	return "Reset Your Password"
}

type PasswordChangedParams struct {
	BaseURL      string
	Email        string
	FirstName    string
	LastName     string
	SupportEmail string
}

func (p PasswordChangedParams) EmailAddress() string {
	return p.Email
}

func (p PasswordChangedParams) Name() (firstName, lastName string) {
	return p.FirstName, p.LastName
}

func (PasswordChangedParams) Template() string {
	return "PasswordChanged"
}

func (PasswordChangedParams) Subject() string {
	return "Password Updated"
}

type PlaidDisconnectedParams struct {
	BaseURL      string
	Email        string
	FirstName    string
	LastName     string
	LinkName     string
	LinkURL      string
	SupportEmail string
}

func (p PlaidDisconnectedParams) EmailAddress() string {
	return p.Email
}

func (p PlaidDisconnectedParams) Name() (firstName, lastName string) {
	return p.FirstName, p.LastName
}

func (PlaidDisconnectedParams) Template() string {
	return "PlaidDisconnected"
}

func (PlaidDisconnectedParams) Subject() string {
	return "Account Disconnected"
}

type TrialAboutToExpireParams struct {
	BaseURL               string
	Email                 string
	FirstName             string
	LastName              string
	TrialExpirationDate   string
	TrialExpirationWindow string
	SupportEmail          string
}

func (p TrialAboutToExpireParams) EmailAddress() string {
	return p.Email
}

func (p TrialAboutToExpireParams) Name() (firstName, lastName string) {
	return p.FirstName, p.LastName
}

func (TrialAboutToExpireParams) Template() string {
	return "TrialAboutToExpire"
}

func (TrialAboutToExpireParams) Subject() string {
	return "Trial About To Expire"
}
