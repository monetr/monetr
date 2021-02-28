package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"net/http"
)

func (c *Controller) handleJobs(p iris.Party) {
	p.Get("/{jobId:string}", func(ctx *context.Context) {
		jobId := ctx.Params().GetStringTrim("jobId")
		if jobId == "" {
			c.returnError(ctx, http.StatusBadRequest, "must provide a job Id")
			return
		}

		repo := c.mustGetAuthenticatedRepository(ctx)

		job, err := repo.GetJob(jobId)
		if err != nil {
			c.wrapPgError(ctx, err, "failed to retrieve job")
			return
		}

		ctx.JSON(map[string]interface{}{
			"job": job,
		})
	})
}
