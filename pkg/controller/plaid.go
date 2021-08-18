package controller

import (
	"context"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/kataras/iris/v12"
	"github.com/monetr/rest-api/pkg/crumbs"
	"github.com/monetr/rest-api/pkg/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"

	"github.com/kataras/iris/v12/core/router"
	"github.com/plaid/plaid-go/plaid"
)

func (c *Controller) handlePlaidLinkEndpoints(p router.Party) {
	p.Put("/update/{linkId:uint64}", c.updatePlaidLink)
	p.Post("/update/callback", c.updatePlaidTokenCallback)
	p.Get("/token/new", c.newPlaidToken)
	p.Post("/token/callback", c.plaidTokenCallback)
	p.Get("/setup/wait/{linkId:uint64}", c.waitForPlaid)
}

func (c *Controller) storeLinkTokenInCache(ctx context.Context, log *logrus.Entry, userId uint64, linkToken string, expiration time.Time) error {
	span := sentry.StartSpan(ctx, "StoreLinkTokenInCache")
	defer span.Finish()

	cache, err := c.cache.GetContext(ctx)
	if err != nil {
		log.WithError(err).Warn("failed to get cache connection")
		return errors.Wrap(err, "failed to get cache connection")
	}
	defer cache.Close()

	key := fmt.Sprintf("plaid:in_progress:%d", userId)
	return errors.Wrap(cache.Send("SET", key, linkToken, "EXAT", expiration.Unix()), "failed to cache link token")
}

func (c *Controller) removeLinkTokenFromCache(ctx context.Context, log *logrus.Entry, userId uint64) error {
	span := sentry.StartSpan(ctx, "RemoteLinkTokenFromCache")
	defer span.Finish()

	cache, err := c.cache.GetContext(ctx)
	if err != nil {
		log.WithError(err).Warn("failed to get cache connection")
		return errors.Wrap(err, "failed to get cache connection")
	}
	defer cache.Close()

	key := fmt.Sprintf("plaid:in_progress:%d", userId)
	return errors.Wrap(cache.Send("DEL", key), "failed to remove link token from cache")
}

// New Plaid Token
// @Summary New Plaid Token
// @id new-plaid-token
// @tags Plaid
// @description Generates a link token from Plaid to be used to authenticate a user's bank account with our application.
// @Security ApiKeyAuth
// @Produce json
// @Router /plaid/token/new [get]
// @Param use_cache query bool false "If true, the API will check and see if a plaid link token already exists for the current user. If one is present then it is returned instead of creating a new link token."
// @Success 200 {object} swag.PlaidNewLinkTokenResponse
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) newPlaidToken(ctx iris.Context) {
	// Retrieve the user's details. We need to pass some of these along to
	// plaid as part of the linking process.
	me, err := c.mustGetAuthenticatedRepository(ctx).GetMe(c.getContext(ctx))
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to get user details for link")
		return
	}

	userId := c.mustGetUserId(ctx)

	log := c.log.WithFields(logrus.Fields{
		"accountId": me.AccountId,
		"userId":    me.UserId,
		"loginId":   me.LoginId,
	})

	checkCacheForLinkToken := func(ctx context.Context) (linkToken string, _ error) {
		span := sentry.StartSpan(ctx, "CheckCacheForLinkToken")
		defer span.Finish()

		cache, err := c.cache.GetContext(ctx)
		if err != nil {
			log.WithError(err).Warn("failed to get cache connection")
			return "", errors.Wrap(err, "failed to get cache connection")
		}
		defer cache.Close()

		// Check and see if there is already a plaid link in progress for the current user.
		result, err := cache.Do("GET", fmt.Sprintf("plaid:in_progress:%d", me.UserId))
		if err != nil {
			log.WithError(err).Warn("failed to retrieve link token from cache")
			return "", errors.Wrap(err, "failed to retrieve link token from cache")
		}

		switch actual := result.(type) {
		case string:
			return actual, nil
		case *string:
			if actual != nil {
				return *actual, nil
			}
		case []byte:
			return string(actual), nil
		}

		return "", nil
	}

	if checkCache, err := ctx.URLParamBool("use_cache"); err == nil && checkCache {
		if linkToken, err := checkCacheForLinkToken(c.getContext(ctx)); err == nil && len(linkToken) > 0 {
			log.Info("successfully found existing link token in cache")
			ctx.JSON(map[string]interface{}{
				"linkToken": linkToken,
			})
			return
		}
	}

	plaidProducts := []string{
		"transactions",
	}

	legalName := ""
	if len(me.LastName) > 0 {
		legalName = fmt.Sprintf("%s %s", me.FirstName, me.LastName)
	}

	var phoneNumber string
	if me.Login.PhoneNumber != nil {
		phoneNumber = me.Login.PhoneNumber.E164()
	}

	var webhook string
	if c.configuration.Plaid.WebhooksEnabled {
		domain := c.configuration.Plaid.WebhooksDomain
		if domain != "" {
			webhook = fmt.Sprintf("%s/plaid/webhook", c.configuration.Plaid.WebhooksDomain)
		} else {
			c.log.Errorf("plaid webhooks are enabled, but they cannot be registered with without a domain")
		}
	}

	redirectUri := fmt.Sprintf("https://%s/plaid/oauth-return", c.configuration.UIDomainName)

	token, err := c.plaid.CreateLinkToken(c.getContext(ctx), plaid.LinkTokenConfigs{
		User: &plaid.LinkTokenUser{
			ClientUserID: strconv.FormatUint(userId, 10),
			LegalName:    legalName,
			PhoneNumber:  phoneNumber,
			EmailAddress: me.Login.Email,
			// TODO (elliotcourant) I'm going to leave these be for now but we need
			//  to loop back and add this once email/phone verification is working.
			PhoneNumberVerifiedTime:  time.Time{},
			EmailAddressVerifiedTime: time.Time{},
		},
		ClientName: "monetr",
		Products:   plaidProducts,
		CountryCodes: []string{
			"US",
		},
		Webhook:               webhook,
		AccountFilters:        nil,
		CrossAppItemAdd:       nil,
		PaymentInitiation:     nil,
		Language:              "en",
		LinkCustomizationName: "",
		RedirectUri:           redirectUri,
	})
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create link token")
		return
	}

	if err = c.storeLinkTokenInCache(c.getContext(ctx), log, me.UserId, token.LinkToken, token.Expiration); err != nil {
		log.WithError(err).Warn("failed to cache link token")
	}

	ctx.JSON(map[string]interface{}{
		"linkToken": token.LinkToken,
	})
}

