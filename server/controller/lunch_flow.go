package controller

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/merge"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
	"github.com/monetr/validation/is"
	"github.com/pkg/errors"
)

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
				validation.Length(1, 100).Error("Lunch flow API Key must be between 1 and 100 characters"),
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

// postLunchFlowLink will create a new lunch flow link from the API request.
// This requires that the user provide a name for the link (which can be changed
// later) as well as the Lunch Flow API URL they want to use and the API Key for
// that URL.
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
		Kind:  models.SecretKindLunchFlow,
		Value: request.APIKey,
	}

	{ // Store the secret and generate an ID
		secrets := c.mustGetSecretsRepository(ctx)
		if err := secrets.Store(c.getContext(ctx), &secret); err != nil {
			return c.wrapPgError(ctx, err, "Failed to store Lunch Flow secret")
		}
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	lunchFlowLink := models.LunchFlowLink{
		SecretId:  secret.SecretId,
		ApiUrl:    request.LunchFlowURL,
		Status:    models.LunchFlowLinkStatusActive,
		CreatedBy: c.mustGetUserId(ctx),
	}

	// The lunch flow link itself needs to be created first.
	if err := repo.CreateLunchFlowLink(
		c.getContext(ctx),
		&lunchFlowLink,
	); err != nil {
		return c.wrapPgError(ctx, err, "Failed to create Lunch Flow link")
	}

	// Then create the regular link record to make it available to the user via
	// the API and UI.
	link := models.Link{
		LinkType:        models.LunchFlowLinkType,
		LunchFlowLinkId: &lunchFlowLink.LunchFlowLinkId,
		LunchFlowLink:   &lunchFlowLink,
		InstitutionName: request.Name,
		CreatedBy:       lunchFlowLink.CreatedBy,
	}

	if err := repo.CreateLink(c.getContext(ctx), &link); err != nil {
		return c.wrapPgError(ctx, err, "Failed to create Lunch Flow link")
	}

	return ctx.JSON(http.StatusOK, link)
}

func (c *Controller) postLunchFlowLinkSync(ctx echo.Context) error {
	// TODO
	return nil
}

func (c *Controller) getLunchFlowLinkBankAccounts(ctx echo.Context) error {
	// TODO
	return nil
}

func (c *Controller) patchLunchFlowLink(ctx echo.Context) error {
	// TODO
	return nil
}
