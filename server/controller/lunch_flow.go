package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/consts"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/currency"
	"github.com/monetr/monetr/server/datasources/lunch_flow"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/merge"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
	"github.com/monetr/validation/is"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

func (c *Controller) getLunchFlowLinks(ctx echo.Context) error {
	repo := c.mustGetAuthenticatedRepository(ctx)
	links, err := repo.GetLunchFlowLinks(c.getContext(ctx))
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve Lunch Flow links")
	}

	return ctx.JSON(http.StatusOK, links)
}

func (c *Controller) getLunchFlowLink(ctx echo.Context) error {
	id, err := ParseID[LunchFlowLink](ctx.Param("lunchFlowLinkId"))
	if err != nil || id.IsZero() {
		return c.badRequest(ctx, "Must specify a valid Lunch Flow Link Id to retrieve")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	link, err := repo.GetLunchFlowLink(c.getContext(ctx), id)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve Lunch Flow link")
	}

	return ctx.JSON(http.StatusOK, link)
}

type postLunchFlowLinkRequest struct {
	Name         string `json:"name"`
	LunchFlowURL string `json:"lunchFlowURL"`
	APIKey       string `json:"apiKey"`
}

// parsePostLunchFlowLinkRequest will take an echo request context and the lunch
// flow request object the caller wants the data parsed into. It will parse the
// request body and validate it and return an error if there is one. Or it will
// return nil and update the passed object.
func (c *Controller) parsePostLunchFlowLinkRequest(
	ctx echo.Context, result *postLunchFlowLinkRequest,
) error {
	rawData := map[string]any{}
	decoder := json.NewDecoder(ctx.Request().Body)
	decoder.UseNumber()
	if err := decoder.Decode(&rawData); err != nil {
		return c.invalidJsonError(ctx, err)
	}

	// Validate the request from the client
	if err := validation.ValidateWithContext(
		c.getContext(ctx),
		&rawData,
		validation.Map(
			validators.Name(validators.Require),
			validation.Key(
				"lunchFlowURL",
				validation.Required.Error("Lunch Flow API URL is required to setup a Lunch Flow link"),
				validation.NewStringRule(func(input string) bool {
					parsed, err := url.Parse(input)
					if err != nil {
						return false
					}
					// Do not allow query parameters in the URL as these will be removed
					// when requests are made!
					if len(parsed.Query()) > 0 {
						return false
					}

					// Require a scheme to be specified
					switch strings.ToLower(parsed.Scheme) {
					case "http", "https":
						// These are considered valid!
					default:
						// Any other scheme is not considered valid here!
						return false
					}

					return true
				}, "Lunch Flow API URL must be a full valid URL"),
			).Required(validators.Require),
			validation.Key(
				"apiKey",
				validation.Required.Error("Lunch Flow API Key must be provided to setup a Lunch Flow link"),
				validation.Length(1, 100).Error("Lunch Flow API Key must be between 1 and 100 characters"),
				is.UTFLetterNumeric,
			).Required(validators.Require),
		),
	); err != nil {
		return err
	}

	// Then merge the data into our request struct!
	if err := merge.Merge(
		result, rawData, merge.ErrorOnUnknownField,
	); err != nil {
		return errors.Wrap(err, "failed to merge request data")
	}

	return nil
}

