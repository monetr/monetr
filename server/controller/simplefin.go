package controller

import (
	"net/http"
	"net/url"

	"github.com/labstack/echo/v5"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/currency"
	"github.com/monetr/monetr/server/datasources/simplefin"
	"github.com/monetr/monetr/server/internal/myownsanity"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/schemas"
	"github.com/pkg/errors"
)

func (c *Controller) getSimpleFinLinks(ctx *echo.Context) error {
	repo := c.mustGetAuthenticatedRepository(ctx)
	links, err := repo.GetSimpleFinLinks(c.getContext(ctx))
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve SimpleFIN links")
	}

	return ctx.JSON(http.StatusOK, links)
}

func (c *Controller) getSimpleFinLink(ctx *echo.Context) error {
	id, err := ParseID[SimpleFinLink](ctx.Param("simpleFinLinkId"))
	if err != nil || id.IsZero() {
		return c.badRequest(ctx, "Must specify a valid SimpleFIN Link Id to retrieve")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	link, err := repo.GetSimpleFinLink(c.getContext(ctx), id)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve SimpleFIN link")
	}

	return ctx.JSON(http.StatusOK, link)
}

func (c *Controller) postSimpleFinLink(ctx *echo.Context) error {
	request, err := parse(
		c,
		ctx,
		new(schemas.PostSimpleFinLinkRequest),
		schemas.PostSimpleFinLink,
	)
	if err != nil {
		return err
	}

	accessUrl, err := simplefin.ClaimSetupToken(
		c.getContext(ctx),
		c.getLog(ctx),
		request.SetupToken,
		c.Configuration.SimpleFin,
	)
	switch errors.Cause(err) {
	case nil:
		// Do nothing
	case simplefin.ErrInvalidSetupToken:
		return c.badRequest(ctx, "Invalid SimpleFIN setup token")
	case simplefin.ErrHostNotAllowed:
		return c.badRequest(ctx, "SimpleFIN setup token is for a server that is not allowed")
	case simplefin.ErrSetupTokenClaimed:
		return c.badRequest(ctx, "SimpleFIN setup token has already been used")
	default:
		// Don't send err.Error() back here, if the claim failed then it can have
		// the raw response body from SimpleFIN in it.
		return c.wrapAndReturnError(
			ctx,
			err,
			http.StatusInternalServerError,
			"Failed to claim SimpleFIN setup token",
		)
	}

	// The setup token is spent at this point. If anything below fails then the
	// transaction gets rolled back and the access URL is gone, so the user will
	// need to make a new setup token.

	// Keep a copy of the URL without the credentials so we can show it to the
	// user. The client already made sure this parses.
	apiUrl, err := url.Parse(accessUrl)
	if err != nil {
		// Don't include the parse error, it would have the credentials in it.
		return c.returnError(
			ctx,
			http.StatusInternalServerError,
			"Failed to parse SimpleFIN access URL, you will need to create a new setup token",
		)
	}
	apiUrl.User = nil

	secret := repository.SecretData{
		Kind:  SecretKindSimpleFin,
		Value: accessUrl,
	}

	{ // Store the secret and generate an ID
		secrets := c.mustGetSecretsRepository(ctx)
		if err := secrets.Store(c.getContext(ctx), &secret); err != nil {
			return c.wrapPgError(ctx, err, "Failed to store SimpleFIN link, you will need to create a new setup token")
		}
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	simpleFinLink := SimpleFinLink{
		Name:     request.Name,
		SecretId: secret.SecretId,
		ApiUrl:   apiUrl.String(),
		Status:   SimpleFinLinkStatusPending,
	}

	if err := repo.CreateSimpleFinLink(
		c.getContext(ctx),
		&simpleFinLink,
	); err != nil {
		return c.wrapPgError(ctx, err, "Failed to create SimpleFIN link, you will need to create a new setup token")
	}

	return ctx.JSON(http.StatusOK, simpleFinLink)
}

// postSimpleFinLinkBankAccountsRefresh takes a SimpleFIN link ID and reconciles
// the accounts SimpleFIN can see with the ones we store locally. Same as Lunch
// Flow it will add new accounts but never remove any. Every call here counts
// against the user's daily SimpleFIN quota, so it only asks for balances.
func (c *Controller) postSimpleFinLinkBankAccountsRefresh(ctx *echo.Context) error {
	linkId, err := ParseID[SimpleFinLink](ctx.Param("simpleFinLinkId"))
	if err != nil || linkId.IsZero() {
		return c.badRequest(ctx, "Must specify a valid SimpleFIN Link Id to retrieve")
	}

	log := c.getLog(ctx)

	repo := c.mustGetAuthenticatedRepository(ctx)
	link, err := repo.GetSimpleFinLink(c.getContext(ctx), linkId)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve SimpleFIN link")
	}

	secretsRepo := c.mustGetSecretsRepository(ctx)
	secret, err := secretsRepo.Read(c.getContext(ctx), link.SecretId)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve SimpleFIN link secret")
	}

	client, err := simplefin.NewSimpleFinClient(
		log,
		secret.Value,
		c.Configuration.SimpleFin,
	)
	if err != nil {
		return c.wrapAndReturnError(
			ctx,
			err,
			http.StatusInternalServerError,
			"Failed to create SimpleFIN API client",
		)
	}

	// Refresh does not change the status of the link even when SimpleFIN says
	// payment is required or access was denied, that is up to the sync job.
	accountSet, err := client.GetAccounts(
		c.getContext(ctx),
		simplefin.GetAccountsOptions{
			BalancesOnly: true,
		},
	)
	switch errors.Cause(err) {
	case nil:
		// Do nothing
	case simplefin.ErrPaymentRequired:
		// SimpleFIN sends a 402 here but we use 402 for our own billing, so don't
		// pass that status along or the UI might think it's monetr that needs to
		// be paid.
		return c.badRequest(ctx, "SimpleFIN requires payment, check your SimpleFIN account")
	case simplefin.ErrForbidden:
		return c.badRequest(ctx, "SimpleFIN access was denied, the token may have been revoked")
	default:
		// Like Lunch Flow this sends the actual error back to the caller. The
		// client never puts the access URL credentials in its errors so this is
		// safe to do here.
		return c.wrapAndReturnError(
			ctx,
			err,
			http.StatusInternalServerError,
			"Failed to retrieve SimpleFIN accounts: %s",
			err.Error(),
		)
	}

	// The deprecated errors field is where the Bridge puts things like quota
	// warnings, so we want those to show up in the logs.
	if len(accountSet.Errors) > 0 {
		log.WarnContext(c.getContext(ctx), "SimpleFIN returned errors with accounts",
			"errors", accountSet.Errors,
		)
	}

	// Always replace the errors on the link, even when there aren't any, so old
	// errors don't stick around after they've been resolved.
	link.LastErrors = myownsanity.Map(
		accountSet.Errlist,
		func(item simplefin.Error) SimpleFinLinkError {
			return SimpleFinLinkError{
				Code:         item.Code,
				Message:      item.Message,
				ConnectionId: item.ConnectionId,
				SimpleFinId:  item.AccountId,
			}
		},
	)
	if err := repo.UpdateSimpleFinLink(c.getContext(ctx), link); err != nil {
		return c.wrapPgError(ctx, err, "Failed to update SimpleFIN link")
	}

	// Accounts only have the connection ID on them, the institution name lives
	// on the connection
	institutionNames := map[string]string{}
	for _, connection := range accountSet.Connections {
		institutionNames[connection.ConnectionId] = connection.Name
	}

	simpleFinAccounts, err := repo.GetAllSimpleFinBankAccountsBySimpleFinLink(
		c.getContext(ctx),
		link.SimpleFinLinkId,
	)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve stored SimpleFIN accounts")
	}

	// The SimpleFIN account ID is only unique within its connection, so we need
	// to match on both
	for _, joined := range myownsanity.LeftJoin(
		accountSet.Accounts,
		simpleFinAccounts,
		func(external simplefin.Account, internal SimpleFinBankAccount) bool {
			return external.Id == internal.SimpleFinId &&
				external.ConnectionId == internal.ConnectionId
		},
	) {
		if len(joined.Join) == 0 {
			log.InfoContext(c.getContext(ctx), "found SimpleFIN account with no record in monetr, creating",
				"simpleFinId", joined.From.Id,
			)

			// SimpleFIN allows the currency to be a URL for custom currencies, we
			// can't do anything with those so just skip the account.
			balanceCurrency, err := currency.GetCurrency("en", joined.From.Currency)
			if err != nil {
				log.WarnContext(c.getContext(ctx), "SimpleFIN account has an unsupported currency, it will be skipped",
					"simpleFinId", joined.From.Id,
					"currency", joined.From.Currency,
					"err", err,
				)
				continue
			}

			// The balances are JSON numbers so they always use . for the decimal
			// and never have grouping, which is what en gives us
			currentBalance, err := currency.ParseCurrency(
				joined.From.Balance.String(),
				balanceCurrency,
			)
			if err != nil {
				log.WarnContext(c.getContext(ctx), "failed to parse SimpleFIN account balance, will default to 0",
					"simpleFinId", joined.From.Id,
					"err", err,
				)
				currentBalance = 0
			}

			// Available balance is optional, if it isn't there then just use the
			// current balance
			availableBalance := currentBalance
			if joined.From.AvailableBalance != nil {
				availableBalance, err = currency.ParseCurrency(
					joined.From.AvailableBalance.String(),
					balanceCurrency,
				)
				if err != nil {
					log.WarnContext(c.getContext(ctx), "failed to parse SimpleFIN account available balance, will default to 0",
						"simpleFinId", joined.From.Id,
						"err", err,
					)
					availableBalance = 0
				}
			}

			// The institution name can't be empty or the insert will fail, so fall
			// back to the account name if we don't know the connection.
			institutionName := myownsanity.CoalesceStrings(
				institutionNames[joined.From.ConnectionId],
				joined.From.Name,
			)

			if err := repo.CreateSimpleFinBankAccount(
				c.getContext(ctx),
				&SimpleFinBankAccount{
					SimpleFinLinkId:  link.SimpleFinLinkId,
					SimpleFinId:      joined.From.Id,
					ConnectionId:     joined.From.ConnectionId,
					Name:             joined.From.Name,
					InstitutionName:  institutionName,
					Currency:         joined.From.Currency,
					Status:           SimpleFinBankAccountStatusInactive,
					CurrentBalance:   currentBalance,
					AvailableBalance: availableBalance,
				},
			); err != nil {
				return c.wrapPgError(ctx, err, "Failed to create SimpleFIN bank account")
			}
		} else if len(joined.Join) > 1 {
			// Report a bug here! If anyone ever sees this in their logs please know
			// that there is a bug and you should report it via github issues on
			// monetr. You may be asked to provide additional information upon
			// reporting this bug!
			log.ErrorContext(c.getContext(ctx),
				"multiple SimpleFIN bank accounts found for the same external ID, this should not be possible!",
				"bug", true,
				"simpleFinId", joined.From.Id,
				"connectionId", joined.From.ConnectionId,
				"count", len(joined.Join),
			)
			crumbs.IndicateBug(
				c.getContext(ctx),
				"Multiple SimpleFIN Bank Accounts found for the same external ID, this should not be possible!",
				map[string]any{
					"simpleFinId":  joined.From.Id,
					"connectionId": joined.From.ConnectionId,
					"count":        len(joined.Join),
				},
			)
		}

		// Otherwise the account already exists and we are good to go!
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (c *Controller) getSimpleFinLinkBankAccounts(ctx *echo.Context) error {
	id, err := ParseID[SimpleFinLink](ctx.Param("simpleFinLinkId"))
	if err != nil || id.IsZero() {
		return c.badRequest(ctx, "Must specify a valid SimpleFIN Link Id to retrieve")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	link, err := repo.GetSimpleFinLink(c.getContext(ctx), id)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve SimpleFIN link")
	}

	simpleFinAccounts, err := repo.GetSimpleFinBankAccountsBySimpleFinLink(
		c.getContext(ctx),
		link.SimpleFinLinkId,
	)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve SimpleFIN bank accounts")
	}

	return ctx.JSON(http.StatusOK, simpleFinAccounts)
}
