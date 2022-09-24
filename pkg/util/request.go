package util

import (
	"strings"

	"github.com/kataras/iris/v12"

	"github.com/google/uuid"
)

func GetRequestID(ctx iris.Context) string {
	values := []string{
		ctx.GetHeader("X-Request-Id"),
		ctx.GetHeader("X-Cloud-Trace-Context"),
	}

	for _, value := range values {
		// The value of the forwared for header can be comma delimited coming from a cloud load balancer.
		items := strings.Split(value, "/")
		if len(items) > 0 && items[0] != "" {
			return items[0]
		}
	}

	if storedRequestId, ok := ctx.Values().Get("X-Request-Id").(string); ok {
		return storedRequestId
	}

	id := uuid.New().String()
	ctx.Values().Set("X-Request-Id", id)

	return id
}