func (c *Controller) postLunchFlowLink(ctx echo.Context) error {
	var request postLunchFlowLinkRequest
	err := c.parsePostLunchFlowLinkRequest(
		ctx,
		&request,
	)
	switch errors.Cause(err).(type) {
	case validation.Errors:
		return ctx.JSON(http.StatusBadRequest, map[string]any{
			"error":    "Invalid request",
			"problems": err,
		})
	case *json.SyntaxError:
		return c.invalidJsonError(ctx, err)
	case nil:
		break
	default:
		return c.badRequestError(ctx, err, "failed to parse post request")
	}

	secret := repository.SecretData{
		Kind:  SecretKindLunchFlow,
		Value: request.APIKey,
	}

	{ // Store the secret and generate an ID
		secrets := c.mustGetSecretsRepository(ctx)
		if err := secrets.Store(c.getContext(ctx), &secret); err != nil {
			return c.wrapPgError(ctx, err, "Failed to store Lunch Flow secret")
		}
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	lunchFlowLink := LunchFlowLink{
		Name:      request.Name,
		SecretId:  secret.SecretId,
		ApiUrl:    request.LunchFlowURL,
		Status:    LunchFlowLinkStatusPending,
		CreatedBy: c.mustGetUserId(ctx),
	}

	// The Lunch Flow link itself needs to be created first.
	if err := repo.CreateLunchFlowLink(
		c.getContext(ctx),
		&lunchFlowLink,
	); err != nil {
		return c.wrapPgError(ctx, err, "Failed to create Lunch Flow link")
	}

	return ctx.JSON(http.StatusOK, lunchFlowLink)
}

// postLunchFlowLinkBankAccountsRefresh is the endpoint that takes a Lunch Flow
// link ID and performs a reconciliation of the accounts available in the API
// versus the ones we store locally. It will not remove local items but it will
// add new ones if they become available. It does not return content but should
// be called during the setup process to fetch accounts and validate that the
// API is working properly.
func (c *Controller) postLunchFlowLinkBankAccountsRefresh(ctx echo.Context) error {
	linkId, err := ParseID[LunchFlowLink](ctx.Param("lunchFlowLinkId"))
	if err != nil || linkId.IsZero() {
		return c.badRequest(ctx, "Must specify a valid Lunch Flow Link Id to retrieve")
	}

	log := c.getLog(ctx)

	repo := c.mustGetAuthenticatedRepository(ctx)
	link, err := repo.GetLunchFlowLink(c.getContext(ctx), linkId)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve Lunch Flow link")
	}

	secretsRepo := c.mustGetSecretsRepository(ctx)
	secret, err := secretsRepo.Read(c.getContext(ctx), link.SecretId)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve Lunch Flow link secret")
	}

	client, err := lunch_flow.NewLunchFlowClient(
		log,
		link.ApiUrl,
		secret.Value,
	)
	if err != nil {
		return c.wrapAndReturnError(
			ctx,
			err,
			http.StatusInternalServerError,
			"Failed to create Lunch Flow API client",
		)
	}

	externalAccounts, err := client.GetAccounts(c.getContext(ctx))
	if err != nil {
		// TODO Should we expose actual error information here to the frontend to
		// make it so that the user does not need to check server logs to debug any
		// issues? This is for self-hosted instances only so it may be worth it?
		return c.wrapAndReturnError(
			ctx,
			err,
			http.StatusInternalServerError,
			"Failed to retrieve accounts from Lunch Flow",
		)
	}

	lunchFlowAccounts, err := repo.GetLunchFlowBankAccountsByLunchFlowLink(
		c.getContext(ctx),
		link.LunchFlowLinkId,
	)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve stored Lunch Flow accounts")
	}

	// Join the stored accounts against what we get from the API
	for _, joined := range myownsanity.LeftJoin(
		externalAccounts,
		lunchFlowAccounts,
		func(external lunch_flow.Account, internal LunchFlowBankAccount) bool {
			return external.Id == lunch_flow.AccountId(internal.LunchFlowId)
		},
	) {
		if len(joined.Join) == 0 {
			log.Info("Found Lunch Flow account with no record in monetr, creating")

			log.Debug("requesting balance information first!")

			var currentBalance int64
			var currencyCode string
			balance, err := client.GetBalance(c.getContext(ctx), joined.From.Id)
			if err != nil {
				log.WithError(err).Warn("failed to retrieve balance information for Lunch Flow account")
			} else {
				currencyCode = myownsanity.CoalesceStrings(
					balance.Currency,
					joined.From.Currency,
					consts.DefaultCurrencyCode,
				)
				currentBalance, err = currency.ParseCurrency(
					balance.Amount.String(),
					currencyCode,
				)
				if err != nil {
					log.WithError(err).Warn("failed to parse account balance, will default to 0")
				}
			}

			if err := repo.CreateLunchFlowBankAccount(
				c.getContext(ctx),
				&LunchFlowBankAccount{
					LunchFlowLinkId: linkId,
					LunchFlowId:     joined.From.Id.String(),
					LunchFlowStatus: LunchFlowBankAccountExternalStatus(joined.From.Status),
					Name:            joined.From.Name,
					InstitutionName: joined.From.InstitutionName,
					Provider:        joined.From.Provider,
					Currency: myownsanity.CoalesceStrings(
						balance.Currency,
						joined.From.Currency,
						consts.DefaultCurrencyCode,
					),
					Status:         LunchFlowBankAccountStatusInactive,
					CurrentBalance: currentBalance,
					CreatedBy:      c.mustGetUserId(ctx),
				},
			); err != nil {
				return c.wrapPgError(ctx, err, "Failed to Lunch Flow bank account")
			}
		} else if len(joined.Join) > 1 {
			// Report a bug here! If anyone ever sees this in their logs please know
			// that there is a bug and you should report it via github issues on
			// monetr. You may be asked to provide additional information upon
			// reporting this bug!
			log.WithFields(logrus.Fields{
				"bug":         true,
				"lunchFlowId": joined.From.Id,
				"count":       len(joined.Join),
			}).Error("multiple Lunch Flow bank accounts found for the same external ID, this should not be possible!")
			crumbs.IndicateBug(
				c.getContext(ctx),
				"Multiple Lunch Flow Bank Accounts found for the same external ID, this should not be possible!",
				map[string]any{
					"lunchFlowId": joined.From.Id,
					"count":       len(joined.Join),
				},
			)
		}

		// Otherwise the account already exists and we are good to go!
	}

	// Return no content to indicate success!
	return ctx.NoContent(http.StatusNoContent)
}

