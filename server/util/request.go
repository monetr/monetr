package util

import (
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func GetRequestID(ctx echo.Context) string {
	values := []string{
		ctx.Request().Header.Get("X-Request-Id"),
		ctx.Request().Header.Get("X-Cloud-Trace-Context"),
	}

	for _, value := range values {
		// The value of the forwared for header can be comma delimited coming from a cloud load balancer.
		items := strings.Split(value, "/")
		if len(items) > 0 && items[0] != "" {
			return items[0]
		}
	}

	if storedRequestId, ok := ctx.Get("X-Request-Id").(string); ok {
		return storedRequestId
	}

	// TODO Get rid of this and replace it with some kind of middleware or fall
	// back to ULID generation.
	id := uuid.New().String()
	ctx.Set("X-Request-Id", id)

	return id
}
