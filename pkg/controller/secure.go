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
	"github.com/monetr/monetr/pkg/cache"
	"github.com/monetr/monetr/pkg/communication"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/repository"
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
	p.Post("/challenge", c.secureChallenge)
	p.Post("/authenticate", c.secureAuthenticate)
	p.Post("/register", c.secureRegister)
}

func (c *Controller) secureChallenge(ctx iris.Context) {
	// The client only needs to provide two fields initially. The email address of the login they are trying to
	// authenticate to, and a ReCAPTCHA result (if the configuration requires ReCAPTCHA).
	var challengeRequest struct {
		Email     string `json:"email"`
		ReCAPTCHA string `json:"captcha"`
	}
	if err := ctx.ReadJSON(&challengeRequest); err != nil {
		c.badRequest(ctx, "invalid challenge request provided")
		return
	}

	// Try to look up the login for the provided email address.
	repo := c.mustGetUnauthenticatedRepository(ctx)
	login, err := repo.GetLoginForChallenge(c.getContext(ctx), challengeRequest.Email)
	if err != nil {
		// If we could not verify the email address exists, force the client to fall back to legacy authentication. This
		// will make them fail normally if the credentials truely are bad.
		ctx.JSON(map[string]interface{}{
			"secure": false,
		})
		return
	}

	// If the stuff needed for SRP is missing then the login is still using legacy authentication. Let the client know.
	if login.Verifier == nil && login.Salt == nil {
		ctx.JSON(map[string]interface{}{
			"secure": false,
		})
		return
	}

	// Create the SRP server side. Right now we are using RFC5054Group8192 but this might change in the future. If it
	// does, then we need to add a column to the database to indicate what group was used for each verifier.
	server := srp.NewSRPServer(srp.KnownGroups[srp.RFC5054Group8192], login.GetVerifier(), nil)
	B := server.EphemeralPublic()

	sessionId, err := c.authenticationSessions.CacheAuthenticationSession(c.getContext(ctx), &cache.AuthenticationSession{
		LoginId: login.LoginId,
		SRP:     server,
	})
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to store authentication session")
		return
	}

	ctx.SetCookie(&http.Cookie{
		Name:     c.configuration.Server.Cookies.AuthenticationSessionName,
		Value:    sessionId,
		Domain:   c.configuration.APIDomainName,
		Expires:  time.Now().Add(5 * time.Minute),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	ctx.JSON(map[string]interface{}{
		"secure": true,
		"public": hex.EncodeToString(B.Bytes()),
		"salt":   hex.EncodeToString(login.Salt),
	})
}

func (c *Controller) secureAuthenticate(ctx iris.Context) {
	sessionId := ctx.GetCookie(c.configuration.Server.Cookies.AuthenticationSessionName)
	if sessionId == "" {
		c.returnError(ctx, http.StatusExpectationFailed, "invalid authentication session")
		return
	}

	var authenticateRequest struct {
		Proof  string `json:"proof"`
		Public string `json:"public"`
	}
	if err := ctx.ReadJSON(&authenticateRequest); err != nil {
		c.badRequest(ctx, "invalid challenge request provided")
		return
	}

	public, ok := new(big.Int).SetString(authenticateRequest.Public, 16)
	if !ok {
		c.badRequest(ctx, "invalid client public key provided")
		return
	}

	proof, err := hex.DecodeString(authenticateRequest.Proof)
	if err != nil {
		c.badRequest(ctx, "invalid client proof provided")
		return
	}

	session, err := c.authenticationSessions.LookupAuthenticationSession(
		c.getContext(ctx),
		sessionId,
	)
	if err != nil {
		c.badRequest(ctx, "invalid authentication session")
		return
	}

	server := session.SRP
	if err = server.SetOthersPublic(public); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "invalid client public key")
		return
	}

	if ok = server.GoodClientProof(proof); !ok {
		c.badRequest(ctx, "invalid client proof")
		return
	}

	// At this point the authentication has succeeded. Remove the session cookie, and return a successful status.
	ctx.RemoveCookie(c.configuration.Server.Cookies.AuthenticationSessionName)

	repo := c.mustGetUnauthenticatedRepository(ctx)
	login, err := repo.GetLoginById(c.getContext(ctx), session.LoginId)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to retrieve login details for authentication session")
		return
	}

	log := c.getLog(ctx).WithField("loginId", login.LoginId)

	if c.configuration.Email.ShouldVerifyEmails() && !login.IsEmailVerified {
		log.Debug("login email address is not verified, please verify before continuing")
		c.failure(ctx, http.StatusPreconditionRequired, EmailNotVerified)
		return
	}

	if login.TOTP != "" {
		panic("TOTP for secure authentication not yet implemented")
	}

	switch len(login.Users) {
	case 0:
		c.returnError(ctx, http.StatusInternalServerError, "user has no accounts")
		return
	case 1:
		user := login.Users[0]

		if hub := sentry.GetHubFromContext(c.getContext(ctx)); hub != nil {
			hub.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetUser(sentry.User{
					ID:       strconv.FormatUint(user.AccountId, 10),
					Username: fmt.Sprintf("account:%d", user.AccountId),
				})
			})
		}

		token, err := c.generateToken(login.LoginId, user.UserId, user.AccountId)
		if err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "could not generate JWT")
			return
		}

		c.updateAuthenticationCookie(ctx, token)

		if !c.configuration.Stripe.IsBillingEnabled() {
			// Return their account token.
			ctx.JSON(map[string]interface{}{
				"isActive": true,
			})
			return
		}

		subscriptionIsActive, err := c.paywall.GetSubscriptionIsActive(c.getContext(ctx), user.AccountId)
		if err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to determine whether or not subscription is active")
			return
		}

		result := map[string]interface{}{
			"isActive": subscriptionIsActive,
		}

		if !subscriptionIsActive {
			result["nextUrl"] = "/account/subscribe"
		}

		ctx.JSON(result)
	default:
		// If the login has more than one user then we want to generate a temp
		// JWT that will only grant them access to API endpoints not specific to
		// an account.
		token, err := c.generateToken(login.LoginId, 0, 0)
		if err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "could not generate token")
			return
		}

		c.updateAuthenticationCookie(ctx, token)

		ctx.JSON(map[string]interface{}{
			"users": login.Users,
		})
	}
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
		if errors.Is(errors.Cause(err), repository.ErrEmailAlreadyExists) {
			c.badRequest(ctx, "a login with this email already exists")
			return
		}

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
