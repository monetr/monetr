package controller

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/pkg/background"
	"github.com/monetr/monetr/pkg/consts"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/internal/myownsanity"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/platypus"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (c *Controller) storeLinkTokenInCache(
	ctx context.Context,
	userId uint64,
	linkId uint64,
	linkToken string,
	expiration time.Time,
) error {
	span := sentry.StartSpan(ctx, "StoreLinkTokenInCache")
	defer span.Finish()

	key := fmt.Sprintf("plaid:in_progress:%d:%d", userId, linkId)
	return errors.Wrap(
		c.cache.SetEzTTL(span.Context(), key, linkToken, expiration.Sub(time.Now())),
		"failed to cache link token",
	)
}

func (c *Controller) checkCacheForLinkToken(
	ctx context.Context,
	userId uint64,
	linkId uint64,
) (string, error) {
	span := sentry.StartSpan(ctx, "StoreLinkTokenInCache")
	defer span.Finish()

	key := fmt.Sprintf("plaid:in_progress:%d:%d", userId, linkId)
	var token string
	if err := c.cache.GetEz(span.Context(), key, &token); err != nil {
		return "", errors.Wrap(err, "failed to retrieve cached link token")
	}
	return token, nil
}

func (c *Controller) removeLinkTokenFromCache(
	ctx context.Context,
	userId uint64,
	linkId uint64,
) error {
	span := sentry.StartSpan(ctx, "RemoteLinkTokenFromCache")
	defer span.Finish()

	key := fmt.Sprintf("plaid:in_progress:%d:%d", userId, linkId)
	return errors.Wrap(
		c.cache.Delete(span.Context(), key),
		"failed to remove cached link token",
	)
}

func (c *Controller) newPlaidToken(ctx echo.Context) error {
	repo := c.mustGetAuthenticatedRepository(ctx)

	// Retrieve the user's details. We need to pass some of these along to
	// plaid as part of the linking process.
	me, err := repo.GetMe(c.getContext(ctx))
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to get user details for link")
	}

	if !c.configuration.Plaid.Enabled {
		return c.returnError(ctx, http.StatusNotAcceptable, "Plaid is not enabled on this server, only manual links are allowed.")
	}

	userId := c.mustGetUserId(ctx)

	log := c.getLog(ctx).WithFields(logrus.Fields{
		"accountId": me.AccountId,
		"userId":    me.UserId,
		"loginId":   me.LoginId,
	})

	numberOfLinks, err := repo.GetNumberOfPlaidLinks(c.getContext(ctx))
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to determine the number of existing plaid links")
	}

	// If there is a configured limit on Plaid links then enforce that limit.
	if maxLinks := c.configuration.Plaid.MaxNumberOfLinks; maxLinks > 0 && numberOfLinks >= maxLinks {
		return c.badRequest(ctx, "max number of Plaid links already reached")
	}

	// If billing is enabled and the current account is trialing, then limit them to a single Plaid link until their
	// trial has expired.
	if c.configuration.Stripe.IsBillingEnabled() {
		trialing, err := c.paywall.GetSubscriptionIsTrialing(c.getContext(ctx), c.mustGetAccountId(ctx))
		if err != nil {
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to determine trial status")
		}

		if trialing && numberOfLinks > 0 {
			log.WithFields(logrus.Fields{
				"numberOfLinks": numberOfLinks,
				"trialing":      trialing,
			}).Debug("cannot add more Plaid links during trial")
			return c.badRequest(ctx, "cannot add additional Plaid links during trial")
		}
	}

	// If we are trying to not send a ton of requests then check the cache to see if we still have a valid link token that
	// we can use.
	if checkCache, err := strconv.ParseBool(ctx.QueryParam("use_cache")); err == nil && checkCache {
		if linkToken, err := c.checkCacheForLinkToken(
			c.getContext(ctx),
			userId,
			0,
		); err == nil && len(linkToken) > 0 {
			log.Info("successfully found existing link token in cache")
			return ctx.JSON(http.StatusOK, map[string]interface{}{
				"linkToken": linkToken,
			})
		}
		log.Info("no link token was found in the cache")
	}

	legalName := ""
	if len(me.LastName) > 0 {
		legalName = fmt.Sprintf("%s %s", me.FirstName, me.LastName)
	}

	var phoneNumber *string
	if me.Login.PhoneNumber != nil {
		phoneNumber = myownsanity.StringP(me.Login.PhoneNumber.E164())
	}

	log.Trace("creating Plaid link token")
	token, err := c.plaid.CreateLinkToken(c.getContext(ctx), platypus.LinkTokenOptions{
		ClientUserID:             strconv.FormatUint(userId, 10),
		LegalName:                legalName,
		PhoneNumber:              phoneNumber,
		PhoneNumberVerifiedTime:  nil,
		EmailAddress:             me.Login.Email,
		EmailAddressVerifiedTime: me.Login.EmailVerifiedAt,
	})
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create link token")
	}

	if err = c.storeLinkTokenInCache(
		c.getContext(ctx),
		me.UserId,
		0, // Since no link exists this should be cached without a link Id.
		token.Token(),
		token.Expiration(),
	); err != nil {
		log.WithError(err).Warn("failed to cache link token")
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"linkToken": token.Token(),
	})
}

