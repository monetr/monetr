package controller_test

import (
	"net/http"
	"testing"
)

func TestGetConfig(t *testing.T) {
	_, e := NewTestApplication(t)
	response := e.GET("/api/config").Expect()

	response.Status(http.StatusOK)
	j := response.JSON()
	j.Path("$.requireLegalName").Boolean().IsFalse()
	j.Path("$.requirePhoneNumber").Boolean().IsFalse()
	j.Path("$.verifyLogin").Boolean().IsFalse()
	j.Path("$.verifyRegister").Boolean().IsFalse()
	j.Path("$.allowSignUp").Boolean().IsTrue()
	j.Path("$.allowForgotPassword").Boolean().IsFalse()
}
