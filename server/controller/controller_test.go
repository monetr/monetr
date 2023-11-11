package controller_test

import (
	"net/http"
	"testing"
)

func TestGetHealth(t *testing.T) {
	_, e := NewTestApplication(t)
	response := e.GET("/api/health").Expect()

	response.Status(http.StatusOK)
	response.JSON().Path("$.dbHealthy").Boolean().True()
	response.JSON().Path("$.apiHealthy").Boolean().True()
}