func (c *Controller) updatePlaidLink(ctx echo.Context) error {
	if !c.configuration.Plaid.Enabled {
		return c.returnError(ctx, http.StatusNotAcceptable, "Plaid is not enabled on this server, only manual links are allowed.")
	}

	linkId, err := strconv.ParseUint(ctx.Param("linkId"), 10, 64)
	if err != nil || linkId == 0 {
		return c.badRequest(ctx, "must specify a link Id")
	}

	updateAccountSelection := urlParamBoolDefault(ctx, "update_account_selection", false)

	log := c.getLog(ctx).WithField("linkId", linkId)

	// Retrieve the user's details. We need to pass some of these along to plaid as part of the linking process.
	repo := c.mustGetAuthenticatedRepository(ctx)

	link, err := repo.GetLink(c.getContext(ctx), linkId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve link")
	}

	if link.LinkType != models.PlaidLinkType {
		return c.badRequest(ctx, "cannot update a non-Plaid link")
	}

	if link.PlaidLink == nil {
		return c.returnError(ctx, http.StatusInternalServerError, "no Plaid details associated with link")
	}

	me, err := repo.GetMe(c.getContext(ctx))
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve user details")
	}

	client, err := c.plaid.NewClientFromLink(c.getContext(ctx), me.AccountId, linkId)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create Plaid client for link")
	}

	token, err := client.UpdateItem(c.getContext(ctx), updateAccountSelection)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create link token to update Plaid link")
	}

	if err = c.storeLinkTokenInCache(
		c.getContext(ctx),
		me.UserId,
		link.LinkId, // Cache the token under the link ID, that way it is only cached for updated for that link.
		token.Token(),
		token.Expiration(),
	); err != nil {
		log.WithError(err).Warn("failed to cache link token")
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"linkToken": token.Token(),
	})
}

