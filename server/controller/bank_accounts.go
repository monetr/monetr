package controller

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	. "github.com/monetr/monetr/server/models"
)

func (c *Controller) getBankAccounts(ctx echo.Context) error {
	repo := c.mustGetAuthenticatedRepository(ctx)
	bankAccounts, err := repo.GetBankAccounts(c.getContext(ctx))
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve bank accounts")
	}

	return ctx.JSON(http.StatusOK, bankAccounts)
}

func (c *Controller) getBankAccount(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	bankAccount, err := repo.GetBankAccount(c.getContext(ctx), bankAccountId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve bank account")
	}

	return ctx.JSON(http.StatusOK, bankAccount)
}

func (c *Controller) getBalances(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	balances, err := repo.GetBalances(c.getContext(ctx), bankAccountId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve balances")
	}

	return ctx.JSON(http.StatusOK, balances)
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
func (c *Controller) postBankAccounts(ctx echo.Context) error {
	var bankAccount BankAccount
	if err := ctx.Bind(&bankAccount); err != nil {
		return c.invalidJson(ctx)
	}

	if bankAccount.LinkId.IsZero() {
		return c.badRequest(ctx, "link ID must be provided")
	}

	bankAccount.BankAccountId = ""
	bankAccount.Name = strings.TrimSpace(bankAccount.Name)
	bankAccount.Mask = strings.TrimSpace(bankAccount.Mask)
	bankAccount.LastUpdated = c.Clock.Now().UTC()

	if bankAccount.Name == "" {
		return c.badRequest(ctx, "bank account must have a name")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	// Bank accounts can only be created this way when they are associated with a link that allows manual
	// management. If the link they specified does not, then a bank account cannot be created for this link.
	isManual, err := repo.GetLinkIsManual(c.getContext(ctx), bankAccount.LinkId)
	if err != nil {
		return c.wrapPgError(ctx, err, "could not validate link is manual")
	}

	if !isManual {
		return c.badRequest(ctx, "cannot create a bank account for a non-manual link")
	}

	if err := repo.CreateBankAccounts(c.getContext(ctx), &bankAccount); err != nil {
		return c.wrapPgError(ctx, err, "could not create bank account")
	}

	return ctx.JSON(http.StatusOK, bankAccount)
}

func (c *Controller) putBankAccounts(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	existingBankAccount, err := repo.GetBankAccount(c.getContext(ctx), bankAccountId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve bank account")
	}

	var request struct {
		AvailableBalance int64             `json:"availableBalance"`
		CurrentBalance   int64             `json:"currentBalance"`
		Mask             string            `json:"mask"`
		Name             string            `json:"name"`
		Status           BankAccountStatus `json:"status"`
	}
	if err = ctx.Bind(&request); err != nil {
		return c.invalidJson(ctx)
	}

	// TODO Eventually we should just query the link to see if its a manual link.
	// But for now if the bank account has a plaid account ID then its probably
	// safe to assume that it is a Plaid managed bank account.
	if existingBankAccount.PlaidBankAccountId != nil {
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
		return c.wrapPgError(ctx, err, "failed to update bank account")
	}

	return ctx.JSON(http.StatusOK, *existingBankAccount)
}
