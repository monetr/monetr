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
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/consts"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/platypus"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (c *Controller) storeLinkTokenInCache(
	ctx context.Context,
	userId ID[User],
	linkId ID[Link],
	linkToken string,
	expiration time.Time,
) error {
	span := sentry.StartSpan(ctx, "StoreLinkTokenInCache")
	defer span.Finish()

	key := fmt.Sprintf("plaid:in_progress:%s:%s", userId, linkId)
	return errors.Wrap(
		// Cache TTL's should not use the internal clock. Because redis has it's own
		// clock.
		c.Cache.SetEzTTL(span.Context(), key, linkToken, time.Until(expiration)),
		"failed to cache link token",
	)
}

func (c *Controller) checkCacheForLinkToken(
	ctx context.Context,
	userId ID[User],
	linkId ID[Link],
) (string, error) {
	span := sentry.StartSpan(ctx, "StoreLinkTokenInCache")
	defer span.Finish()

	key := fmt.Sprintf("plaid:in_progress:%s:%s", userId, linkId)
	var token string
	if err := c.Cache.GetEz(span.Context(), key, &token); err != nil {
		return "", errors.Wrap(err, "failed to retrieve cached link token")
	}
	return token, nil
}

func (c *Controller) removeLinkTokenFromCache(
	ctx context.Context,
	userId ID[User],
	linkId ID[Link],
) error {
	span := sentry.StartSpan(ctx, "RemoteLinkTokenFromCache")
	defer span.Finish()

	key := fmt.Sprintf("plaid:in_progress:%s:%s", userId, linkId)
	return errors.Wrap(
		c.Cache.Delete(span.Context(), key),
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

	if !c.Configuration.Plaid.Enabled {
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
	if maxLinks := c.Configuration.Links.MaxNumberOfLinks; maxLinks > 0 && numberOfLinks >= maxLinks {
		return c.badRequest(ctx, "max number of Plaid links already reached")
	}

	// If billing is enabled and the current account is trialing, then limit them to a single Plaid link until their
	// trial has expired.
	if c.Configuration.Stripe.IsBillingEnabled() {
		trialing, err := c.Billing.GetSubscriptionIsTrialing(
			c.getContext(ctx),
			c.mustGetAccountId(ctx),
		)
		if err != nil {
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to determine trial status")
		}

		if trialing && numberOfLinks > 0 {
			log.WithFields(logrus.Fields{
				"numberOfLinks": numberOfLinks,
				"trialing":      trialing,
			}).Warn("cannot add more Plaid links during trial")
			return c.badRequest(ctx, "Cannot add additional Plaid links during trial")
		}
	}

	// If we are trying to not send a ton of requests then check the cache to see if we still have a valid link token that
	// we can use.
	if checkCache, err := strconv.ParseBool(ctx.QueryParam("use_cache")); err == nil && checkCache {
		if linkToken, err := c.checkCacheForLinkToken(
			c.getContext(ctx),
			userId,
			"",
		); err == nil && len(linkToken) > 0 {
			log.Info("successfully found existing link token in cache")
			return ctx.JSON(http.StatusOK, map[string]any{
				"linkToken": linkToken,
			})
		}
		log.Info("no link token was found in the cache")
	}

	legalName := ""
	if len(me.Login.LastName) > 0 {
		legalName = fmt.Sprintf("%s %s", me.Login.FirstName, me.Login.LastName)
	}

	log.Trace("creating Plaid link token")
	token, err := c.Plaid.CreateLinkToken(c.getContext(ctx), platypus.LinkTokenOptions{
		ClientUserID:             userId.String(),
		LegalName:                legalName,
		EmailAddress:             me.Login.Email,
		EmailAddressVerifiedTime: me.Login.EmailVerifiedAt,
	})
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create link token")
	}

	if err = c.storeLinkTokenInCache(
		c.getContext(ctx),
		me.UserId,
		"", // Since no link exists this should be cached without a link Id.
		token.Token(),
		token.Expiration(),
	); err != nil {
		log.WithError(err).Warn("failed to cache link token")
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"linkToken": token.Token(),
	})
}

