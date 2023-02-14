package controller

import (
	"net/http"
	"strings"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/monetr/monetr/pkg/models"
)

func (c *Controller) handleBankAccounts(p iris.Party) {
	p.Get("/", c.getBankAccounts)
	p.Put("/{bankAccountId:uint64}", c.putBankAccounts)
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
func (c *Controller) getBankAccounts(ctx iris.Context) {
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
func (c *Controller) getBalances(ctx iris.Context) {
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
func (c *Controller) postBankAccounts(ctx iris.Context) {
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

func (c *Controller) putBankAccounts(ctx iris.Context) {
	bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
	if bankAccountId == 0 {
		c.badRequest(ctx, "must specify a valid bank account Id")
		return
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	existingBankAccount, err := repo.GetBankAccount(c.getContext(ctx), bankAccountId)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to retrieve bank account")
		return
	}

	var request struct {
		AvailableBalance int64                    `json:"availableBalance"`
		CurrentBalance   int64                    `json:"currentBalance"`
		Mask             string                   `json:"mask"`
		Name             string                   `json:"name"`
		Status           models.BankAccountStatus `json:"status"`
	}
	if err = ctx.ReadJSON(&request); err != nil {
		c.invalidJson(ctx)
		return
	}

	// TODO Eventually we should just query the link to see if its a manual link. But for now if the bank account has a
	//  plaid account ID then its probably safe to assume that it is a Plaid managed bank account.
	if existingBankAccount.PlaidAccountId != "" {
		existingBankAccount.Name = strings.TrimSpace(request.Name)
	} else {
		existingBankAccount.AvailableBalance = request.AvailableBalance
		existingBankAccount.CurrentBalance = request.CurrentBalance
		existingBankAccount.Name = strings.TrimSpace(request.Name)
		existingBankAccount.Mask = strings.TrimSpace(request.Mask)
		// TODO Verify the provided status string is correct.
		existingBankAccount.Status = request.Status
	}

	// TODO This might not reflect a correct updatedAt in the resulting value. Because this is not being passed by
	//   reference.
	if err = repo.UpdateBankAccounts(c.getContext(ctx), *existingBankAccount); err != nil {
		c.wrapPgError(ctx, err, "failed to update bank account")
		return
	}

	ctx.JSON(*existingBankAccount)
}
