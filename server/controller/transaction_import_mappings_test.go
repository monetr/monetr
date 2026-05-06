package controller_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/monetr/monetr/server/internal/fixtures"
)

func TestPostTransactionImportMapping(t *testing.T) {
	t.Run("empty mapping returns the full structured validation error", func(t *testing.T) {
		// An empty mapping triggers every required-field error across every
		// sub-spec. The response body is asserted as a complete map so that any
		// drift in the JSON shape (added/removed fields, wrong nesting, lost oneOf
		// envelope) trips this test.
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.POST("/api/mappings").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"mapping": map[string]any{
					"amount": map[string]any{
						"kind": "test",
						"fields": []any{
							map[string]any{
								"name": "foo",
							},
						},
					},
					"headers": []string{
						"test",
					},
				},
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Object().IsEqual(map[string]any{
			"error": "Invalid request",
			"problems": map[string]any{
				"amount": map[string]any{
					"oneOf": []any{
						map[string]any{
							"fields": map[string]any{
								"0": map[string]any{
									"oneOf": []any{
										map[string]any{
											"name": `must be one of: ["test"]`,
										},
										map[string]any{
											"derivedKind": "cannot be blank",
											"name":        "must be blank",
										},
									},
								},
							},
							"kind": `must equal "sign"`,
						},
						map[string]any{
							"credit": "cannot be blank",
							"debit":  "cannot be blank",
							"fields": map[string]any{
								"0": map[string]any{
									"oneOf": []any{
										map[string]any{
											"name": `must be one of: ["test"]`,
										},
										map[string]any{
											"derivedKind": "cannot be blank",
											"name":        "must be blank",
										},
									},
								},
							},
							"kind": `must equal "type"`,
						},
						map[string]any{
							"fields": map[string]any{
								"0": map[string]any{
									"oneOf": []any{
										map[string]any{
											"name": `must be one of: ["test"]`,
										},
										map[string]any{
											"derivedKind": "cannot be blank",
											"name":        "must be blank",
										},
									},
								},
							},
							"kind": `must equal "column"`,
						},
					},
				},
				"balance": map[string]any{
					"oneOf": []any{
						map[string]any{
							"kind": "cannot be blank",
						},
						map[string]any{
							"fields": "cannot be blank",
							"kind":   `must equal "field"`,
						},
					},
				},
				"date": map[string]any{
					"fields": "cannot be blank",
					"format": "cannot be blank",
				},
				"id": map[string]any{
					"fields": "cannot be blank",
					"kind":   "cannot be blank",
				},
				"memo": map[string]any{
					"oneOf": []any{
						map[string]any{
							"name": "cannot be blank",
						},
						map[string]any{
							"derivedKind": "cannot be blank",
						},
					},
				},
			},
		})
	})

	t.Run("valid mapping is accepted", func(t *testing.T) {
		// A fully consistent mapping satisfies one variant of every union and every
		// required field, with FieldRef names referencing entries in Headers. The
		// endpoint does not yet persist or echo the mapping back, so the assertion
		// here is just that validation passes (200, no problems body).
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.POST("/api/mappings").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"mapping": map[string]any{
					"id": map[string]any{
						"kind": "native",
						"fields": []any{
							map[string]any{
								"name": "Id",
							},
						},
					},
					"amount": map[string]any{
						"kind": "sign",
						"fields": []any{
							map[string]any{
								"name": "Amount",
							},
						},
					},
					"memo": map[string]any{
						"name": "Description",
					},
					"date": map[string]any{
						"fields": []any{
							map[string]any{
								"name": "Date",
							},
						},
						"format": "YYYY-MM-DD",
					},
					"balance": map[string]any{
						"kind": "none",
					},
					"headers": []string{
						"Date",
						"Description",
						"Amount",
						"Id",
					},
				},
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.transactionImportMappingId").String().NotEmpty()
		response.JSON().Path("$.signature").String().NotEmpty()
	})
}