func (c *Controller) putUpdatePlaidLink(ctx echo.Context) error {
	if !c.Configuration.Plaid.Enabled {
		return c.returnError(ctx, http.StatusNotAcceptable, "Plaid is not enabled on this server, only manual links are allowed.")
	}

	linkId, err := ParseID[Link](ctx.Param("linkId"))
	if err != nil || linkId.IsZero() {
		return c.badRequest(ctx, "must specify a valid link Id")
	}

	updateAccountSelection := urlParamBoolDefault(ctx, "update_account_selection", false)

	log := c.getLog(ctx).WithField("linkId", linkId)

	// Retrieve the user's details. We need to pass some of these along to plaid as part of the linking process.
	repo := c.mustGetAuthenticatedRepository(ctx)

	link, err := repo.GetLink(c.getContext(ctx), linkId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve link")
	}

	if link.LinkType != PlaidLinkType {
		return c.badRequest(ctx, "cannot update a non-Plaid link")
	}

	if link.PlaidLink == nil {
		return c.returnError(ctx, http.StatusInternalServerError, "no Plaid details associated with link")
	}

	me, err := repo.GetMe(c.getContext(ctx))
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve user details")
	}

	client, err := c.Plaid.NewClientFromLink(c.getContext(ctx), me.AccountId, linkId)
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

	return ctx.JSON(http.StatusOK, map[string]any{
		"linkToken": token.Token(),
	})
}

func (c *Controller) updatePlaidTokenCallback(ctx echo.Context) error {
	if !c.Configuration.Plaid.Enabled {
		return c.returnError(ctx, http.StatusNotAcceptable, "Plaid is not enabled on this server, only manual links are allowed.")
	}

	var callbackRequest struct {
		LinkId      ID[Link] `json:"linkId"`
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

	result, err := c.Plaid.ExchangePublicToken(c.getContext(ctx), callbackRequest.PublicToken)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to exchange token")
	}

	secrets := c.mustGetSecretsRepository(ctx)
	secret, err := secrets.Read(c.getContext(ctx), link.PlaidLink.SecretId)
	if err != nil {
		log.WithError(err).Warn("failed to retrieve access token for existing plaid link")
	}

	if secret.Value != result.AccessToken {
		log.Info("access token for link has been updated")
		secret.Value = result.AccessToken
		if err = secrets.Store(c.getContext(ctx), secret); err != nil {
			log.WithError(err).Warn("failed to store updated access token")
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to store updated access token")
		}
	} else {
		log.Info("access token for link has not changed")
	}

	link.PlaidLink.Status = PlaidLinkStatusSetup
	link.PlaidLink.ErrorCode = nil
	if err = repo.UpdatePlaidLink(c.getContext(ctx), link.PlaidLink); err != nil {
		return c.wrapPgError(ctx, err, "failed to update link status")
	}

	currentBankAccounts, err := repo.GetPlaidBankAccountsByLinkId(c.getContext(ctx), link.LinkId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve existing bank accounts")
	}
	currentBankAccountPlaidIds := map[string]struct{}{}
	for _, bankAccount := range currentBankAccounts {
		currentBankAccountPlaidIds[bankAccount.PlaidId] = struct{}{}
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
		client, err := c.Plaid.NewClientFromLink(c.getContext(ctx), link.AccountId, link.LinkId)
		if err != nil {
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create plaid client for link")
		}

		// Retrieve the details for those bank accounts from Plaid.
		// TODO We should just retrieve all the accounts, any that are missing in
		// this list were probably removed during the account update selection
		// anyway. Don't delete those bank accounts, but mark them as no longer in
		// sync.
		plaidAccounts, err := client.GetAccounts(c.getContext(ctx), newBankAccountPlaidIds...)
		if err != nil {
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve new bank accounts")
		}

		now := time.Now()
		accounts := make([]*BankAccount, len(plaidAccounts))
		for i := range plaidAccounts {
			plaidAccount := plaidAccounts[i]
			plaidBankAccount := PlaidBankAccount{
				PlaidLinkId:      *link.PlaidLinkId,
				PlaidId:          plaidAccount.GetAccountId(),
				Name:             plaidAccount.GetName(),
				OfficialName:     plaidAccount.GetOfficialName(),
				Mask:             plaidAccount.GetMask(),
				AvailableBalance: plaidAccount.GetBalances().GetAvailable(),
				CurrentBalance:   plaidAccount.GetBalances().GetCurrent(),
				LimitBalance:     plaidAccount.GetBalances().GetLimit(),
				Currency:         plaidAccount.GetCurrencyCode(),
			}
			if err := repo.CreatePlaidBankAccount(c.getContext(ctx), &plaidBankAccount); err != nil {
				return c.wrapPgError(ctx, err, "failed to create plaid bank account")
			}

			accounts[i] = &BankAccount{
				AccountId:          repo.AccountId(),
				LinkId:             link.LinkId,
				PlaidBankAccountId: &plaidBankAccount.PlaidBankAccountId,
				PlaidBankAccount:   &plaidBankAccount,
				AvailableBalance:   plaidAccount.GetBalances().GetAvailable(),
				CurrentBalance:     plaidAccount.GetBalances().GetCurrent(),
				LimitBalance:       plaidAccount.GetBalances().GetLimit(),
				Name:               plaidAccount.GetName(),
				Mask:               plaidAccount.GetMask(),
				AccountType:        BankAccountType(plaidAccount.GetType()),
				AccountSubType:     BankAccountSubType(plaidAccount.GetSubType()),
				Currency:           plaidAccount.GetCurrencyCode(),
				LastUpdated:        now,
			}
		}
		if err = repo.CreateBankAccounts(c.getContext(ctx), accounts...); err != nil {
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create new bank accounts")
		}
	}

	err = background.TriggerSyncPlaid(c.getContext(ctx), c.JobRunner, background.SyncPlaidArguments{
		AccountId: link.AccountId,
		LinkId:    link.LinkId,
	})
	if err != nil {
		log.WithError(err).Warn("failed to trigger pulling latest transactions after updating plaid link")
	}

	return ctx.JSON(http.StatusOK, link)
}

