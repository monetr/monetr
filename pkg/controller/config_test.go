package controller_test

import (
	"net/http"
	"testing"
)

func TestGetConfig(t *testing.T) {
	e := NewTestApplication(t)
	response := e.GET("/config").Expect()

	response.Status(http.StatusOK)
	j := response.JSON()
	j.Path("$.requireLegalName").Boolean().False()
	j.Path("$.requirePhoneNumber").Boolean().False()
	j.Path("$.verifyLogin").Boolean().False()
	j.Path("$.verifyRegister").Boolean().False()
	j.Path("$.allowSignUp").Boolean().True()
	j.Path("$.allowForgotPassword").Boolean().False()
}