func TestGetTransactionImportMappings(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.GET("/api/mappings").
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Array().IsEmpty()
	})

	t.Run("returns mappings ordered most-recent first", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		token := GivenILogin(t, e, user.Login.Email, password)

		first := e.POST("/api/mappings").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"mapping": map[string]any{
					"id":      map[string]any{"kind": "native", "fields": []any{map[string]any{"name": "Id"}}},
					"amount":  map[string]any{"kind": "sign", "fields": []any{map[string]any{"name": "Amount"}}},
					"memo":    map[string]any{"name": "Description"},
					"date":    map[string]any{"fields": []any{map[string]any{"name": "Date"}}, "format": "YYYY-MM-DD"},
					"balance": map[string]any{"kind": "none"},
					"headers": []string{"Date", "Description", "Amount", "Id"},
				},
			}).
			Expect()
		first.Status(http.StatusOK)
		firstId := first.JSON().Path("$.transactionImportMappingId").String().Raw()

		app.Clock.Add(1 * time.Minute)

		second := e.POST("/api/mappings").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"mapping": map[string]any{
					"id":      map[string]any{"kind": "native", "fields": []any{map[string]any{"name": "Reference"}}},
					"amount":  map[string]any{"kind": "sign", "fields": []any{map[string]any{"name": "Value"}}},
					"memo":    map[string]any{"name": "Memo"},
					"date":    map[string]any{"fields": []any{map[string]any{"name": "Posted"}}, "format": "YYYY-MM-DD"},
					"balance": map[string]any{"kind": "none"},
					"headers": []string{"Posted", "Memo", "Value", "Reference"},
				},
			}).
			Expect()
		second.Status(http.StatusOK)
		secondId := second.JSON().Path("$.transactionImportMappingId").String().Raw()

		response := e.GET("/api/mappings").
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Array().Length().IsEqual(2)
		response.JSON().Path("$[0].transactionImportMappingId").IsEqual(secondId)
		response.JSON().Path("$[1].transactionImportMappingId").IsEqual(firstId)
	})

	t.Run("filters by signature", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		token := GivenILogin(t, e, user.Login.Email, password)

		matching := e.POST("/api/mappings").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"mapping": map[string]any{
					"id":      map[string]any{"kind": "native", "fields": []any{map[string]any{"name": "Id"}}},
					"amount":  map[string]any{"kind": "sign", "fields": []any{map[string]any{"name": "Amount"}}},
					"memo":    map[string]any{"name": "Description"},
					"date":    map[string]any{"fields": []any{map[string]any{"name": "Date"}}, "format": "YYYY-MM-DD"},
					"balance": map[string]any{"kind": "none"},
					"headers": []string{"Date", "Description", "Amount", "Id"},
				},
			}).
			Expect()
		matching.Status(http.StatusOK)
		matchingId := matching.JSON().Path("$.transactionImportMappingId").String().Raw()
		signature := matching.JSON().Path("$.signature").String().Raw()

		other := e.POST("/api/mappings").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"mapping": map[string]any{
					"id":      map[string]any{"kind": "native", "fields": []any{map[string]any{"name": "Reference"}}},
					"amount":  map[string]any{"kind": "sign", "fields": []any{map[string]any{"name": "Value"}}},
					"memo":    map[string]any{"name": "Memo"},
					"date":    map[string]any{"fields": []any{map[string]any{"name": "Posted"}}, "format": "YYYY-MM-DD"},
					"balance": map[string]any{"kind": "none"},
					"headers": []string{"Posted", "Memo", "Value", "Reference"},
				},
			}).
			Expect()
		other.Status(http.StatusOK)

		response := e.GET("/api/mappings").
			WithQuery("signature", signature).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Array().Length().IsEqual(1)
		response.JSON().Path("$[0].transactionImportMappingId").IsEqual(matchingId)
		response.JSON().Path("$[0].signature").IsEqual(signature)
	})

	t.Run("paginates with limit and offset", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		token := GivenILogin(t, e, user.Login.Email, password)

		for i := 0; i < 3; i++ {
			response := e.POST("/api/mappings").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"mapping": map[string]any{
						"id":      map[string]any{"kind": "native", "fields": []any{map[string]any{"name": "Id"}}},
						"amount":  map[string]any{"kind": "sign", "fields": []any{map[string]any{"name": "Amount"}}},
						"memo":    map[string]any{"name": "Description"},
						"date":    map[string]any{"fields": []any{map[string]any{"name": "Date"}}, "format": "YYYY-MM-DD"},
						"balance": map[string]any{"kind": "none"},
						"headers": []string{"Date", "Description", "Amount", "Id"},
					},
				}).
				Expect()
			response.Status(http.StatusOK)
			app.Clock.Add(1 * time.Minute)
		}

		page1 := e.GET("/api/mappings").
			WithQuery("limit", 2).
			WithCookie(TestCookieName, token).
			Expect()
		page1.Status(http.StatusOK)
		page1.JSON().Array().Length().IsEqual(2)

		page2 := e.GET("/api/mappings").
			WithQuery("limit", 2).
			WithQuery("offset", 2).
			WithCookie(TestCookieName, token).
			Expect()
		page2.Status(http.StatusOK)
		page2.JSON().Array().Length().IsEqual(1)
	})

	t.Run("rejects limit above 10", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.GET("/api/mappings").
			WithQuery("limit", 11).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").IsEqual("limit cannot be greater than 10")
	})

	t.Run("rejects limit below 1", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.GET("/api/mappings").
			WithQuery("limit", 0).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").IsEqual("limit must be at least 1")
	})

	t.Run("rejects negative offset", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.GET("/api/mappings").
			WithQuery("offset", -1).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").IsEqual("offset cannot be less than 0")
	})

	t.Run("isolates by account", func(t *testing.T) {
		app, e := NewTestApplication(t)
		userA, passwordA := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		tokenA := GivenILogin(t, e, userA.Login.Email, passwordA)

		seedA := e.POST("/api/mappings").
			WithCookie(TestCookieName, tokenA).
			WithJSON(map[string]any{
				"mapping": map[string]any{
					"id":      map[string]any{"kind": "native", "fields": []any{map[string]any{"name": "Id"}}},
					"amount":  map[string]any{"kind": "sign", "fields": []any{map[string]any{"name": "Amount"}}},
					"memo":    map[string]any{"name": "Description"},
					"date":    map[string]any{"fields": []any{map[string]any{"name": "Date"}}, "format": "YYYY-MM-DD"},
					"balance": map[string]any{"kind": "none"},
					"headers": []string{"Date", "Description", "Amount", "Id"},
				},
			}).
			Expect()
		seedA.Status(http.StatusOK)

		userB, passwordB := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		tokenB := GivenILogin(t, e, userB.Login.Email, passwordB)

		response := e.GET("/api/mappings").
			WithCookie(TestCookieName, tokenB).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Array().IsEmpty()
	})
}
