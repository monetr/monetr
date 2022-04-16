package controller

import (
	"encoding"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/1Password/srp"
	"github.com/getsentry/sentry-go"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
	"github.com/monetr/monetr/pkg/build"
	"github.com/monetr/monetr/pkg/communication"
	"github.com/monetr/monetr/pkg/models"
	"github.com/pkg/errors"
	"github.com/stripe/stripe-go/v72"
)

var (
	// This is needed to make sure we can store the SRP session.
	_ encoding.BinaryMarshaler   = &srp.SRP{}
	_ encoding.BinaryUnmarshaler = &srp.SRP{}
)

func hexToBigInt(input string) (*big.Int, error) {
	result, ok := (&big.Int{}).SetString(input, 16)
	if !ok {
		return nil, errors.New("bad hexadecimal big integer")
	}
	return result, nil
}

func (c *Controller) handleSecureAuthentication(p router.Party) {
	p.Post("/register", c.secureRegister)
}

func (c *Controller) secureRegister(ctx iris.Context) {
	var registerRequest struct {
		Email     string  `json:"email"`
		FirstName string  `json:"firstName"`
		LastName  string  `json:"lastName"`
		Timezone  string  `json:"timezone"`
		Agree     bool    `json:"agree"`
		Verifier  string  `json:"verifier"`
		Salt      string  `json:"salt"`
		Captcha   string  `json:"captcha"`
		BetaCode  *string `json:"betaCode"`
	}
	if err := ctx.ReadJSON(&registerRequest); err != nil {
		c.badRequest(ctx, "failed to parse request")
		return
	}

	// If ReCAPTCHA is required for registration then this will validate the input. If it is not required then this will
	// do nothing.
	if err := c.validateRegistrationCaptcha(c.getContext(ctx), registerRequest.Captcha); err != nil {
		c.badRequest(ctx, "invalid ReCAPTCHA provided")
		return
	}

	registerRequest.Email = strings.TrimSpace(registerRequest.Email)
	registerRequest.FirstName = strings.TrimSpace(registerRequest.FirstName)
	if registerRequest.BetaCode != nil {
		*registerRequest.BetaCode = strings.TrimSpace(*registerRequest.BetaCode)
	}

	{ // Parse the email address provided to make sure it is valid.
		address, err := mail.ParseAddress(registerRequest.Email)
		if err != nil {
			c.badRequest(ctx, "email address provided is not valid")
			return
		}

		// If the parsed email address does not equal the provided one then something is definitely wrong.
		if !strings.EqualFold(address.Address, registerRequest.Email) {
			// I have no idea what would actually produce this code path.
			c.badRequest(ctx, "could not verify that the provided email address was legitimate")
			return
		}
	}

	if registerRequest.FirstName == "" {
		c.badRequest(ctx, "first name must be provided")
		return
	}

	timezone, err := time.LoadLocation(registerRequest.Timezone)
	if err != nil {
		c.badRequest(ctx, "valid timezone must be provided")
		return
	}

	salt, err := hex.DecodeString(registerRequest.Salt)
	if err != nil {
		c.badRequest(ctx, "invalid salt provided")
		return
	} else if len(salt) < 32 {
		// The salt should be at least 32 bytes.
		c.badRequest(ctx, "salt is not strong enough")
		return
	}

	// We have no way of actually knowing how strong someone's password is anymore, or even how long it is. Going to
	// take a shot in the dark and say that if the encrypted verifier is less than 32 characters then something is
	// definitely wrong.
	if len(registerRequest.Verifier) < 32 {
		c.badRequest(ctx, "password is not strong enough")
		return
	}

	verifier, err := hexToBigInt(registerRequest.Verifier)
	if err != nil {
		c.badRequest(ctx, "invalid verifier provided")
		return
	}

	repo := c.mustGetUnauthenticatedRepository(ctx)

	var beta *models.Beta
	if c.configuration.Beta.EnableBetaCodes {
		if registerRequest.BetaCode == nil || *registerRequest.BetaCode == "" {
			c.badRequest(ctx, "beta code required for registration")
			return
		}

		beta, err = repo.ValidateBetaCode(c.getContext(ctx), *registerRequest.BetaCode)
		if err != nil {
			c.wrapPgError(ctx, err, "could not verify beta code")
			return
		}
	}

	login := models.LoginWithVerifier{
		Login: models.Login{
			Email:     registerRequest.Email,
			FirstName: registerRequest.FirstName,
			LastName:  registerRequest.LastName,
			IsEnabled: true,
		},
		Verifier: verifier.Bytes(),
		Salt:     salt,
	}

	if err = repo.CreateSecureLogin(c.getContext(ctx), &login); err != nil {
		c.wrapPgError(ctx, err, "failed to register")
		return
	}

	var stripeCustomerId *string
	if c.configuration.Stripe.Enabled {
		c.log.Debug("creating stripe customer for new user")
		name := registerRequest.FirstName + " " + registerRequest.LastName
		result, err := c.stripe.CreateCustomer(c.getContext(ctx), stripe.CustomerParams{
			Email: &registerRequest.Email,
			Name:  &name,
			Params: stripe.Params{
				Metadata: map[string]string{
					"environment": c.configuration.Environment,
					"revision":    build.Revision,
					"release":     build.Release,
				},
			},
		})
		if err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create stripe customer")
			return
		}

		stripeCustomerId = &result.ID
	}

	account := models.Account{
		Timezone:             timezone.String(),
		StripeCustomerId:     stripeCustomerId,
		StripeSubscriptionId: nil,
	}
	// Now that the login exists we can create the account, at the time of
	// writing this we are only using the local time zone of the server, but in
	// the future I want to have it somehow use the user's timezone.
	if err = repo.CreateAccountV2(c.getContext(ctx), &account); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError,
			"failed to create account",
		)
		return
	}

	if hub := sentry.GetHubFromContext(c.getContext(ctx)); hub != nil {
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetUser(sentry.User{
				ID:       strconv.FormatUint(account.AccountId, 10),
				Username: fmt.Sprintf("account:%d", account.AccountId),
			})
		})
	}

	user := models.User{
		LoginId:          login.LoginId,
		AccountId:        account.AccountId,
		FirstName:        registerRequest.FirstName,
		LastName:         registerRequest.LastName,
		StripeCustomerId: stripeCustomerId,
	}

	// Now that we have an accountId we can create the user object which will
	// bind the login and the account together.
	err = repo.CreateUser(
		c.getContext(ctx),
		login.LoginId,
		account.AccountId,
		&user,
	)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError,
			"failed to create user",
		)
		return
	}

	if beta != nil {
		if err = repo.UseBetaCode(c.getContext(ctx), beta.BetaID, user.UserId); err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError,
				"failed to use beta code",
			)
			return
		}
	}

	// If SMTP is enabled and we are verifying emails then we want to create a
	// registration record and send the user a verification email.
	if c.configuration.Email.ShouldVerifyEmails() {
		verificationToken, err := c.emailVerification.CreateEmailVerificationToken(c.getContext(ctx), registerRequest.Email)
		if err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "could not generate email verification token")
			return
		}

		if err = c.communication.SendVerificationEmail(c.getContext(ctx), communication.VerifyEmailParams{
			Login: login.Login,
			VerifyURL: fmt.Sprintf("%s/verify/email?token=%s",
				c.configuration.GetUIURL(),
				url.QueryEscape(verificationToken),
			),
		}); err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to send verification email")
			return
		}

		ctx.JSON(map[string]interface{}{
			"message":             "A verification email has been sent to your email address, please verify your email.",
			"requireVerification": true,
		})
		return
	}

	// If we are not requiring email verification to activate an account we can
	// simply return a token here for the user to be signed in.
	token, err := c.generateToken(login.LoginId, user.UserId, account.AccountId)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError,
			"failed to create JWT",
		)
		return
	}

	user.Login = &login.Login
	user.Account = &account

	c.updateAuthenticationCookie(ctx, token)

	if !c.configuration.Stripe.IsBillingEnabled() {
		ctx.JSON(map[string]interface{}{
			"nextUrl":             "/setup",
			"user":                user,
			"isActive":            true,
			"requireVerification": false,
		})
		return
	}

	ctx.JSON(map[string]interface{}{
		"nextUrl":             "/account/subscribe",
		"user":                user,
		"isActive":            false,
		"requireVerification": false,
	})
	return
}
