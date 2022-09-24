package util

import (
	"strings"

	"github.com/kataras/iris/v12"
)

// GetForwardedFor will return the IP address provided by the request header X-Forwarded-For or X-Real-Ip.
func GetForwardedFor(ctx iris.Context) string {
	values := []string{
		ctx.GetHeader("X-Forwarded-For"),
		ctx.GetHeader("X-Real-Ip"),
	}
	for _, value := range values {
		// The value of the forwared for header can be comma delimited coming from a cloud load balancer.
		items := strings.Split(value, ",")
		if len(items) > 0 && items[0] != "" {
			return items[0]
		}
	}

	return ""
}