// Token Callback for Plaid Link
// @Summary Updated Token Callback
// @id updated-token-callback
// @tags Plaid
// @Description This is used when handling an update flow for a Plaid link. Rather than returning the public token to the normal callback endpoint, this one should be used instead. This one assumes the link already exists and handles it slightly differently than it would for a new link.
// @Security ApiKeyAuth
// @Produce json
// @Accept json
// @Param Request body swag.UpdatePlaidTokenCallbackRequest true "Update token callback request."
// @Router /plaid/update/callback [post]
// @Success 200 {object} swag.LinkResponse
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) updatePlaidTokenCallback(ctx echo.Context) error {
	if !c.configuration.Plaid.Enabled {
		return c.returnError(ctx, http.StatusNotAcceptable, "Plaid is not enabled on this server, only manual links are allowed.")
	}

	var callbackRequest struct {
		LinkId      uint64   `json:"linkId"`
		PublicToken string   `json:"publicToken"`
		AccountIds  []string `json:"accountIds"`
	}
	if err := ctx.Bind(&callbackRequest); err != nil {
		return c.invalidJson(ctx)
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	link, err := repo.GetLink(c.getContext(ctx), callbackRequest.LinkId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve link")
	}
	log := c.getLog(ctx)

	if err := c.removeLinkTokenFromCache(
		c.getContext(ctx),
		c.mustGetUserId(ctx),
		link.LinkId,
	); err != nil {
		log.WithError(err).Warn("failed to remove link token from cache")
	}

	result, err := c.plaid.ExchangePublicToken(c.getContext(ctx), callbackRequest.PublicToken)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to exchange token")
	}

	currentAccessToken, err := c.plaidSecrets.GetAccessTokenForPlaidLinkId(
		c.getContext(ctx),
		repo.AccountId(),
		link.PlaidLink.ItemId,
	)
	if err != nil {
		log.WithError(err).Warn("failed to retrieve access token for existing plaid link")
	}

	if currentAccessToken != result.AccessToken {
		log.Info("access token for link has been updated")
		if err = c.plaidSecrets.UpdateAccessTokenForPlaidLinkId(
			c.getContext(ctx),
			repo.AccountId(),
			link.PlaidLink.ItemId,
			result.AccessToken,
		); err != nil {
			log.WithError(err).Warn("failed to store updated access token")
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to store updated access token")
		}
	} else {
		log.Info("access token for link has not changed")
	}

	link.LinkStatus = models.LinkStatusSetup
	link.ErrorCode = nil
	if err = repo.UpdateLink(c.getContext(ctx), link); err != nil {
		return c.wrapPgError(ctx, err, "failed to update link status")
	}

	currentBankAccounts, err := repo.GetBankAccountsByLinkId(c.getContext(ctx), link.LinkId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve existing bank accounts")
	}
	currentBankAccountPlaidIds := map[string]struct{}{}
	for _, bankAccount := range currentBankAccounts {
		currentBankAccountPlaidIds[bankAccount.PlaidAccountId] = struct{}{}
	}
	newBankAccountPlaidIds := make([]string, 0, len(callbackRequest.AccountIds))
	for _, accountId := range callbackRequest.AccountIds {
		if _, ok := currentBankAccountPlaidIds[accountId]; ok {
			continue
		}

		newBankAccountPlaidIds = append(newBankAccountPlaidIds, accountId)
	}

	// If there are any new bank accounts due to the updated selection.
	if len(newBankAccountPlaidIds) > 0 {
		client, err := c.plaid.NewClientFromLink(c.getContext(ctx), link.AccountId, link.LinkId)
		if err != nil {
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create plaid client for link")
		}

		// Retrieve the details for those bank accounts from Plaid.
		// TODO We should just retrieve all the accounts, any that are missing in this list were probably removed during the
		// account update selection anyway. Don't delete those bank accounts, but mark them as no longer in sync.
		plaidAccounts, err := client.GetAccounts(c.getContext(ctx), newBankAccountPlaidIds...)
		if err != nil {
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve new bank accounts")
		}

		now := time.Now()
		accounts := make([]models.BankAccount, len(plaidAccounts))
		for i, plaidAccount := range plaidAccounts {
			accounts[i] = models.BankAccount{
				AccountId:         repo.AccountId(),
				LinkId:            link.LinkId,
				PlaidAccountId:    plaidAccount.GetAccountId(),
				AvailableBalance:  plaidAccount.GetBalances().GetAvailable(),
				CurrentBalance:    plaidAccount.GetBalances().GetCurrent(),
				Name:              plaidAccount.GetName(),
				Mask:              plaidAccount.GetMask(),
				PlaidName:         plaidAccount.GetName(),
				PlaidOfficialName: plaidAccount.GetOfficialName(),
				Type:              models.BankAccountType(plaidAccount.GetType()),
				SubType:           models.BankAccountSubType(plaidAccount.GetSubType()),
				LastUpdated:       now,
			}
		}
		if err = repo.CreateBankAccounts(c.getContext(ctx), accounts...); err != nil {
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create new bank accounts")
		}
	}

	err = background.TriggerPullTransactions(c.getContext(ctx), c.jobRunner, background.PullTransactionsArguments{
		AccountId: link.AccountId,
		LinkId:    link.LinkId,
		Start:     time.Now().Add(-7 * 24 * time.Hour), // Last 7 days.
		End:       time.Now(),
	})
	if err != nil {
		log.WithError(err).Warn("failed to trigger pulling latest transactions after updating plaid link")
	}

	return ctx.JSON(http.StatusOK, link)
}

