package controller

type ApiError struct {
	Error string `json:"error"`
}

type InvalidBankAccountIdError struct {
	Error string `json:"error" example:"invalid bank account Id provided"`
}