func (c *Controller) postPlaidTokenCallback(ctx echo.Context) error {
	if !c.Configuration.Plaid.Enabled {
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
		"",
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
	result, err := c.Plaid.ExchangePublicToken(c.getContext(ctx), callbackRequest.PublicToken)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to exchange token")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	var webhook string
	if c.Configuration.Plaid.WebhooksEnabled {
		webhook = c.Configuration.Plaid.GetWebhooksURL()
		if webhook == "" {
			log.Errorf("plaid webhooks are enabled, but they cannot be registered with without a domain")
		}
	}

	secrets := c.mustGetSecretsRepository(ctx)
	secret := repository.SecretData{
		Kind:  SecretKindPlaid,
		Value: result.AccessToken,
	}
	if err = secrets.Store(c.getContext(ctx), &secret); err != nil {
		log.WithError(err).Errorf("failed to store access token")
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to store access token")
	}

	plaidLink := PlaidLink{
		PlaidId:         result.ItemId,
		SecretId:        secret.SecretId,
		Products:        consts.PlaidProductStrings(),
		WebhookUrl:      webhook,
		Status:          PlaidLinkStatusPending,
		InstitutionId:   callbackRequest.InstitutionId,
		InstitutionName: callbackRequest.InstitutionName,
	}
	if err = repo.CreatePlaidLink(c.getContext(ctx), &plaidLink); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create Plaid link")
	}

	link := Link{
		AccountId:       repo.AccountId(),
		PlaidLinkId:     &plaidLink.PlaidLinkId,
		InstitutionName: callbackRequest.InstitutionName,
		LinkType:        PlaidLinkType,
	}
	if err = repo.CreateLink(c.getContext(ctx), &link); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create link")
	}

	// Create a plaid client for the new link.
	client, err := c.Plaid.NewClient(c.getContext(ctx), &link, result.AccessToken, result.ItemId)
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

	now := c.Clock.Now().UTC()
	accounts := make([]*BankAccount, len(plaidAccounts))
	for i := range plaidAccounts {
		plaidAccount := plaidAccounts[i]
		plaidBankAccount := PlaidBankAccount{
			PlaidLinkId:      plaidLink.PlaidLinkId,
			PlaidId:          plaidAccount.GetAccountId(),
			Name:             plaidAccount.GetName(),
			OfficialName:     plaidAccount.GetOfficialName(),
			Mask:             plaidAccount.GetMask(),
			AvailableBalance: plaidAccount.GetBalances().GetAvailable(),
			CurrentBalance:   plaidAccount.GetBalances().GetCurrent(),
			LimitBalance:     plaidAccount.GetBalances().GetLimit(),
			Currency:         plaidAccount.GetCurrencyCode(),
		}
		if err := repo.CreatePlaidBankAccount(
			c.getContext(ctx),
			&plaidBankAccount,
		); err != nil {
			return c.wrapPgError(ctx, err, "failed to create plaid bank account")
		}

		accounts[i] = &BankAccount{
			LinkId:             link.LinkId,
			PlaidBankAccountId: &plaidBankAccount.PlaidBankAccountId,
			AvailableBalance:   plaidAccount.GetBalances().GetAvailable(),
			CurrentBalance:     plaidAccount.GetBalances().GetCurrent(),
			LimitBalance:       plaidAccount.GetBalances().GetLimit(),
			Name:               plaidAccount.GetName(),
			Mask:               plaidAccount.GetMask(),
			AccountType:        BankAccountType(plaidAccount.GetType()),
			AccountSubType:     BankAccountSubType(plaidAccount.GetSubType()),
			Currency:           plaidAccount.GetCurrencyCode(),
			LastUpdated:        now,
			Status:             ActiveBankAccountStatus,
		}
	}

	if err = repo.CreateBankAccounts(c.getContext(ctx), accounts...); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create bank accounts")
	}

	if !c.Configuration.Plaid.WebhooksEnabled {
		err = background.TriggerSyncPlaid(c.getContext(ctx), c.JobRunner, background.SyncPlaidArguments{
			AccountId: link.AccountId,
			LinkId:    link.LinkId,
		})
		if err != nil {
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to pull initial transactions")
		}
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"linkId": link.LinkId,
	})
}

