package controller

import (
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"net/http"
	"strings"
	"time"
)

// @tag.name Expenses
func (c *Controller) handleSpending(p iris.Party) {
	p.Get("/{bankAccountId:uint64}/spending", c.getExpenses)
	p.Post("/{bankAccountId:uint64}/spending", c.postExpenses)
	p.Post("/{bankAccountId:uint64}/spending/transfer", c.postSpendingTransfer)

	p.Put("/{bankAccountId:uint64}/spending/{expenseId:uint64}", func(ctx *context.Context) {

	})

	p.Delete("/{bankAccountId:uint64}/spending/{expenseId:uint64}", func(ctx *context.Context) {

	})
}

// List Spending
// @id list-spending
// @tags Spending
// @description List all of the spending for the specified bank account.
// @Security ApiKeyAuth
// @Param bankAccountId path int true "Bank Account ID"
// @Router /bank_accounts/{bankAccountId}/spending [get]
// @Success 200 {array} models.Spending
// @Failure 400 {object} InvalidBankAccountIdError Invalid Bank Account ID.
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) getExpenses(ctx *context.Context) {
	bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
	if bankAccountId == 0 {
		c.returnError(ctx, http.StatusBadRequest, "must specify valid bank account Id")
		return
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	expenses, err := repo.GetExpenses(bankAccountId)
	if err != nil {
		c.wrapPgError(ctx, err, "could not retrieve expenses")
		return
	}

	ctx.JSON(expenses)
}

// Create Spending
// @id create-spending
// @tags Spending
// @summary Create an spending for the specified bank account.
// @security ApiKeyAuth
// @accept json
// @product json
// @Param bankAccountId path int true "Bank Account ID"
// @Param Spending body models.Spending true "New spending"
// @Router /bank_accounts/{bankAccountId}/spending [post]
// @Success 200 {object} models.Spending
// @Failure 400 {object} InvalidBankAccountIdError "Invalid Bank Account ID."
// @Failure 400 {object} ApiError "Malformed JSON or invalid RRule."
// @Failure 500 {object} ApiError "Failed to persist data."
func (c *Controller) postExpenses(ctx *context.Context) {
	bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
	if bankAccountId == 0 {
		c.returnError(ctx, http.StatusBadRequest, "must specify valid bank account Id")
		return
	}

	spending := &models.Spending{}
	if err := ctx.ReadJSON(spending); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed JSON")
		return
	}

	spending.SpendingId = 0 // Make sure we create a new spending.
	spending.BankAccountId = bankAccountId
	spending.Name = strings.TrimSpace(spending.Name)
	spending.Description = strings.TrimSpace(spending.Description)

	if spending.Name == "" {
		c.returnError(ctx, http.StatusBadRequest, "spending must have a name")
		return
	}

	spending.LastRecurrence = nil
	spending.NextRecurrence = spending.RecurrenceRule.After(time.Now(), false)

	repo := c.mustGetAuthenticatedRepository(ctx)

	// We need to calculate what the next contribution will be for this new spending. So we need to retrieve it's funding
	// schedule. This also helps us validate that the user has provided a valid funding schedule id.
	fundingSchedule, err := repo.GetFundingSchedule(bankAccountId, spending.FundingScheduleId)
	if err != nil {
		c.wrapPgError(ctx, err, "could not find funding schedule specified")
		return
	}

	// We also need to know the current account's timezone, as contributions are made at midnight in that user's
	// timezone.
	account, err := repo.GetAccount()
	if err != nil {
		c.wrapPgError(ctx, err, "failed to retrieve account details")
		return
	}

	// Once we have all that data we can calculate the new expenses next contribution amount.
	if err = spending.CalculateNextContribution(
		account.Timezone,
		fundingSchedule.NextOccurrence,
		fundingSchedule.Rule,
	); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to calculate the next contribution for the new spending")
		return
	}

	if err = repo.CreateExpense(spending); err != nil {
		c.wrapPgError(ctx, err, "failed to create spending")
		return
	}

	ctx.JSON(spending)
}

type SpendingTransfer struct {
	FromSpendingId *uint64 `json:"fromSpendingId"`
	ToSpendingId   *uint64 `json:"toSpendingId"`
	Amount         int64   `json:"amount"`
}

// Transfer To or From Spending
// @id transfer-spending
// @tags Spending
// @summary Transfer allocated funds to or from a spending object.
// @security ApiKeyAuth
// @accept json
// @product json
// @Param bankAccountId path int true "Bank Account ID"
// @Param Spending body SpendingTransfer true "Transfer"
// @Router /bank_accounts/{bankAccountId}/spending/transfer [post]
// @Success 200 {array} models.Spending
// @Failure 400 {object} InvalidBankAccountIdError "Invalid Bank Account ID."
// @Failure 400 {object} ApiError "Malformed JSON or invalid RRule."
// @Failure 500 {object} ApiError "Failed to persist data."
func (c *Controller) postSpendingTransfer(ctx *context.Context) {

}
