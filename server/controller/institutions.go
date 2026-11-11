package controller

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

// Retrieve institution details.
// @Summary Get Institution Details
// @id get-institution-details
// @tags Institutions
// @description Retrieve Plaid institution details using Plaid's institution ID.
// @Security ApiKeyAuth
// @Param institutionId path string true "Institution ID"
// @Router /institutions/{institutionId} [get]
// @Failure 402 {object} SubscriptionNotActiveError The user's subscription is not active.
// @Failure 404 {object} ApiError Plaid is not enabled on this server.
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) getInstitutionDetails(ctx *echo.Context) error {
	// Institution details only come from Plaid, without it there is no client to
	// ask so there is nothing to return.
	if !c.Configuration.Plaid.GetEnabled() {
		return c.notFound(ctx, "Plaid is not enabled on this server")
	}

	institutionId := ctx.Param("institutionId")
	if institutionId == "" {
		return c.badRequest(ctx, "must specify an institution ID")
	}

	plaidInstitution, err := c.PlaidInstitutions.GetInstitution(c.getContext(ctx), institutionId)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve institution details")
	}

	return ctx.JSON(http.StatusOK, plaidInstitution)
}
