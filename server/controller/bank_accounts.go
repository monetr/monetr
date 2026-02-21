package controller

import (
	"bytes"
	"net/http"

	locale "github.com/elliotcourant/go-lclocale"
	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/consts"
	"github.com/monetr/monetr/server/internal/myownsanity"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/validation"
	"github.com/sirupsen/logrus"
)

func (c *Controller) getBankAccounts(ctx echo.Context) error {
	repo := c.mustGetAuthenticatedRepository(ctx)
	var err error
	var bankAccounts []BankAccount

	// If the client is filtering by Link ID then use this query instead. As it
	// will not exclude deleted bank accounts which is important for the link
	// details view.
	if linkId := ctx.QueryParam("link_id"); linkId != "" {
		bankAccounts, err = repo.GetBankAccountsByLinkId(
			c.getContext(ctx),
			ID[Link](linkId),
		)
	} else {
		bankAccounts, err = repo.GetBankAccounts(c.getContext(ctx))
	}

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
	log := c.getLog(ctx)
	repo := c.mustGetAuthenticatedRepository(ctx)

	account, err := c.Accounts.GetAccount(
		c.getContext(ctx),
		c.mustGetAccountId(ctx),
	)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to read account details")
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

	var bankAccount BankAccount
	// Pre-set the default values for these fields.
	bankAccount.Status = ActiveBankAccountStatus
	bankAccount.AccountType = DepositoryBankAccountType
	bankAccount.AccountSubType = CheckingBankAccountSubType
	bankAccount.Currency = currencyCode

	switch err := bankAccount.UnmarshalRequest(
		c.getContext(ctx),
		ctx.Request().Body,
		bankAccount.CreateValidators()...,
	).(type) {
	case validation.Errors:
		return ctx.JSON(http.StatusBadRequest, map[string]any{
			"error":    "Invalid request",
			"problems": err,
		})
	case nil:
		break
	default:
		return c.wrapAndReturnError(
			ctx,
			err,
			http.StatusBadRequest,
			"failed to parse request",
		)
	}

	// Some fields cannot be overwritten, so we set those after we unmarshal.
	bankAccount.LastUpdated = c.Clock.Now().UTC()

	// Bank accounts can only be created via the API on manual links or on Lunch
	// Flow or other simple integration links. Validate this before proceeding.
	link, err := repo.GetLink(c.getContext(ctx), bankAccount.LinkId)
	if err != nil {
		return c.wrapPgError(ctx, err, "Could not validate link allows bank account creation")
	}

	// If we are a lunch flow link then we can only create lunch flow bank
	// accounts!
	switch link.LinkType {
	case LunchFlowLinkType:
		if bankAccount.LunchFlowBankAccountId == nil ||
			bankAccount.LunchFlowBankAccountId.IsZero() {
			return ctx.JSON(http.StatusBadRequest, map[string]any{
				"error": "Invalid request",
				"problems": map[string]any{
					"lunchFlowBankAccountId": "Lunch Flow Bank Account ID required to create a bank account for this link",
				},
			})
		}

		lunchFlowBankAccount, err := repo.GetLunchFlowBankAccount(
			c.getContext(ctx),
			*bankAccount.LunchFlowBankAccountId,
		)
		if err != nil {
			return c.wrapPgError(ctx, err, "Failed to retrieve Lunch Flow bank account")
		}
		lunchFlowBankAccount.Status = LunchFlowBankAccountStatusActive
		if err := repo.UpdateLunchFlowBankAccount(
			c.getContext(ctx),
			lunchFlowBankAccount,
		); err != nil {
			return c.wrapPgError(ctx, err, "Failed to update Lunch Flow bank account")
		}
	case ManualLinkType:
	default:
		// Otherwise if we are not a manual link then we simply don't allow bank
		// accounts to be created.
		return ctx.JSON(http.StatusBadRequest, map[string]any{
			"error": "Invalid request",
			"problems": map[string]any{
				"linkId": "Cannot create a bank account for a non-manual link, specify a manual Link ID",
			},
		})
	}

	if err := repo.CreateBankAccounts(
		c.getContext(ctx),
		&bankAccount,
	); err != nil {
		return c.wrapPgError(ctx, err, "Could not create bank account")
	}

	return ctx.JSON(http.StatusOK, bankAccount)
}

