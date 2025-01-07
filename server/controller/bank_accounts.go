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

func (c *Controller) postBankAccounts(ctx echo.Context) error {
	var bankAccount BankAccount
	if err := ctx.Bind(&bankAccount); err != nil {
		return c.invalidJson(ctx)
	}

	if bankAccount.LinkId.IsZero() {
		return c.badRequest(ctx, "Link ID must be provided")
	}

	var err error
	bankAccount.BankAccountId = ""
	bankAccount.Name, err = c.cleanString(ctx, "Name", bankAccount.Name)
	if err != nil {
		return err
	}

	bankAccount.Mask, err = c.cleanString(ctx, "Mask", bankAccount.Mask)
	if err != nil {
		return err
	}

	// TODO Should mask be enforced to be numeric only?
	if len(bankAccount.Mask) > 4 {
		return c.badRequest(ctx, "Mask cannot be more than 4 characters")
	}

	bankAccount.Status = ParseBankAccountStatus(bankAccount.Status)
	bankAccount.Type = ParseBankAccountType(bankAccount.Type)
	bankAccount.SubType = ParseBankAccountSubType(bankAccount.SubType)
	bankAccount.LastUpdated = c.Clock.Now().UTC()

	if bankAccount.Name == "" {
		return c.badRequest(ctx, "Bank account must have a name")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	// Bank accounts can only be created this way when they are associated with a link that allows manual
	// management. If the link they specified does not, then a bank account cannot be created for this link.
	isManual, err := repo.GetLinkIsManual(c.getContext(ctx), bankAccount.LinkId)
	if err != nil {
		return c.wrapPgError(ctx, err, "Could not validate link is manual")
	}

	if !isManual {
		return c.badRequest(ctx, "Cannot create a bank account for a non-manual link")
	}

	if err := repo.CreateBankAccounts(c.getContext(ctx), &bankAccount); err != nil {
		return c.wrapPgError(ctx, err, "Could not create bank account")
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
		AvailableBalance int64              `json:"availableBalance"`
		CurrentBalance   int64              `json:"currentBalance"`
		Mask             string             `json:"mask"`
		Name             string             `json:"name"`
		Status           BankAccountStatus  `json:"status"`
		Type             BankAccountType    `json:"accountType"`
		SubType          BankAccountSubType `json:"accountSubType"`
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
		existingBankAccount.Name, err = c.cleanString(ctx, "Name", request.Name)
		if err != nil {
			return err
		}

		// TODO Should mask be enforced to have a max of 4 characters?
		existingBankAccount.Mask, err = c.cleanString(ctx, "Mask", request.Mask)
		if err != nil {
			return err
		}
		existingBankAccount.Status = ParseBankAccountStatus(request.Status)
		existingBankAccount.Type = ParseBankAccountType(request.Type)
		existingBankAccount.SubType = ParseBankAccountSubType(request.SubType)
	}

	if err = repo.UpdateBankAccount(c.getContext(ctx), existingBankAccount); err != nil {
		return c.wrapPgError(ctx, err, "failed to update bank account")
	}

	return ctx.JSON(http.StatusOK, *existingBankAccount)
}
