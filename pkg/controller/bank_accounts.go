package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/monetr/monetr/pkg/models"
	_ "github.com/monetr/monetr/pkg/swag"
	"net/http"
	"strings"
	"time"
)

func (c *Controller) handleBankAccounts(p iris.Party) {
	p.Get("/", c.getBankAccounts)
	p.Get("/{bankAccountId:uint64}/balances", c.getBalances)
	p.Post("/", c.postBankAccounts)
}

// List All Bank Accounts
// @Summary List All Bank Accounts
// @id list-all-bank-accounts
// @tags Bank Accounts
// @description Lists all of the bank accounts for the currently authenticated user.
// @Produce json
// @Security ApiKeyAuth
// @Router /bank_accounts [get]
// @Success 200 {array} swag.BankAccountResponse
// @Failure 402 {object} SubscriptionNotActiveError The user's subscription is not active.
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) getBankAccounts(ctx *context.Context) {
	repo := c.mustGetAuthenticatedRepository(ctx)
	bankAccounts, err := repo.GetBankAccounts(c.getContext(ctx))
	if err != nil {
		c.wrapPgError(ctx, err, "failed to retrieve bank accounts")
		return
	}

	ctx.JSON(bankAccounts)
}

// Get Bank Account Balances
// @Summary Get Bank Account Balances
// @id get-bank-account-balances
// @tags Bank Accounts
// @description Get the balances for the specified bank account (including calculated balances).
// @Security ApiKeyAuth
// @Produce json
// @Param bankAccountId path int true "Bank Account ID"
// @Router /bank_accounts/{bankAccountId}/balances [get]
// @Success 200 {object} repository.Balances
// @Failure 400 {object} InvalidBankAccountIdError Invalid Bank Account ID.
// @Failure 402 {object} SubscriptionNotActiveError The user's subscription is not active.
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) getBalances(ctx *context.Context) {
	bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
	if bankAccountId == 0 {
		c.returnError(ctx, http.StatusBadRequest, "must specify valid bank account Id")
		return
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	balances, err := repo.GetBalances(c.getContext(ctx), bankAccountId)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to retrieve balances")
		return
	}

	ctx.JSON(balances)
}

// Create Bank Account
// @Summary Create Bank Account
// @ID create-bank-account
// @tags Bank Accounts
// @description Create a bank account for the provided link. Note: Bank accounts can only be created this way for manual links. Attempting to create a bank account for a link that is associated with Plaid will result in an error.
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param newBankAccount body swag.CreateBankAccountRequest true "New Bank Account"
// @Router /bank_accounts [post]
// @Success 200 {object} swag.BankAccountResponse
// @Failure 402 {object} SubscriptionNotActiveError The user's subscription is not active.
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) postBankAccounts(ctx *context.Context) {
	var bankAccount models.BankAccount
	if err := ctx.ReadJSON(&bankAccount); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed JSON")
		return
	}

	if bankAccount.LinkId == 0 {
		c.returnError(ctx, http.StatusBadRequest, "link Id must be provided")
		return
	}

	bankAccount.BankAccountId = 0
	bankAccount.Name = strings.TrimSpace(bankAccount.Name)
	bankAccount.Mask = strings.TrimSpace(bankAccount.Mask)
	bankAccount.LastUpdated = time.Now().UTC()

	if bankAccount.Name == "" {
		c.returnError(ctx, http.StatusBadRequest, "bank account must have a name")
		return
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	// Bank accounts can only be created this way when they are associated with a link that allows manual
	// management. If the link they specified does not, then a bank account cannot be created for this link.
	isManual, err := repo.GetLinkIsManual(c.getContext(ctx), bankAccount.LinkId)
	if err != nil {
		c.wrapPgError(ctx, err, "could not validate link is manual")
		return
	}

	if !isManual {
		c.returnError(ctx, http.StatusBadRequest, "cannot create a bank account for a non-manual link")
		return
	}

	if err := repo.CreateBankAccounts(c.getContext(ctx), bankAccount); err != nil {
		c.wrapPgError(ctx, err, "could not create bank account")
		return
	}

	ctx.JSON(bankAccount)
}