func (c *Controller) getWaitForPlaid(ctx echo.Context) error {
	if !c.Configuration.Plaid.Enabled {
		return c.returnError(ctx, http.StatusNotAcceptable, "Plaid is not enabled on this server, only manual links are allowed.")
	}
	linkId, err := ParseID[Link](ctx.Param("linkId"))
	if err != nil || linkId.IsZero() {
		return c.badRequest(ctx, "must specify a valid link Id")
	}

	log := c.Log.WithFields(logrus.Fields{
		"accountId": c.mustGetAccountId(ctx),
		"linkId":    linkId,
	})

	repo := c.mustGetAuthenticatedRepository(ctx)
	link, err := repo.GetLink(c.getContext(ctx), linkId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve link")
	}

	if link.LinkType != PlaidLinkType {
		return c.badRequest(ctx, "Link is not a Plaid link")
	}

	// If the link is done just return.
	if link.PlaidLink.Status == PlaidLinkStatusSetup {
		crumbs.Debug(c.getContext(ctx), "Link is setup, no need to poll.", nil)
		return ctx.NoContent(http.StatusOK)
	}

	channelName := fmt.Sprintf("initial:plaid:link:%s:%s", link.AccountId, link.LinkId)

	listener, err := c.PubSub.Subscribe(
		c.getContext(ctx),
		link.AccountId,
		channelName,
	)
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

	crumbs.Debug(c.getContext(ctx), "Waiting for notification on channel", map[string]any{
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

func (c *Controller) postPlaidLinkSync(ctx echo.Context) error {
	if !c.Configuration.Plaid.Enabled {
		return c.returnError(ctx, http.StatusNotAcceptable, "Plaid is not enabled on this server, only manual links are allowed.")
	}

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

	if link.LinkType != PlaidLinkType {
		return c.badRequest(ctx, "cannot manually sync a non-Plaid link")
	}

	switch link.PlaidLink.Status {
	case PlaidLinkStatusSetup, PlaidLinkStatusError:
		log.Debug("link is not revoked, triggering manual sync")
	default:
		log.WithField("status", link.PlaidLink.Status).Warn("link is not in a valid status, it cannot be manually synced")
		return c.badRequest(ctx, "link is not in a valid status, it cannot be manually synced")
	}

	plaidLink := link.PlaidLink
	if lastManualSync := plaidLink.LastManualSync; lastManualSync != nil && lastManualSync.After(c.Clock.Now().Add(-30*time.Minute)) {
		return c.returnError(ctx, http.StatusTooEarly, "link has been manually synced too recently")
	}

	plaidLink.LastManualSync = myownsanity.Pointer(c.Clock.Now().UTC())
	if err := repo.UpdatePlaidLink(c.getContext(ctx), plaidLink); err != nil {
		return c.wrapPgError(ctx, err, "could not manually sync link")
	}

	err = background.TriggerSyncPlaid(c.getContext(ctx), c.JobRunner, background.SyncPlaidArguments{
		AccountId: link.AccountId,
		LinkId:    link.LinkId,
	})
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to trigger manual sync")
	}

	return ctx.NoContent(http.StatusAccepted)
}
