package controller

import (
	"bytes"
	"net/http"
	"regexp"
	"strings"

	locale "github.com/elliotcourant/go-lclocale"
	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/consts"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/validation"
	"github.com/sirupsen/logrus"
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

	log := c.getLog(ctx)

	// If the client has not specified a currency code then we should determine
	// the currency code based on the user's locale.
	if bankAccount.Currency == "" {
		account, err := c.Accounts.GetAccount(c.getContext(ctx), c.mustGetAccountId(ctx))
		if err != nil {
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "Failed to read account details")
		}

		// If we cannot determine what currencyCode we should default to based on the
		// locale, then fallback to monetr's global default.
		currencyCode := consts.DefaultCurrencyCode
		// Try to retrieve currency information for the user's locale from the
		// operating system.
		lconv, err := locale.GetLConv(account.Locale)
		if err != nil || lconv == nil {
			log.
				WithFields(logrus.Fields{
					"locale": account.Locale,
				}).
				WithError(err).
				Warn("failed to get currency information for account's locale, application default currency will be used")
		} else {
			// If it worked then clean up the code from the OS and use it.
			currencyCode = string(bytes.TrimSpace(lconv.IntCurrSymbol))
		}

		// Set the bank account's currency based on the user's locale.
		bankAccount.Currency = currencyCode
	} else {
		// Clean up the currency code the user provided.
		currencyCode := strings.ToUpper(strings.TrimSpace(bankAccount.Currency))
		// Check to see if the system we are on supports that currency code by
		// checking if there is fractional digit information about it.
		if _, err := locale.GetCurrencyInternationalFractionalDigits(currencyCode); err != nil {
			log.
				WithFields(logrus.Fields{
					"input":    bankAccount.Currency,
					"currency": currencyCode,
				}).
				WithError(err).
				Warn("could not find currency information for the specified currency code")
			return c.badRequest(ctx, "Provided currency code is not valid")
		}
		// If the currency code specified by the client is valid then use that code
		// for the account.
		bankAccount.Currency = currencyCode
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

	// TODO Eventually we should just query the link to see if its a manual link.
	// But for now if the bank account has a plaid account ID then its probably
	// safe to assume that it is a Plaid managed bank account.
	if existingBankAccount.PlaidBankAccountId == nil {
		var request struct {
			AvailableBalance int64              `json:"availableBalance"`
			CurrentBalance   int64              `json:"currentBalance"`
			LimitBalance     int64              `json:"limitBalance"`
			Mask             string             `json:"mask"`
			Name             string             `json:"name"`
			Currency         *string            `json:"currency"`
			Status           BankAccountStatus  `json:"status"`
			Type             BankAccountType    `json:"accountType"`
			SubType          BankAccountSubType `json:"accountSubType"`
		}
		if err = ctx.Bind(&request); err != nil {
			return c.invalidJson(ctx)
		}

		err = validation.ValidateStructWithContext(c.getContext(ctx), &request,
			validation.Field(
				&request.Mask,
				validation.Match(regexp.MustCompile(`\d{4}`)).Error("Mask must be a 4 digit string"),
			),
			validation.Field(
				&request.Name,
				validation.Length(1, 300).Error("Name must be between 1 and 300 characters"),
			),
			validation.Field(
				&request.Currency,
				validation.In(
					locale.GetInstalledCurrencies()...,
				).Error("Currency must be one supported by the server"),
			),
			validation.Field(
				&request.LimitBalance,
				validation.Min(0).Error("Limit balance cannot be negative"),
			),
			validation.Field(
				&request.Status,
				validation.In(
					ActiveBankAccountStatus,
					InactiveBankAccountStatus,
					UnknownBankAccountStatus,
				).Error("Invalid bank account status"),
			),
			validation.Field(
				&request.Type,
				validation.In(
					DepositoryBankAccountType,
					CreditBankAccountType,
					LoanBankAccountType,
					InvestmentBankAccountType,
					OtherBankAccountType,
				).Error("Invalid bank account type"),
			),
			validation.Field(
				&request.SubType,
				validation.In(
					CheckingBankAccountSubType,
					SavingsBankAccountSubType,
					HSABankAccountSubType,
					CDBankAccountSubType,
					MoneyMarketBankAccountSubType,
					PayPalBankAccountSubType,
					PrepaidBankAccountSubType,
					CashManagementBankAccountSubType,
					EBTBankAccountSubType,
					CreditCardBankAccountSubType,
					AutoBankAccountSubType,
					OtherBankAccountSubType,
				).Error("Invalid bank account sub type"),
			),
		)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]any{
				"error":    "Invalid request",
				"problems": err,
			})
		}

		existingBankAccount.AvailableBalance = request.AvailableBalance
		existingBankAccount.CurrentBalance = request.CurrentBalance
		existingBankAccount.LimitBalance = request.LimitBalance
		existingBankAccount.Name = request.Name
		existingBankAccount.Mask = request.Mask
		existingBankAccount.Status = request.Status
		existingBankAccount.Type = request.Type
		existingBankAccount.SubType = request.SubType
	} else {
		var request struct {
			Name string `json:"name"`
		}
		if err = ctx.Bind(&request); err != nil {
			return c.invalidJson(ctx)
		}

		err = validation.ValidateStructWithContext(c.getContext(ctx), &request,
			validation.Field(
				&request.Name,
				validation.Length(1, 300).Error("Name must be between 1 and 300 characters"),
			),
		)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]any{
				"error":    "Invalid request",
				"problems": err,
			})
		}

		existingBankAccount.Name = request.Name
	}

	if err = repo.UpdateBankAccount(
		c.getContext(ctx),
		existingBankAccount,
	); err != nil {
		return c.wrapPgError(ctx, err, "failed to update bank account")
	}

	return ctx.JSON(http.StatusOK, *existingBankAccount)
}