// Plaid Token Callback
// @Summary Plaid Token Callback
// @id plaid-token-callback
// @tags Plaid
// @description Receives the public token after a user has authenticated their bank account to exchange with plaid.
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param Request body swag.NewPlaidTokenCallbackRequest true "New token callback request."
// @Router /plaid/token/callback [post]
// @Success 200 {object} swag.PlaidTokenCallbackResponse
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) plaidTokenCallback(ctx echo.Context) error {
	if !c.configuration.Plaid.Enabled {
		return c.returnError(ctx, http.StatusNotAcceptable, "Plaid is not enabled on this server, only manual links are allowed.")
	}

	var callbackRequest struct {
		PublicToken     string   `json:"publicToken"`
		InstitutionId   string   `json:"institutionId"`
		InstitutionName string   `json:"institutionName"`
		AccountIds      []string `json:"accountIds"`
	}
	if err := ctx.Bind(&callbackRequest); err != nil {
		return c.invalidJson(ctx)
	}

	log := c.getLog(ctx)

	if err := c.removeLinkTokenFromCache(
		c.getContext(ctx),
		c.mustGetUserId(ctx),
		0,
	); err != nil {
		log.WithError(err).Warn("failed to remove link token from cache")
	}

	if len(callbackRequest.AccountIds) == 0 {
		return c.badRequest(ctx, "must select at least one account")
	}

	callbackRequest.PublicToken = strings.TrimSpace(callbackRequest.PublicToken)
	if callbackRequest.PublicToken == "" {
		return c.badRequest(ctx, "must provide a public token")
	}

	log.Debug("exchanging public token for plaid access token")
	result, err := c.plaid.ExchangePublicToken(c.getContext(ctx), callbackRequest.PublicToken)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to exchange token")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	var webhook string
	if c.configuration.Plaid.WebhooksEnabled {
		webhook = c.configuration.Plaid.GetWebhooksURL()
		if webhook == "" {
			log.Errorf("plaid webhooks are enabled, but they cannot be registered with without a domain")
		}
	}

	if err = c.plaidSecrets.UpdateAccessTokenForPlaidLinkId(
		c.getContext(ctx),
		repo.AccountId(),
		result.ItemId,
		result.AccessToken,
	); err != nil {
		log.WithError(err).Errorf("failed to store access token")
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to store access token")
	}

	plaidLink := models.PlaidLink{
		ItemId:          result.ItemId,
		Products:        consts.PlaidProductStrings(),
		WebhookUrl:      webhook,
		InstitutionId:   callbackRequest.InstitutionId,
		InstitutionName: callbackRequest.InstitutionName,
		UsePlaidSync:    true,
	}
	if err = repo.CreatePlaidLink(c.getContext(ctx), &plaidLink); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to store credentials")
	}

	link := models.Link{
		AccountId:          repo.AccountId(),
		PlaidLinkId:        &plaidLink.PlaidLinkID,
		PlaidInstitutionId: &plaidLink.InstitutionId,
		LinkType:           models.PlaidLinkType,
		LinkStatus:         models.LinkStatusPending,
		InstitutionName:    callbackRequest.InstitutionName,
		CreatedByUserId:    repo.UserId(),
	}
	if err = repo.CreateLink(c.getContext(ctx), &link); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create link")
	}

	// Create a plaid client for the new link.
	client, err := c.plaid.NewClient(c.getContext(ctx), &link, result.AccessToken, result.ItemId)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create Plaid client")
	}

	// Then use that client to retrieve that link's bank accounts.
	plaidAccounts, err := client.GetAccounts(c.getContext(ctx), callbackRequest.AccountIds...)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve accounts")
	}

	if len(plaidAccounts) == 0 {
		return c.returnError(ctx, http.StatusInternalServerError, "could not retrieve details for any accounts")
	}

	now := time.Now().UTC()
	accounts := make([]models.BankAccount, len(plaidAccounts))
	for i, plaidAccount := range plaidAccounts {
		accounts[i] = models.BankAccount{
			AccountId:         repo.AccountId(),
			LinkId:            link.LinkId,
			PlaidAccountId:    plaidAccount.GetAccountId(),
			AvailableBalance:  plaidAccount.GetBalances().GetAvailable(),
			CurrentBalance:    plaidAccount.GetBalances().GetCurrent(),
			Name:              plaidAccount.GetName(),
			Mask:              plaidAccount.GetMask(),
			PlaidName:         plaidAccount.GetName(),
			PlaidOfficialName: plaidAccount.GetOfficialName(),
			Type:              models.BankAccountType(plaidAccount.GetType()),
			SubType:           models.BankAccountSubType(plaidAccount.GetSubType()),
			LastUpdated:       now,
		}
	}
	if err = repo.CreateBankAccounts(c.getContext(ctx), accounts...); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create bank accounts")
	}

	if !c.configuration.Plaid.WebhooksEnabled {
		if plaidLink.UsePlaidSync {
			err = background.TriggerSyncPlaid(c.getContext(ctx), c.jobRunner, background.SyncPlaidArguments{
				AccountId: link.AccountId,
				LinkId:    link.LinkId,
			})
		} else {
			err = background.TriggerPullTransactions(c.getContext(ctx), c.jobRunner, background.PullTransactionsArguments{
				AccountId: link.AccountId,
				LinkId:    link.LinkId,
				Start:     time.Now().Add(-30 * 24 * time.Hour), // Last 30 days.
				End:       time.Now(),
			})
		}
		if err != nil {
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to pull initial transactions")
		}
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"linkId":  link.LinkId,
	})
}

