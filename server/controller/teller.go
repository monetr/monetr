package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/sirupsen/logrus"
)

func (c *Controller) postTellerLink(ctx echo.Context) error {
	if !c.configuration.Teller.GetEnabled() {
		return c.returnError(ctx, http.StatusNotAcceptable, "Teller is not enabled on this server.")
	}

	var request struct {
		AccessToken string `json:"accessToken"`
		User        struct {
			Id string `json:"id"`
		} `json:"user"`
		Enrollment struct {
			Id           string `json:"id"`
			Institituion struct {
				Name string `json:"name"`
			} `json:"institution"`
		} `json:"enrollment"`
		Signatures []string `json:"signatures"`
	}
	if err := ctx.Bind(&request); err != nil {
		return c.invalidJson(ctx)
	}

	log := c.getLog(ctx)
	request.AccessToken = strings.TrimSpace(request.AccessToken)
	if request.AccessToken == "" {
		return c.badRequest(ctx, "must provide an access token")
	}

	request.Enrollment.Id = strings.TrimSpace(request.Enrollment.Id)
	if request.Enrollment.Id == "" {
		return c.badRequest(ctx, "must provide an enrollment Id")
	}

	request.User.Id = strings.TrimSpace(request.User.Id)
	if request.User.Id == "" {
		return c.badRequest(ctx, "must provide a user Id")
	}

	secretsRepo := c.mustGetSecretsRepository(ctx)
	secret := repository.Secret{
		Kind:   models.TellerSecretKind,
		Secret: request.AccessToken,
	}
	log.Debug("storing teller access token")
	if err := secretsRepo.Store(c.getContext(ctx), &secret); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to store access token")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	tellerLink := models.TellerLink{
		SecretId:             secret.SecretId,
		EnrollmentId:         request.Enrollment.Id,
		UserId:               request.User.Id,
		Status:               models.TellerLinkStatusPending,
		ErrorCode:            nil,
		InstitituionName:     request.Enrollment.Institituion.Name,
		LastManualSync:       nil,
		LastSuccessfulUpdate: nil,
		LastAttemptedUpdate:  nil,
	}
	if err := repo.CreateTellerLink(c.getContext(ctx), &tellerLink); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create Teller link")
	}

	link := models.Link{
		LinkType:        models.TellerLinkType,
		TellerLinkId:    &tellerLink.TellerLinkId,
		TellerLink:      &tellerLink,
		InstitutionName: request.Enrollment.Institituion.Name,
		Description:     nil,
	}
	if err := repo.CreateLink(c.getContext(ctx), &link); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create link")
	}

	background.TriggerSyncTeller(c.getContext(ctx), c.jobRunner, background.SyncTellerArguments{
		AccountId: link.AccountId,
		LinkId:    link.LinkId,
		Trigger:   "initial",
	})

	return ctx.JSON(http.StatusOK, link)
}

func (c *Controller) getWaitForTeller(ctx echo.Context) error {
	if !c.configuration.Teller.GetEnabled() {
		return c.returnError(ctx, http.StatusNotAcceptable, "Teller is not enabled on this server.")
	}

	linkId, _ := strconv.ParseUint(ctx.Param("linkId"), 10, 64)
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

	if link.LinkType != models.TellerLinkType {
		return c.badRequest(ctx, "Link is not a Teller link")
	}

	// If the link is done just return.
	if link.TellerLink.Status == models.TellerLinkStatusSetup {
		crumbs.Debug(c.getContext(ctx), "Link is setup, no need to poll.", nil)
		return ctx.NoContent(http.StatusOK)
	}

	channelName := fmt.Sprintf("initial:teller:link:%d:%d", link.AccountId, link.LinkId)

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

	deadLine := time.NewTimer(25 * time.Second)
	defer deadLine.Stop()

	select {
	case <-deadLine.C:
		log.Trace("timed out waiting for link to be setup")
		return ctx.NoContent(http.StatusRequestTimeout)
	case <-listener.Channel():
		// Just exit successfully, any message on this channel is considered a success.
		log.Trace("link setup successfully")
		time.Sleep(5 * time.Second)
		return ctx.NoContent(http.StatusOK)
	}
}
