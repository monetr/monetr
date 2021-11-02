package controller

import (
	"net/http"

	"github.com/kataras/iris/v12"
	"github.com/monetr/monetr/pkg/swag"
)

func (c *Controller) institutionsController(p iris.Party) {
	p.Get("/{institutionId:string}", c.getInstitutionDetails)
}

// Retrieve institution details.
// @Summary Get Institution Details
// @id get-institution-details
// @tags Institutions
// @description Retrieve Plaid institution details using Plaid's institution ID.
// @Security ApiKeyAuth
// @Param institutionId path string true "Institution ID"
// @Router /institutions/{institutionId} [get]
// @Success 200 {object} swag.InstitutionResponse
// @Failure 402 {object} SubscriptionNotActiveError The user's subscription is not active.
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) getInstitutionDetails(ctx iris.Context) {
	institutionId := ctx.Params().GetString("institutionId")
	if institutionId == "" {
		c.returnError(ctx, http.StatusBadRequest, "must specify an institution ID")
		return
	}

	plaidInstitution, err := c.plaidInstitutions.GetInstitution(c.getContext(ctx), institutionId)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve institution details")
		return
	}

	response := swag.NewInstitutionResponse(plaidInstitution)
	ctx.JSON(response)
}