// Wait For Plaid Account Data
// @Summary Wait For Plaid Account Data
// @id wait-for-plaid-data
// @tags Plaid
// @description Long poll endpoint that will timeout if data has not yet been pulled. Or will return 200 if data is ready.
// @Security ApiKeyAuth
// @Param linkId path int true "Link ID for the plaid link that is being setup. NOTE: Not Plaid's ID, this is a numeric ID we assign to the object that is returned from the callback endpoint."
// @Router /plaid/link/setup/wait/{linkId:uint64} [get]
// @Success 200
// @Success 408
func (c *Controller) waitForPlaid(ctx echo.Context) error {
	if !c.configuration.Plaid.Enabled {
		return c.returnError(ctx, http.StatusNotAcceptable, "Plaid is not enabled on this server, only manual links are allowed.")
	}
	linkId, err := strconv.ParseUint(ctx.Param("linkId"), 10, 64)
	if linkId == 0 {
		return c.badRequest(ctx, "must specify a job Id")
	}

	log := c.log.WithFields(logrus.Fields{
		"accountId": c.mustGetAccountId(ctx),
		"linkId":    linkId,
	})

	repo := c.mustGetAuthenticatedRepository(ctx)
	link, err := repo.GetLink(c.getContext(ctx), linkId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve link")
	}

	// If the link is done just return.
	if link.LinkStatus == models.LinkStatusSetup {
		crumbs.Debug(c.getContext(ctx), "Link is setup, no need to poll.", nil)
		return ctx.NoContent(http.StatusOK)
	}

	channelName := fmt.Sprintf("initial:plaid:link:%d:%d", link.AccountId, link.LinkId)

	listener, err := c.ps.Subscribe(c.getContext(ctx), channelName)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to listen on channel")
	}
	defer func() {
		if err = listener.Close(); err != nil {
			log.WithFields(logrus.Fields{
				"accountId": c.mustGetAccountId(ctx),
				"linkId":    linkId,
			}).WithError(err).Error("failed to gracefully close listener")
		}
	}()

	crumbs.Debug(c.getContext(ctx), "Waiting for notification on channel", map[string]interface{}{
		"channel": channelName,
	})

	log.Debugf("waiting for link to be setup on channel: %s", channelName)

	span := sentry.StartSpan(c.getContext(ctx), "Wait For Notification")
	defer span.Finish()

	deadLine := time.NewTimer(30 * time.Second)
	defer deadLine.Stop()

	select {
	case <-deadLine.C:
		log.Trace("timed out waiting for link to be setup")
		return ctx.NoContent(http.StatusRequestTimeout)
	case <-listener.Channel():
		// Just exit successfully, any message on this channel is considered a success.
		log.Trace("link setup successfully")
		return ctx.NoContent(http.StatusOK)
	}
}

func (c *Controller) postSyncPlaidManually(ctx echo.Context) error {
	if !c.configuration.Plaid.Enabled {
		return c.returnError(ctx, http.StatusNotAcceptable, "Plaid is not enabled on this server, only manual links are allowed.")
	}

	var request struct {
		LinkId uint64 `json:"linkId"`
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

	if link.LinkType != models.PlaidLinkType {
		return c.badRequest(ctx, "cannot manually sync a non-Plaid link")
	}

	switch link.LinkStatus {
	case models.LinkStatusSetup, models.LinkStatusError:
		log.Debug("link is not revoked, triggering manual sync")
	default:
		log.WithField("status", link.LinkStatus).Warn("link is not in a valid status, it cannot be manually synced")
		return c.badRequest(ctx, "link is not in a valid status, it cannot be manually synced")
	}

	if ok, err := repo.UpdateLinkManualSyncTimestampMaybe(c.getContext(ctx), link.LinkId); err != nil {
		return c.wrapPgError(ctx, err, "could not manually sync link")
	} else if !ok {
		return c.returnError(ctx, http.StatusTooEarly, "link has been manually synced too recently")
	}

	if link.PlaidLink.UsePlaidSync {
		err = background.TriggerSyncPlaid(c.getContext(ctx), c.jobRunner, background.SyncPlaidArguments{
			AccountId: link.AccountId,
			LinkId:    link.LinkId,
		})
	} else {
		err = background.TriggerPullTransactions(c.getContext(ctx), c.jobRunner, background.PullTransactionsArguments{
			AccountId: link.AccountId,
			LinkId:    link.LinkId,
			Start:     time.Now().Add(-14 * 24 * time.Hour), // Last 14 days.
			End:       time.Now(),
		})
	}
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to trigger manual sync")
	}

	return ctx.NoContent(http.StatusAccepted)
}
