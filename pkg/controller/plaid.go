package controller

import (
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/kataras/iris/v12"
	"github.com/monetrapp/rest-api/pkg/models"
	"net/http"
	"strconv"
	"time"

	"github.com/kataras/iris/v12/core/router"
	"github.com/plaid/plaid-go/plaid"
)

func (c *Controller) handlePlaidLinkEndpoints(p router.Party) {
	p.Get("/token/new", c.newPlaidToken)
	p.Post("/token/callback", c.plaidTokenCallback)
	p.Get("/setup/wait/{jobId:string}", c.waitForPlaid)
}

// New Plaid Token
// @Summary New Plaid Token
// @id new-plaid-token
// @tags Plaid
// @description Generates a link token from Plaid to be used to authenticate a user's bank account with our application.
// @Security ApiKeyAuth
// @Produce json
// @Router /plaid/token/new [get]
// @Success 200
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) newPlaidToken(ctx iris.Context) {
	// Retrieve the user's details. We need to pass some of these along to
	// plaid as part of the linking process.
	me, err := c.mustGetAuthenticatedRepository(ctx).GetMe()
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to get user details for link")
	}

	userId := c.mustGetUserId(ctx)
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

	plaidSpan := sentry.StartSpan(c.getContext(ctx), "Create Plaid Link Token")
	token, err := c.plaid.CreateLinkToken(plaid.LinkTokenConfigs{
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
		ClientName: "Hard",
		Products:   plaidProducts,
		CountryCodes: []string{
			"US",
		},
		// TODO (elliotcourant) Implement webhook once we are running in kube.
		Webhook:               webhook,
		AccountFilters:        nil,
		CrossAppItemAdd:       nil,
		PaymentInitiation:     nil,
		Language:              "en",
		LinkCustomizationName: "",
		RedirectUri:           "",
	})
	if err != nil {
		plaidSpan.Status = sentry.SpanStatusInternalError
		plaidSpan.Finish()
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create link token")
		return
	}
	plaidSpan.Status = sentry.SpanStatusOK
	plaidSpan.Finish()

	ctx.JSON(map[string]interface{}{
		"linkToken": token.LinkToken,
	})
}

// Plaid Token Callback
// @Summary Plaid Token Callback
// @id plaid-token-callback
// @tags Plaid
// @description Receives the public token after a user has authenticated their bank account to exchange with plaid.
// @Security ApiKeyAuth
// @Produce json
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

	if len(callbackRequest.AccountIds) == 0 {
		c.returnError(ctx, http.StatusBadRequest, "must select at least one account")
		return
	}

	result, err := c.plaid.ExchangePublicToken(callbackRequest.PublicToken)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to exchange token")
		return
	}

	plaidAccounts, err := c.plaid.GetAccountsWithOptions(result.AccessToken, plaid.GetAccountsOptions{
		AccountIDs: callbackRequest.AccountIds,
	})
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve accounts")
		return
	}

	if len(plaidAccounts.Accounts) == 0 {
		c.returnError(ctx, http.StatusInternalServerError, "could not retrieve details for any accounts")
		return
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	plaidLink := models.PlaidLink{
		ItemId:      result.ItemID,
		AccessToken: result.AccessToken,
		Products: []string{
			// TODO (elliotcourant) Make this based on what product's we sent in the create link token request.
			"transactions",
		},
		WebhookUrl:      "",
		InstitutionId:   callbackRequest.InstitutionId,
		InstitutionName: callbackRequest.InstitutionName,
	}
	if err := repo.CreatePlaidLink(&plaidLink); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to store credentials")
		return
	}

	link := models.Link{
		AccountId:       repo.AccountId(),
		PlaidLinkId:     &plaidLink.PlaidLinkID,
		LinkType:        models.PlaidLinkType,
		InstitutionName: callbackRequest.InstitutionName,
		CreatedByUserId: repo.UserId(),
	}
	if err = repo.CreateLink(&link); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create link")
		return
	}

	accounts := make([]models.BankAccount, len(plaidAccounts.Accounts))
	for i, plaidAccount := range plaidAccounts.Accounts {
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
			Type:              plaidAccount.Type,
			SubType:           plaidAccount.Subtype,
		}
	}
	if err = repo.CreateBankAccounts(accounts...); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create bank accounts")
		return
	}

	jobId, err := c.job.TriggerPullInitialTransactions(repo.AccountId(), repo.UserId(), link.LinkId)
	if err != nil {
		// TODO (elliotcourant) This error would technically throw out all of our data above. Including credentials.
		//  This might cause the account to appear as not linked when it technically is. Maybe this should not
		//  cause such a failure state?
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to queue transaction job")
		return
	}

	ctx.JSON(map[string]interface{}{
		"success": true,
		"jobId":   jobId,
	})
}

// Wait For Plaid Account Data
// @Summary Wait For Plaid Account Data
// @id wait-for-plaid-data
// @tags Plaid
// @description Long poll endpoint that will timeout if data has not yet been pulled. Or will return 200 if data is ready.
// @Security ApiKeyAuth
// @Param jobId path string true "Job ID returned from the token callback endpoint"
// @Router /plaid/setup/wait/{jobId:string} [get]
// @Success 200
// @Success 408
func (c *Controller) waitForPlaid(ctx iris.Context) {
	jobId := ctx.Params().GetStringDefault("jobId", "")
	if jobId == "" {
		c.badRequest(ctx, "must specify a job Id")
		return
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	job, err := repo.GetJob(jobId)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to retrieve job")
		return
	}

	// If the job is done just return.
	if job.FinishedAt != nil {
		return
	}

	channelName := fmt.Sprintf("job_%d_%s", c.mustGetAccountId(ctx), jobId)

	listener := c.db.Listen(ctx.Request().Context(), channelName)
	defer listener.Close()

	deadLine := time.NewTimer(30 * time.Second)
	defer deadLine.Stop()

ListenLoop:
	for {
		select {
		case <-deadLine.C:
			ctx.StatusCode(http.StatusRequestTimeout)
			break ListenLoop
		case notification := <-listener.Channel():
			if notification.Channel == channelName {
				break ListenLoop
			}
		}
	}

	if err = listener.Unlisten(ctx.Request().Context()); err != nil {
		c.log.WithError(err).Warnf("failed to stop listening on channel %s", channelName)
	}
}