func (c *Controller) getLunchFlowLinkBankAccounts(ctx echo.Context) error {
	id, err := ParseID[LunchFlowLink](ctx.Param("lunchFlowLinkId"))
	if err != nil || id.IsZero() {
		return c.badRequest(ctx, "Must specify a valid Lunch Flow Link Id to retrieve")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	link, err := repo.GetLunchFlowLink(c.getContext(ctx), id)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve Lunch Flow link")
	}

	lunchFlowAccounts, err := repo.GetLunchFlowBankAccountsByLunchFlowLink(
		c.getContext(ctx),
		link.LunchFlowLinkId,
	)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve Lunch Flow bank accounts")
	}

	return ctx.JSON(http.StatusOK, lunchFlowAccounts)
}

func (c *Controller) postLunchFlowLinkSync(ctx echo.Context) error {
	var request struct {
		LinkId ID[Link] `json:"linkId"`
	}
	if err := ctx.Bind(&request); err != nil {
		return c.invalidJson(ctx)
	}

	log := c.getLog(ctx).WithFields(logrus.Fields{
		"linkId": request.LinkId,
	})

	repo := c.mustGetAuthenticatedRepository(ctx)
	link, err := repo.GetLink(c.getContext(ctx), request.LinkId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve link")
	}

	if link.LinkType != LunchFlowLinkType {
		return c.badRequest(ctx, "cannot manually sync a non-Lunch Flow link")
	}

	switch link.LunchFlowLink.Status {
	case LunchFlowLinkStatusActive, LunchFlowLinkStatusError:
		log.WithField("status", link.LunchFlowLink.Status).Debug("link is not deactivated, triggering manual sync")
	case LunchFlowLinkStatusDeactivated:
		return c.badRequest(ctx, "Link is not active and will not be synced")
	}

	lunchFlowLink := link.LunchFlowLink
	if lastManualSync := lunchFlowLink.LastManualSync; lastManualSync != nil && lastManualSync.After(c.Clock.Now().Add(-30*time.Minute)) {
		return c.returnError(ctx, http.StatusTooEarly, "Link has been manually synced too recently")
	}

	lunchFlowLink.LastManualSync = myownsanity.Pointer(c.Clock.Now())
	if err := repo.UpdateLunchFlowLink(c.getContext(ctx), lunchFlowLink); err != nil {
		return c.wrapPgError(ctx, err, "could not manually sync link")
	}

	bankAccounts, err := repo.GetBankAccountsByLinkId(
		c.getContext(ctx),
		link.LinkId,
	)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve bank accounts to sync")
	}

	// Get the database instance from the request context, this should be a
	// database transaction for this endpoint.
	db := c.mustGetDatabase(ctx)
	for _, bankAccount := range bankAccounts {
		log.WithFields(logrus.Fields{
			"bankAccountId": bankAccount.BankAccountId,
		}).Debug("triggering Lunch Flow sync for bank account")
		if err := background.TriggerSyncLunchFlowTxn(
			c.getContext(ctx),
			c.JobRunner,
			db,
			background.SyncLunchFlowArguments{
				AccountId:     bankAccount.AccountId,
				BankAccountId: bankAccount.BankAccountId,
				LinkId:        bankAccount.LinkId,
			},
		); err != nil {
			log.WithFields(logrus.Fields{
				"bankAccountId": bankAccount.BankAccountId,
			}).WithError(err).Warn("failed to enqueue Lunch Flow sync")
		}
	}

	return ctx.NoContent(http.StatusAccepted)
}

