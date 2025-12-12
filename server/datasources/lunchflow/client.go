package lunchflow

import (
	"context"
	"encoding/json"
)

type AccountId string

type Account struct {
	Id              AccountId `json:"id"`
	Name            string    `json:"name"`
	InstitutionName string    `json:"institution_name"`
	InstitutionLogo *string   `json:"institution_logo"`
	Provider        string    `json:"provider"`
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
}

type Balance struct {
	Amount   json.Number `json:"amount"`
	Currency string      `json:"currency"`
}

type Transaction struct {
	Id          string      `json:"id"`
	AccountId   AccountId   `json:"accountId"`
	Amount      json.Number `json:"amount"`
	Currency    string      `json:"currency"`
	Date        string      `json:"date"`
	Merchant    *string     `json:"merchant"`
	Description *string     `json:"description"`
}

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=client.go -package=mockgen -destination=../../internal/mockgen/lunchflow_client.go LunchFlowClient
type LunchFlowClient interface {
	GetAccounts(ctx context.Context) ([]Account, error)
	GetBalance(ctx context.Context, accountId AccountId) (*Balance, error)
	GetTransactions(ctx context.Context, accountId AccountId) ([]Transaction, error)
}