// Update Plaid Link
// @Summary Update Plaid Link
// @id update-plaid-link
// @tags Plaid
// @description Update an existing Plaid link, this can be used to re-authenticate a link if it requires it or to potentially solve an error state.
// @Security ApiKeyAuth
// @Produce json
// @Router /plaid/update/{linkId:uint64} [put]
// @Param linkId path uint64 true "The Link Id that you wish to put into update mode, must be a Plaid link."
// @Success 200 {object} swag.PlaidNewLinkTokenResponse
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) updatePlaidLink(ctx iris.Context) {
	linkId := ctx.Params().GetUint64Default("linkId", 0)
	if linkId == 0 {
		c.badRequest(ctx, "must specify a link Id")
		return
	}

	log := c.getLog(ctx).WithField("linkId", linkId)

	// Retrieve the user's details. We need to pass some of these along to
	// plaid as part of the linking process.
	repo := c.mustGetAuthenticatedRepository(ctx)

	link, err := repo.GetLink(c.getContext(ctx), linkId)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to retrieve link")
		return
	}

	if link.LinkType != models.PlaidLinkType {
		c.badRequest(ctx, "cannot update a non-Plaid link")
		return
	}

	if link.PlaidLink == nil {
		c.returnError(ctx, http.StatusInternalServerError, "no Plaid details associated with link")
		return
	}

	me, err := repo.GetMe(c.getContext(ctx))
	if err != nil {
		c.wrapPgError(ctx, err, "failed to retrieve user details")
		return
	}

	legalName := ""
	if len(me.LastName) > 0 {
		legalName = fmt.Sprintf("%s %s", me.FirstName, me.LastName)
	} else {
		// TODO Handle a missing last name, we need a legal name Plaid.
		//  Should this be considered an error state?
	}

	var phoneNumber string
	if me.Login.PhoneNumber != nil {
		phoneNumber = me.Login.PhoneNumber.E164()
	}

	var webhook string
	if c.configuration.Plaid.WebhooksEnabled {
		domain := c.configuration.Plaid.WebhooksDomain
		if domain != "" {
			webhook = fmt.Sprintf("%s/plaid/webhook", c.configuration.Plaid.WebhooksDomain)
		} else {
			log.Errorf("plaid webhooks are enabled, but they cannot be registered with without a domain")
		}
	}

	accessToken, err := c.plaidSecrets.GetAccessTokenForPlaidLinkId(c.getContext(ctx), repo.AccountId(), link.PlaidLink.ItemId)
	if err != nil {
		log.WithError(err).Errorf("failed to retrieve current access token")
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve current access token")
		return
	}

	redirectUri := fmt.Sprintf("https://%s/plaid/oauth-return", c.configuration.UIDomainName)

	token, err := c.plaid.CreateLinkToken(c.getContext(ctx), plaid.LinkTokenConfigs{
		User: &plaid.LinkTokenUser{
			ClientUserID: strconv.FormatUint(me.UserId, 10),
			LegalName:    legalName,
			PhoneNumber:  phoneNumber,
			EmailAddress: me.Login.Email,
			// TODO Add in email/phone verification.
			PhoneNumberVerifiedTime:  time.Time{},
			EmailAddressVerifiedTime: time.Time{},
		},
		ClientName: "monetr",
		CountryCodes: []string{
			"US",
		},
		Webhook:               webhook,
		AccountFilters:        nil,
		CrossAppItemAdd:       nil,
		PaymentInitiation:     nil,
		Language:              "en",
		LinkCustomizationName: "",
		RedirectUri:           redirectUri,
		AccessToken:           accessToken,
	})
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create link token")
		return
	}

	if err = c.storeLinkTokenInCache(c.getContext(ctx), log, me.UserId, token.LinkToken, token.Expiration); err != nil {
		log.WithError(err).Warn("failed to cache link token")
	}

	ctx.JSON(map[string]interface{}{
		"linkToken": token.LinkToken,
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
func (c *Controller) updatePlaidTokenCallback(ctx iris.Context) {
	var callbackRequest struct {
		LinkId      uint64 `json:"linkId"`
		PublicToken string `json:"publicToken"`
	}
	if err := ctx.ReadJSON(&callbackRequest); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed json")
		return
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	link, err := repo.GetLink(c.getContext(ctx), callbackRequest.LinkId)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to retrieve link")
		return
	}

	result, err := c.plaid.ExchangePublicToken(c.getContext(ctx), callbackRequest.PublicToken)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to exchange token")
		return
	}

	log := c.getLog(ctx)

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
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to store updated access token")
			return
		}
	} else {
		log.Info("access token for link has not changed")
	}

	link.LinkStatus = models.LinkStatusSetup
	link.ErrorCode = nil
	if err = repo.UpdateLink(c.getContext(ctx), link); err != nil {
		c.wrapPgError(ctx, err, "failed to update link status")
		return
	}

	_, err = c.job.TriggerPullLatestTransactions(link.AccountId, link.LinkId, 0)
	if err != nil {
		log.WithError(err).Warn("failed to trigger pulling latest transactions after updating plaid link")
	}

	ctx.JSON(link)
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
func (c *Controller) plaidTokenCallback(ctx iris.Context) {
	var callbackRequest struct {
		PublicToken     string   `json:"publicToken"`
		InstitutionId   string   `json:"institutionId"`
		InstitutionName string   `json:"institutionName"`
		AccountIds      []string `json:"accountIds"`
	}
	if err := ctx.ReadJSON(&callbackRequest); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed json")
		return
	}

	log := c.getLog(ctx)

	if err := c.removeLinkTokenFromCache(c.getContext(ctx), log, c.mustGetUserId(ctx)); err != nil {
		log.WithError(err).Warn("failed to remove link token from cache")
	}

	if len(callbackRequest.AccountIds) == 0 {
		c.returnError(ctx, http.StatusBadRequest, "must select at least one account")
		return
	}

	log.Debug("exchanging public token for plaid access token")
	result, err := c.plaid.ExchangePublicToken(c.getContext(ctx), callbackRequest.PublicToken)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to exchange token")
		return
	}

	plaidAccounts, err := c.plaid.GetAccounts(c.getContext(ctx), result.AccessToken, plaid.GetAccountsOptions{
		AccountIDs: callbackRequest.AccountIds,
	})
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve accounts")
		return
	}

	if len(plaidAccounts) == 0 {
		c.returnError(ctx, http.StatusInternalServerError, "could not retrieve details for any accounts")
		return
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	var webhook string
	if c.configuration.Plaid.WebhooksEnabled {
		domain := c.configuration.Plaid.WebhooksDomain
		if domain != "" {
			webhook = fmt.Sprintf("%s/plaid/webhook", c.configuration.Plaid.WebhooksDomain)
		} else {
			c.log.Errorf("plaid webhooks are enabled, but they cannot be registered with without a domain")
		}
	}

	if err = c.plaidSecrets.UpdateAccessTokenForPlaidLinkId(
		c.getContext(ctx),
		repo.AccountId(),
		result.ItemID,
		result.AccessToken,
	); err != nil {
		log.WithError(err).Errorf("failed to store access token")
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to store access token")
		return
	}

	plaidLink := models.PlaidLink{
		ItemId: result.ItemID,
		Products: []string{
			// TODO (elliotcourant) Make this based on what product's we sent in the create link token request.
			"transactions",
		},
		WebhookUrl:      webhook,
		InstitutionId:   callbackRequest.InstitutionId,
		InstitutionName: callbackRequest.InstitutionName,
	}
	if err = repo.CreatePlaidLink(c.getContext(ctx), &plaidLink); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to store credentials")
		return
	}

	link := models.Link{
		AccountId:       repo.AccountId(),
		PlaidLinkId:     &plaidLink.PlaidLinkID,
		LinkType:        models.PlaidLinkType,
		LinkStatus:      models.LinkStatusPending,
		InstitutionName: callbackRequest.InstitutionName,
		CreatedByUserId: repo.UserId(),
	}
	if err = repo.CreateLink(c.getContext(ctx), &link); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create link")
		return
	}

	now := time.Now().UTC()
	accounts := make([]models.BankAccount, len(plaidAccounts))
	for i, plaidAccount := range plaidAccounts {
		accounts[i] = models.BankAccount{
			AccountId:         repo.AccountId(),
			LinkId:            link.LinkId,
			PlaidAccountId:    plaidAccount.AccountID,
			AvailableBalance:  int64(plaidAccount.Balances.Available * 100),
			CurrentBalance:    int64(plaidAccount.Balances.Current * 100),
			Name:              plaidAccount.Name,
			Mask:              plaidAccount.Mask,
			PlaidName:         plaidAccount.Name,
			PlaidOfficialName: plaidAccount.OfficialName,
			Type:              models.BankAccountType(plaidAccount.Type),
			SubType:           models.BankAccountSubType(plaidAccount.Subtype),
			LastUpdated:       now,
		}
	}
	if err = repo.CreateBankAccounts(c.getContext(ctx), accounts...); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create bank accounts")
		return
	}

	var jobIdStr *string
	if !c.configuration.Plaid.WebhooksEnabled {
		jobId, err := c.job.TriggerPullInitialTransactions(link.AccountId, link.CreatedByUserId, link.LinkId)
		if err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to pull initial transactions")
			return
		}

		jobIdStr = &jobId
	}

	ctx.JSON(map[string]interface{}{
		"success": true,
		"linkId":  link.LinkId,
		"jobId":   jobIdStr,
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
func (c *Controller) waitForPlaid(ctx iris.Context) {
	linkId := ctx.Params().GetUint64Default("linkId", 0)
	if linkId == 0 {
		c.badRequest(ctx, "must specify a job Id")
		return
	}

	log := c.log.WithFields(logrus.Fields{
		"accountId": c.mustGetAccountId(ctx),
		"linkId":    linkId,
	})

	repo := c.mustGetAuthenticatedRepository(ctx)
	link, err := repo.GetLink(c.getContext(ctx), linkId)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to retrieve link")
		return
	}

	// If the link is done just return.
	if link.LinkStatus == models.LinkStatusSetup {
		crumbs.Debug(c.getContext(ctx), "Link is setup, no need to poll.", nil)
		return
	}

	channelName := fmt.Sprintf("initial:plaid:link:%d:%d", link.AccountId, link.LinkId)

	listener, err := c.ps.Subscribe(c.getContext(ctx), channelName)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to listen on channel")
		return
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
		ctx.StatusCode(http.StatusRequestTimeout)
		log.Trace("timed out waiting for link to be setup")
		return
	case <-listener.Channel():
		// Just exit successfully, any message on this channel is considered a success.
		log.Trace("link setup successfully")
		return
	}
}