func (c *Controller) patchBankAccount(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	existingBankAccount, err := repo.GetBankAccount(c.getContext(ctx), bankAccountId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve bank account")
	}

	switch err := existingBankAccount.UnmarshalRequest(
		c.getContext(ctx),
		ctx.Request().Body,
		existingBankAccount.UpdateValidator()...,
	).(type) {
	case validation.Errors:
		return ctx.JSON(http.StatusBadRequest, map[string]any{
			"error":    "Invalid request",
			"problems": err,
		})
	case nil:
		break
	default:
		return c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "failed to parse patch request")
	}

	if err = repo.UpdateBankAccount(
		c.getContext(ctx),
		existingBankAccount,
	); err != nil {
		return c.wrapPgError(ctx, err, "failed to update bank account")
	}

	return ctx.JSON(http.StatusOK, *existingBankAccount)
}

func (c *Controller) putBankAccount(ctx echo.Context) error {
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
			// AvailableBalance int64   `json:"availableBalance"`
			// CurrentBalance   int64   `json:"currentBalance"`
			// LimitBalance     int64   `json:"limitBalance"`
			// Mask     string  `json:"mask"`
			Name     string  `json:"name"`
			Currency *string `json:"currency"`
			// Status           BankAccountStatus  `json:"status"`
			// Type             BankAccountType    `json:"accountType"`
			// SubType          BankAccountSubType `json:"accountSubType"`
		}
		if err = ctx.Bind(&request); err != nil {
			return c.invalidJson(ctx)
		}

		err = validation.ValidateStructWithContext(c.getContext(ctx), &request,
			// validation.Field(
			// 	&request.Mask,
			// 	validation.Match(regexp.MustCompile(`\d{4}`)).Error("Mask must be a 4 digit string"),
			// ),
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
			// validation.Field(
			// 	&request.LimitBalance,
			// 	validation.Min(0).Error("Limit balance cannot be negative"),
			// ),
			// validation.Field(
			// 	&request.Status,
			// 	validation.In(
			// 		ActiveBankAccountStatus,
			// 		InactiveBankAccountStatus,
			// 		UnknownBankAccountStatus,
			// 	).Error("Invalid bank account status"),
			// ),
			// validation.Field(
			// 	&request.Type,
			// 	validation.In(
			// 		DepositoryBankAccountType,
			// 		CreditBankAccountType,
			// 		LoanBankAccountType,
			// 		InvestmentBankAccountType,
			// 		OtherBankAccountType,
			// 	).Error("Invalid bank account type"),
			// ),
			// validation.Field(
			// 	&request.SubType,
			// 	validation.In(
			// 		CheckingBankAccountSubType,
			// 		SavingsBankAccountSubType,
			// 		HSABankAccountSubType,
			// 		CDBankAccountSubType,
			// 		MoneyMarketBankAccountSubType,
			// 		PayPalBankAccountSubType,
			// 		PrepaidBankAccountSubType,
			// 		CashManagementBankAccountSubType,
			// 		EBTBankAccountSubType,
			// 		CreditCardBankAccountSubType,
			// 		AutoBankAccountSubType,
			// 		OtherBankAccountSubType,
			// 	).Error("Invalid bank account sub type"),
			// ),
		)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]any{
				"error":    "Invalid request",
				"problems": err,
			})
		}

		// existingBankAccount.AvailableBalance = request.AvailableBalance
		// existingBankAccount.CurrentBalance = request.CurrentBalance
		// existingBankAccount.LimitBalance = request.LimitBalance
		existingBankAccount.Name = request.Name
		// existingBankAccount.Mask = request.Mask
		// existingBankAccount.Status = request.Status
		// existingBankAccount.Type = request.Type
		// existingBankAccount.SubType = request.SubType
		if request.Currency != nil {
			existingBankAccount.Currency = *request.Currency
		}
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

func (c *Controller) deleteBankAccount(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "Must specify a valid bank account Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	existingBankAccount, err := repo.GetBankAccount(
		c.getContext(ctx),
		bankAccountId,
	)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve bank account")
	}

	if existingBankAccount.PlaidBankAccount != nil {
		return c.badRequest(ctx, "Plaid bank account cannot be removed this way")
	}

	// TODO Handle Lunch flow bank account status here!

	existingBankAccount.DeletedAt = myownsanity.TimeP(c.Clock.Now())
	if err = repo.UpdateBankAccount(
		c.getContext(ctx),
		existingBankAccount,
	); err != nil {
		return c.wrapPgError(ctx, err, "failed to update bank account")
	}

	return ctx.NoContent(http.StatusOK)
}
