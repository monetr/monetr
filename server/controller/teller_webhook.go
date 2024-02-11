package controller

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/teller"
	"github.com/sirupsen/logrus"
)

type tellerWebhook struct {
	Id      string `json:"id"`
	Payload struct {
		EnrollmentId string `json:"enrollment_id"`
		Reason       string `json:"reason"`
	} `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
}

func (c *Controller) postTellerWebhook(ctx echo.Context) error {
	if !c.configuration.Teller.GetEnabled() {
		return c.returnError(ctx, http.StatusNotAcceptable, "Teller is not enabled on this server.")
	}

	log := c.getLog(ctx)

	signature := ctx.Request().Header.Get("Teller-Signature")
	if strings.TrimSpace(signature) == "" {
		log.Debug("teller webhook is missing Teller-Signature header, ignoring request")
		return c.returnError(ctx, http.StatusUnauthorized, "unauthorized")
	}

	var signatureTimestamp int64
	var requestSignatures []string
	parts := strings.Split(signature, ",")
	for i := range parts {
		part := parts[i]

		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			log.WithField("signature", signature).Warn("signature has malformed part")
			return c.badRequest(ctx, "invalid request signature")
		}

		// If the part matches `t=signature_timestamp` then set the timestamp to
		// this value. Otherwise assume that it is matching the
		// `v1=signature_with_new_secret,v1=signature_with_old_secret` format.
		switch kv[0] {
		case "t":
			unix, err := strconv.ParseInt(kv[1], 10, 64)
			if err != nil {
				return c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "invalid signature timestamp")
			}
			signatureTimestamp = unix
		case "v1":
			requestSignatures = append(requestSignatures, kv[1])
		default:
			log.WithField("signaturePiece", part).Warn("unrecognized signature key pair")
		}
	}
	// If the timestamp is 0 or older than 3 minutes then reject it!
	if signatureTimestamp == 0 || time.Unix(signatureTimestamp, 0).Before(c.clock.Now().Add(-3*time.Minute)) {
		return c.badRequest(ctx, "teller signature timestamp is not valid")
	}

	log.Tracef("teller webhook has %d signature(s)", len(requestSignatures))

	webhookSecrets := c.configuration.Teller.WebhookSigningSecret
	if len(webhookSecrets) == 0 {
		log.Warn("no webhook signing secrets configured, teller webhook cannot be verified!")
		return ctx.NoContent(http.StatusOK)
	}

	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		log.WithError(err).Error("failed to read teller webhook request body")
		return c.badRequest(ctx, "invalid request body")
	}
	defer ctx.Request().Body.Close()

	// Then using all of the secrets we have configured, generate the signature
	// for that secret and the provided signature timestamp. One of the signatures
	// in the request must match one of these potential signatures.
	potentials := teller.GenerateWebhookSignatures(
		time.Unix(signatureTimestamp, 0),
		body,
		webhookSecrets,
	)
	log.Debugf("generated %d potential signature(s)", len(potentials))

	// Validate the provided signatures
	var matchedSignature *string
SignatureLoop:
	for x := range requestSignatures {
		requestSignature := requestSignatures[x]
		for y := range potentials {
			potentialSignature := potentials[y]

			if requestSignature == potentialSignature {
				matchedSignature = &potentialSignature
				break SignatureLoop
			}
		}
	}

	if matchedSignature == nil {
		log.WithFields(logrus.Fields{
			"requestSignatures":  requestSignatures,
			"potentialSignature": potentials,
		}).Warn("no matching signature found")
		return c.returnError(ctx, http.StatusUnauthorized, "unauthorized")
	}

	log.WithField("matchedSignature", *matchedSignature).Debug("teller webhook signature matches")

	var webhook tellerWebhook
	if err := json.Unmarshal(body, &webhook); err != nil {
		return c.invalidJson(ctx)
	}

	log = log.WithFields(logrus.Fields{
		"webhookId":   webhook.Id,
		"webhookType": webhook.Type,
	})

	if hub := sentry.GetHubFromContext(c.getContext(ctx)); hub != nil {
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetTag("webhook", "teller")
			scope.SetTag("teller.webhook.type", webhook.Type)
		})
	}

	repo := c.mustGetUnauthenticatedRepository(ctx)

	switch webhook.Type {
	case "webhook.test":
		log.Info("recieved test webhook successfully!")
		return ctx.NoContent(http.StatusOK)
	case "enrollment.disconnected":
		log = log.WithField("enrollmentId", webhook.Payload.EnrollmentId)
		log.Debug("looking up link by teller enrollment Id")
		link, err := repo.GetLinkByTellerEnrollmentId(
			c.getContext(ctx),
			webhook.Payload.EnrollmentId,
		)
		if err != nil {
			return c.wrapPgError(ctx, err, "failed to find link for enrollment")
		}

		authenticatedRepo := repository.NewRepositoryFromSession(
			c.clock,
			link.CreatedByUserId,
			link.AccountId,
			c.mustGetDatabase(ctx),
		)

		crumbs.IncludeUserInScope(c.getContext(ctx), link.AccountId)
		crumbs.AddTag(c.getContext(ctx), "linkId", strconv.FormatUint(link.LinkId, 10))

		tellerLink := link.TellerLink
		tellerLink.Status = models.TellerLinkStatusDisconnected
		tellerLink.ErrorCode = &webhook.Payload.Reason
		if err := authenticatedRepo.UpdateTellerLink(c.getContext(ctx), tellerLink); err != nil {
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to process teller webhook")
		}
	default:
		log.Warn("unrecognized teller webhook type recieved, unable to handle")
		return ctx.NoContent(http.StatusOK)
	}

	return ctx.NoContent(http.StatusOK)
}