func (c *Controller) getLunchFlowLinkSyncProgress(ctx echo.Context) error {
	linkId, err := ParseID[Link](ctx.Param("linkId"))
	if err != nil || linkId.IsZero() {
		return c.badRequest(ctx, "must specify a valid link Id")
	}

	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	log := c.getLog(ctx)

	repo := c.mustGetAuthenticatedRepository(ctx)

	link, err := repo.GetLink(c.getContext(ctx), linkId)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve link")
	}

	if link.LinkType != LunchFlowLinkType {
		return c.badRequest(ctx, "Link must be a Lunch Flow link type")
	}

	channel := fmt.Sprintf(
		"account:%s:link:%s:bank_account:%s:lunch_flow_sync_progress",
		c.mustGetAccountId(ctx), linkId, bankAccountId,
	)
	listener, err := c.PubSub.Subscribe(
		c.getContext(ctx),
		c.mustGetAccountId(ctx),
		channel,
	)
	if err != nil {
		return c.wrapAndReturnError(
			ctx,
			err,
			http.StatusInternalServerError,
			"Failed to subscribe to Lunch Flow sync progress",
		)
	}
	defer listener.Close()

	timeout := time.NewTimer(5 * time.Minute)
	defer timeout.Stop()

	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		log.Debug("started Lunch Flow sync websocket!")
	ListenerLoop:
		for {
			select {
			case <-timeout.C:
				log.Warn("exiting Lunch Flow sync progress due to timeout")
				break ListenerLoop
			case notification := <-listener.Channel():
				var status struct {
					BankAccountId ID[BankAccount]                `json:"bankAccountId"`
					Status        background.LunchFlowSyncStatus `json:"status"`
				}
				if err := json.Unmarshal([]byte(notification.Payload()), &status); err != nil {
					log.WithError(err).Warn("failed to unmarshal notification for websocket!")
					continue
				}
				log.WithField("notification", notification).Debug("received Lunch Flow sync progress notification")
				if err := c.sendWebsocketMessage(
					ctx,
					ws,
					status,
				); err != nil {
					log.WithError(err).Warn("failed to send websocket message")
					return
				}

				if status.Status == background.LunchFlowSyncStatusComplete ||
					status.Status == background.LunchFlowSyncStatusError {
					log.Info("received terminal status for Lunch Flow sync, exiting socket")
					break ListenerLoop
				}
			}
		}

		log.Debug("finished listening for Lunch Flow sync updates")
	}).ServeHTTP(ctx.Response(), ctx.Request())

	return nil
}
