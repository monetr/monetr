package controller

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	locale "github.com/elliotcourant/go-lclocale"
	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/communication"
	"github.com/monetr/monetr/server/consts"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/powchallenge"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/schemas"
	"github.com/monetr/monetr/server/security"
	"github.com/monetr/monetr/server/zoneinfo"
	"github.com/pkg/errors"
)

const ClearAuthentication = ""

// updateAuthenticationCookie is used to maintain the authentication credentials that the client uses to communicate
// with the API. When this is called with a token that token will be returned to the client in the response as a
// Set-Cookie header. If a blank token is provided then the cookie is updated to expire immediately and the value of the
// cookie is set to blank.
func (c *Controller) updateAuthenticationCookie(ctx echo.Context, token string) {
	sameSite := http.SameSiteDefaultMode
	if c.Configuration.Server.Cookies.SameSiteStrict {
		sameSite = http.SameSiteStrictMode
	}

	expiration := c.Clock.Now().AddDate(0, 0, 14)
	if token == "" {
		expiration = c.Clock.Now().Add(-1 * time.Second)
	}

	if c.Configuration.Server.Cookies.Name == "" {
		panic("authentication cookie name is blank")
	}

	// Set the path to be `/` unless the external URL has specified a prefix. For
	// example, if the external URL is `http://homelab.local/monetr` then we would
	// only want to set cookies for `/monetr` as the path.
	path := c.Configuration.Server.GetBaseURL().Path
	if path == "" {
		path = "/"
	}

	ctx.SetCookie(&http.Cookie{
		Name:     c.Configuration.Server.Cookies.Name,
		Value:    token,
		Path:     path,
		Domain:   c.Configuration.Server.GetHostname(),
		Expires:  expiration,
		MaxAge:   0,
		Secure:   c.Configuration.Server.GetIsCookieSecure(),
		HttpOnly: true,
		SameSite: sameSite,
	})
}

// postChallenge issues a challenge for an unauthenticated auth endpoint, or
// 404s when proof of work is disabled.
func (c *Controller) postChallenge(ctx echo.Context) error {
	c.scrubSentryBody(ctx)
	if !c.Configuration.ProofOfWork.Enabled {
		return c.notFound(ctx, "proof of work is not enabled")
	}

	var request struct {
		Purpose string `json:"purpose"`
	}
	if err := ctx.Bind(&request); err != nil {
		return c.invalidJson(ctx)
	}

	var purpose powchallenge.Purpose
	switch strings.TrimSpace(strings.ToLower(request.Purpose)) {
	case "register":
		purpose = powchallenge.PurposeRegister
	case "login":
		purpose = powchallenge.PurposeLogin
	case "forgot":
		purpose = powchallenge.PurposeForgot
	case "resend":
		purpose = powchallenge.PurposeResend
	default:
		return c.badRequest(ctx, "invalid proof of work purpose")
	}

	challenge, err := c.Challenger.Issue(c.getContext(ctx), purpose)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to issue proof of work challenge")
	}

	return ctx.JSON(http.StatusOK, challenge)
}

func (c *Controller) postLogin(ctx echo.Context) error {
	c.scrubSentryBody(ctx)
	// When proof of work is enabled the request must include a challenge and a
	// nonce, so we validate against the schema that knows about those fields. When
	// it is disabled the plain login schema is used which rejects them outright.
	loginSchema := schemas.LoginSchema
	if c.Configuration.ProofOfWork.Enabled {
		loginSchema = schemas.LoginChallengeSchema
	}
	loginRequest, err := parse(
		c,
		ctx,
		&schemas.LoginRequest{},
		loginSchema,
	)
	if err != nil {
		return err
	}

	// The cheap fail-fast before we do any real work. No-op when disabled.
	if err := c.validateProofOfWork(
		ctx,
		powchallenge.PurposeLogin,
		loginRequest.Challenge,
		loginRequest.Nonce,
	); err != nil {
		return err // validateProofOfWork returns a valid http error.
	}

	secureRepo := c.mustGetSecurityRepository(ctx)
	login, requiresPasswordChange, err := secureRepo.Login(
		c.getContext(ctx),
		loginRequest.Email,
		loginRequest.Password,
	)
	switch errors.Cause(err) {
	case repository.ErrInvalidCredentials:
		return c.returnError(
			ctx,
			http.StatusUnauthorized,
			"Invalid email and password",
		)
	case nil:
		// If no error was returned then do nothing.
	default:
		return c.wrapPgError(ctx, err, "Failed to authenticate")
	}

	// I want to track how many of these types of things we get.
	crumbs.AddTag(
		c.getContext(ctx),
		"requiresPasswordChange",
		fmt.Sprint(requiresPasswordChange),
	)

	log := c.getLog(ctx).With("loginId", login.LoginId)

	// If we want to verify emails and the login does not have a verified email address, then return an error to the
	// user.
	if c.Configuration.Email.ShouldVerifyEmails() && !login.IsEmailVerified {
		log.DebugContext(
			c.getContext(ctx),
			"login email address is not verified, please verify before continuing",
		)
		return c.failure(ctx, http.StatusPreconditionRequired, EmailNotVerifiedError{})
	}

	if requiresPasswordChange {
		// If the server is not configured to allow password resets return an error.
		if !c.Configuration.Email.AllowPasswordReset() {
			return c.returnError(
				ctx,
				http.StatusNotAcceptable,
				"Login requires password reset, but password reset is not allowed",
			)
		}

		passwordResetToken, err := c.ClientTokens.Create(
			5*time.Minute, // Use a much shorter lifetime than usually would be configured.
			security.Claims{
				Scope:        security.ResetPasswordScope,
				EmailAddress: login.Email,
				UserId:       "",
				AccountId:    "",
				LoginId:      login.LoginId.String(),
			},
		)
		if err != nil {
			return c.wrapAndReturnError(
				ctx,
				err,
				http.StatusInternalServerError,
				"Failed to generate a password reset token",
			)
		}
		ctx.Set(authenticationKey, security.Claims{
			LoginId: login.LoginId.String(),
			Scope:   security.ResetPasswordScope,
		})

		log.InfoContext(c.getContext(ctx), "login requires a password change")
		return c.failure(
			ctx,
			http.StatusPreconditionRequired,
			PasswordResetRequiredError{
				ResetToken: passwordResetToken,
			},
		)
	}

	switch len(login.Users) {
	case 0:
		// TODO (elliotcourant) Should we allow them to create an account?
		return c.returnError(
			ctx,
			http.StatusInternalServerError,
			"User has no accounts",
		)
	case 1:
		user := login.Users[0]

		crumbs.IncludeUserInScope(c.getContext(ctx), user.AccountId)

		// Check if the login requires MFA in order to authenticate.
		if login.TOTPEnabledAt != nil {
			log.DebugContext(c.getContext(ctx), "login requires TOTP MFA")
			ctx.Set(authenticationKey, security.Claims{
				LoginId:   login.LoginId.String(),
				AccountId: user.AccountId.String(),
				UserId:    user.UserId.String(),
				Scope:     security.MultiFactorScope,
			})

			token, err := c.ClientTokens.Create(
				5*time.Minute,
				security.Claims{
					Scope:        security.MultiFactorScope,
					EmailAddress: login.Email,
					UserId:       user.UserId.String(),
					AccountId:    user.AccountId.String(),
					LoginId:      user.LoginId.String(),
				},
			)
			if err != nil {
				return c.wrapAndReturnError(
					ctx,
					err,
					http.StatusInternalServerError,
					"Could not generate token",
				)
			}
			c.updateAuthenticationCookie(ctx, token)

			return c.failure(ctx, http.StatusPreconditionRequired, MFARequiredError{})
		}

		ctx.Set(authenticationKey, security.Claims{
			LoginId:   login.LoginId.String(),
			AccountId: user.AccountId.String(),
			UserId:    user.UserId.String(),
			Scope:     security.AuthenticatedScope,
		})

		token, err := c.ClientTokens.Create(
			14*24*time.Hour,
			security.Claims{
				Scope:        security.AuthenticatedScope,
				EmailAddress: login.Email,
				UserId:       user.UserId.String(),
				AccountId:    user.AccountId.String(),
				LoginId:      user.LoginId.String(),
				ReissueCount: 0, // First time the token is being issued
			},
		)
		if err != nil {
			return c.wrapAndReturnError(
				ctx,
				err,
				http.StatusInternalServerError,
				"Could not generate token",
			)
		}

		result := map[string]any{
			"isActive": true,
		}

		c.updateAuthenticationCookie(ctx, token)

		if !c.Configuration.Stripe.IsBillingEnabled() {
			// Return their account token.
			return ctx.JSON(http.StatusOK, result)
		}

		subscriptionIsActive, err := c.Billing.GetSubscriptionIsActive(
			c.getContext(ctx),
			user.AccountId,
		)
		if err != nil {
			return c.wrapAndReturnError(
				ctx,
				err,
				http.StatusInternalServerError,
				"Failed to determine whether or not subscription is active",
			)
		}

		result["isActive"] = subscriptionIsActive

		if !subscriptionIsActive {
			result["nextUrl"] = "/account/subscribe"
		}

		return ctx.JSON(http.StatusOK, result)
	default:
		// If the login has more than one user then we want to generate a temp
		// JWT that will only grant them access to API endpoints not specific to
		// an account.
		return c.badRequest(
			ctx,
			"Multiple accounts not implemented, please contact support",
		)
	}
}

func (c *Controller) postMultifactor(ctx echo.Context) error {
	c.scrubSentryBody(ctx)
	var request struct {
		TOTP string `json:"totp"`
	}
	if err := ctx.Bind(&request); err != nil {
		return c.invalidJson(ctx)
	}

	request.TOTP = strings.TrimSpace(request.TOTP)
	if request.TOTP == "" {
		return c.badRequest(ctx, "TOTP code is required")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	me, err := repo.GetMe(c.getContext(ctx))
	if err != nil {
		return c.unauthorizedError(ctx, err)
	}

	if err := me.Login.VerifyTOTP(request.TOTP, c.Clock.Now()); err != nil {
		return c.returnError(ctx, http.StatusUnauthorized, "Invalid TOTP code")
	}

	token, err := c.ClientTokens.Create(
		14*24*time.Hour,
		security.Claims{
			Scope:        security.AuthenticatedScope,
			EmailAddress: me.Login.Email,
			UserId:       me.UserId.String(),
			AccountId:    me.AccountId.String(),
			LoginId:      me.LoginId.String(),
		},
	)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "could not generate token")
	}

	c.updateAuthenticationCookie(ctx, token)

	result := map[string]any{
		"isActive": true,
	}

	if !c.Configuration.Stripe.IsBillingEnabled() {
		// Return their account token.
		return ctx.JSON(http.StatusOK, result)
	}

	subscriptionIsActive, err := c.Billing.GetSubscriptionIsActive(c.getContext(ctx), me.AccountId)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to determine whether or not subscription is active")
	}

	result["isActive"] = subscriptionIsActive

	if !subscriptionIsActive {
		result["nextUrl"] = "/account/subscribe"
	}

	return ctx.JSON(http.StatusOK, result)
}

func (c *Controller) logoutEndpoint(ctx echo.Context) error {
	if _, err := ctx.Cookie(c.Configuration.Server.Cookies.Name); err == http.ErrNoCookie {
		return ctx.NoContent(http.StatusOK)
	}

	c.updateAuthenticationCookie(ctx, ClearAuthentication)
	return ctx.NoContent(http.StatusOK)
}

func (c *Controller) postRegister(ctx echo.Context) error {
	c.scrubSentryBody(ctx)
	if !c.Configuration.AllowSignUp {
		return c.notFound(ctx, "sign up is not enabled on this server")
	}

	var registerRequest struct {
		Email     string  `json:"email"`
		Password  string  `json:"password"`
		FirstName string  `json:"firstName"`
		LastName  string  `json:"lastName"`
		Timezone  string  `json:"timezone"`
		Locale    string  `json:"locale"`
		BetaCode  *string `json:"betaCode"`
		Challenge string  `json:"challenge"`
		Nonce     uint64  `json:"nonce"`
	}
	if err := ctx.Bind(&registerRequest); err != nil {
		return c.invalidJson(ctx)
	}

	// The cheap fail-fast before we do any real work. No-op when disabled.
	if err := c.validateProofOfWork(
		ctx,
		powchallenge.PurposeRegister,
		registerRequest.Challenge,
		registerRequest.Nonce,
	); err != nil {
		return err // validateProofOfWork returns a valid http error.
	}

	log := c.getLog(ctx)

	registerRequest.Email = strings.TrimSpace(registerRequest.Email)
	registerRequest.FirstName = strings.TrimSpace(registerRequest.FirstName)
	registerRequest.Locale = strings.TrimSpace(registerRequest.Locale)
	registerRequest.Password = strings.TrimSpace(registerRequest.Password)
	if len(registerRequest.Password) > 71 {
		return c.badRequest(ctx, "Password must be less than 72 characters")
	}
	if registerRequest.BetaCode != nil {
		*registerRequest.BetaCode = strings.TrimSpace(*registerRequest.BetaCode)
	}

	if registerRequest.Locale == "" {
		return c.badRequest(ctx, "Locale must be specified to register")
	}

	if _, err := locale.GetLConv(registerRequest.Locale); err != nil {
		log.WarnContext(
			c.getContext(ctx),
			"invalid locale in register request, falling back to global default",
			"locale", registerRequest.Locale,
			"err", err,
		)
		registerRequest.Locale = consts.DefaultLocale
	}

	if err := c.validateRegistration(
		ctx,
		registerRequest.Email,
		registerRequest.Password,
		registerRequest.FirstName,
	); err != nil {
		return err // validateRegistration also returns a valid http error that can just be passed through.
	}

	// If the email's domain is on the sign up blocklist we reject it, but we
	// deliberately return the SAME error a client gets for an email that is
	// already taken so we do not advertise to a would be abuser that their domain
	// is specifically blocked. From the outside a blocked domain is
	// indistinguishable from an address that already exists. The real reason, and
	// which domain it was, is only ever visible to us here in the logs.
	if domain, blocked := c.Configuration.Email.BlockedEmailDomain(
		registerRequest.Email,
	); blocked {
		log.WarnContext(
			c.getContext(ctx),
			"registration blocked, email domain is not allowed",
			"emailDomain", domain,
		)
		return c.failure(ctx, http.StatusBadRequest, EmailAlreadyExists{})
	}

	timezone, err := zoneinfo.Timezone(registerRequest.Timezone)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "failed to parse timezone")
	}

	// If the registration details provided look good then we want to create an
	// unauthenticated repo. This will give us some basic database access
	// without being able to access user information directly. It is essentially
	// a write only interface to the database.
	repo, err := c.getUnauthenticatedRepository(ctx)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError,
			"cannot register user",
		)
	}

	var beta *models.Beta
	if c.Configuration.Beta.EnableBetaCodes {
		if registerRequest.BetaCode == nil || *registerRequest.BetaCode == "" {
			return c.badRequest(ctx, "beta code required for registration")
		}

		beta, err = repo.ValidateBetaCode(c.getContext(ctx), *registerRequest.BetaCode)
		if err != nil {
			return c.wrapPgError(ctx, err, "could not verify beta code")
		}
	}

	// Create the user's login record in the database, this will return the login
	// record including the new login's loginId which we will need below.
	login, err := repo.CreateLogin(
		c.getContext(ctx),
		registerRequest.Email,
		registerRequest.Password,
		registerRequest.FirstName,
		registerRequest.LastName,
	)
	if err != nil {
		switch errors.Cause(err) {
		case repository.ErrEmailAlreadyExists:
			return c.failure(ctx, http.StatusBadRequest, EmailAlreadyExists{})
		default:
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError,
				"failed to create login",
			)
		}
	}
	log = log.With("loginId", login.LoginId)

	var trialEndsAt *time.Time
	if c.Configuration.Stripe.IsBillingEnabled() {
		expiration := c.Clock.Now().AddDate(0, 0, c.Configuration.Stripe.FreeTrialDays)
		log.DebugContext(c.getContext(ctx), "billing is enabled, new account for login will be on a trial", "trialDays", c.Configuration.Stripe.FreeTrialDays, "trialEndsAt", expiration)

		trialEndsAt = &expiration
	}

	account := models.Account{
		Timezone:    timezone.String(),
		TrialEndsAt: trialEndsAt,
		Locale:      registerRequest.Locale,
	}
	// Now that the login exists we can create the account, at the time of
	// writing this we are only using the local time zone of the server, but in
	// the future I want to have it somehow use the user's timezone.
	if err = repo.CreateAccountV2(c.getContext(ctx), &account); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError,
			"failed to create account",
		)
	}

	crumbs.IncludeUserInScope(c.getContext(ctx), account.AccountId)

	user := models.User{
		LoginId:   login.LoginId,
		AccountId: account.AccountId,
		Role:      models.UserRoleOwner,
	}

	// Now that we have an accountId we can create the user object which will
	// bind the login and the account together.
	err = repo.CreateUser(
		c.getContext(ctx),
		&user,
	)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError,
			"failed to create user",
		)
	}

	if beta != nil {
		if err = repo.UseBetaCode(c.getContext(ctx), beta.BetaId, user.UserId); err != nil {
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError,
				"failed to use beta code",
			)
		}
	}

	// If SMTP is enabled and we are verifying emails then we want to create a
	// registration record and send the user a verification email.
	if c.Configuration.Email.ShouldVerifyEmails() {
		verificationToken, err := c.ClientTokens.Create(
			c.Configuration.Email.Verification.TokenLifetime,
			security.Claims{
				Scope:        security.VerifyEmailScope,
				EmailAddress: login.Email,
				UserId:       "",
				AccountId:    "",
				LoginId:      login.LoginId.String(),
			},
		)
		if err != nil {
			return c.wrapAndReturnError(
				ctx,
				err,
				http.StatusInternalServerError,
				"could not generate email verification token",
			)
		}

		if err = c.sendVerificationEmail(
			ctx,
			login,
			verificationToken,
		); err != nil {
			return err
		}

		return ctx.JSON(http.StatusOK, map[string]any{
			"message":             "A verification email has been sent to your email address, please verify your email.",
			"requireVerification": true,
		})
	}

	// If we are not requiring email verification to activate an account we can
	// simply return a token here for the user to be signed in.
	token, err := c.ClientTokens.Create(
		14*24*time.Hour,
		security.Claims{
			Scope:        security.AuthenticatedScope,
			EmailAddress: login.Email,
			UserId:       user.UserId.String(),
			AccountId:    user.AccountId.String(),
			LoginId:      user.LoginId.String(),
		})
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError,
			"failed to create token",
		)
	}

	user.Login = login
	user.Account = &account

	c.updateAuthenticationCookie(ctx, token)

	return ctx.JSON(http.StatusOK, map[string]any{
		"nextUrl":             "/setup",
		"requireVerification": false,
	})
}

func (c *Controller) verifyEndpoint(ctx echo.Context) error {
	c.scrubSentryBody(ctx)
	if !c.Configuration.Email.ShouldVerifyEmails() {
		return c.notFound(ctx, "email verification is not enabled")
	}

	var verifyRequest struct {
		Token string `json:"token"`
	}
	if err := ctx.Bind(&verifyRequest); err != nil {
		return c.invalidJson(ctx)
	}

	if strings.TrimSpace(verifyRequest.Token) == "" {
		return c.badRequest(ctx, "Token cannot be blank")
	}

	claims, err := c.ClientTokens.Parse(verifyRequest.Token)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "Invalid email verification")
	}

	if err := claims.RequireScope(security.VerifyEmailScope); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "Invalid email verification")
	}

	repo := c.mustGetUnauthenticatedRepository(ctx)
	if err := repo.SetEmailVerified(c.getContext(ctx), claims.EmailAddress); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "Invalid email verification")
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"nextUrl": "/login",
		"message": "Your email is now verified. Please login.",
	})
}

func (c *Controller) resendVerification(ctx echo.Context) error {
	c.scrubSentryBody(ctx)
	log := c.getLog(ctx)
	if !c.Configuration.Email.ShouldVerifyEmails() {
		return c.notFound(ctx, "email verification is not enabled")
	}

	var request struct {
		Email     string `json:"email"`
		Challenge string `json:"challenge"`
		Nonce     uint64 `json:"nonce"`
	}
	if err := ctx.Bind(&request); err != nil {
		return c.invalidJson(ctx)
	}

	// Before the captcha, it is the cheap fail-fast. No-op when disabled.
	if err := c.validateProofOfWork(ctx, powchallenge.PurposeResend, request.Challenge, request.Nonce); err != nil {
		return err // validateProofOfWork returns a valid http error.
	}

	request.Email = strings.TrimSpace(strings.ToLower(request.Email))
	if request.Email == "" {
		return c.badRequest(ctx, "email must be provided to resend verification link")
	}

	unauthedRepo := c.mustGetUnauthenticatedRepository(ctx)
	login, err := unauthedRepo.GetLoginForEmail(c.getContext(ctx), request.Email)
	if err != nil {
		log.WarnContext(c.getContext(ctx), "failed to get login for email address to resend verification", "err", err)
		return ctx.NoContent(http.StatusOK)
	}

	verificationToken, err := c.ClientTokens.Create(
		c.Configuration.Email.Verification.TokenLifetime,
		security.Claims{
			Scope:        security.VerifyEmailScope,
			EmailAddress: login.Email,
			UserId:       "",
			AccountId:    "",
			LoginId:      login.LoginId.String(),
		},
	)
	if err != nil {
		c.reportWrappedError(ctx, err, "failed to regenerate email verification token")
		return ctx.NoContent(http.StatusOK)
	}

	if err = c.sendVerificationEmail(
		ctx,
		login,
		verificationToken,
	); err != nil {
		return err
	}

	return ctx.NoContent(http.StatusOK)
}

func (c *Controller) postForgotPassword(ctx echo.Context) error {
	c.scrubSentryBody(ctx)
	if !c.Configuration.Email.AllowPasswordReset() {
		return c.notFound(ctx, "password reset not enabled")
	}

	var sendForgotPasswordRequest struct {
		Email     string `json:"email"`
		Challenge string `json:"challenge"`
		Nonce     uint64 `json:"nonce"`
	}
	if err := ctx.Bind(&sendForgotPasswordRequest); err != nil {
		return c.invalidJson(ctx)
	}

	// Clean up some of the input provided just in case.
	sendForgotPasswordRequest.Email = strings.TrimSpace(strings.ToLower(sendForgotPasswordRequest.Email))

	// Before the captcha, it is the cheap fail-fast. No-op when disabled.
	if err := c.validateProofOfWork(ctx, powchallenge.PurposeForgot, sendForgotPasswordRequest.Challenge, sendForgotPasswordRequest.Nonce); err != nil {
		return err // validateProofOfWork returns a valid http error.
	}

	if sendForgotPasswordRequest.Email == "" {
		return c.badRequest(ctx, "Must provide an email address.")
	}

	// We need to retrieve the login record for the email in order to verify that the login actually exists, as well as
	// make sure the login email address has been verified.
	login, err := c.mustGetUnauthenticatedRepository(ctx).GetLoginForEmail(
		c.getContext(ctx),
		sendForgotPasswordRequest.Email,
	)
	if err != nil {
		crumbs.Debug(c.getContext(ctx), "No password reset email will be sent, login for email not found", map[string]any{
			"error": err.Error(),
		})
		// Don't return an error to the client, don't want them to know if it failed
		// to send.
		return ctx.NoContent(http.StatusOK)
	}

	// If the login's email is not verified then don't send an email verification
	// message.
	if c.Configuration.Email.ShouldVerifyEmails() && !login.GetEmailIsVerified() {
		// TODO Would be best to also sleep for some random amount of time so that
		// it doesn't just seem like we respond instantly.
		return ctx.NoContent(http.StatusOK)
	}

	// Generate the password reset token.
	// TODO When we allow for email address to be changed, we will also need to verify that a token being used is for
	//  the same login that it was generated for. Otherwise someone could generate a token for an email, then change the
	//  email of the login -> change the email of another login to the first email; then reset the password for a
	//  different login. I don't think this is a security risk as the actor would need to have access to both logins to
	//  begin with. But it might cause some goofy issues and is definitely not desired behavior.
	passwordResetToken, err := c.ClientTokens.Create(
		c.Configuration.Email.ForgotPassword.TokenLifetime,
		security.Claims{
			Scope:        security.ResetPasswordScope,
			EmailAddress: login.Email,
			UserId:       "",
			AccountId:    "",
			LoginId:      login.LoginId.String(),
		},
	)
	if err != nil {
		return c.wrapAndReturnError(
			ctx,
			err,
			http.StatusInternalServerError,
			"Failed to generate a password reset token",
		)
	}

	if err = c.sendPasswordReset(
		ctx,
		login,
		passwordResetToken,
	); err != nil {
		return err
	}

	return ctx.NoContent(http.StatusOK)
}

func (c *Controller) resetPassword(ctx echo.Context) error {
	c.scrubSentryBody(ctx)
	// The reset endpoint is intentionally not gated on Email.AllowPasswordReset;
	// the PASETO reset token is the security control here, and operators may mint
	// tokens out-of-band (via the login:reset-password CLI) on deployments where
	// SMTP is intentionally disabled. The forgot-password email-sending endpoint
	// is still gated above, so end users on no-SMTP deployments still can't
	// self-serve.

	var resetPasswordRequest struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := ctx.Bind(&resetPasswordRequest); err != nil {
		return c.invalidJson(ctx)
	}

	resetPasswordRequest.Token = strings.TrimSpace(resetPasswordRequest.Token)
	resetPasswordRequest.Password = strings.TrimSpace(resetPasswordRequest.Password)

	// The token is what verifies that the user is who they say they are even without a password. The token is emailed
	// to their verified email address.
	if resetPasswordRequest.Token == "" {
		return c.badRequest(ctx, "Token must be provided to reset password.")
	}

	if len(resetPasswordRequest.Password) < 8 {
		return c.badRequest(ctx, "Password must be at least 8 characters long.")
	}

	resetClaims, err := c.ClientTokens.Parse(resetPasswordRequest.Token)
	if err != nil {
		return c.badRequestError(ctx, err, "Failed to validate password reset token")
	}

	// Make sure the token has the correct scope on it. Otherwise a user should
	// not be able to use it to reset their password.
	if err := resetClaims.RequireScope(security.ResetPasswordScope); err != nil {
		return c.badRequestError(ctx, err, "Failed to validate password reset token")
	}

	unauthenticatedRepo := c.mustGetUnauthenticatedRepository(ctx)

	// Retrieve the login for the email address in the token.
	login, err := unauthenticatedRepo.GetLoginForEmail(c.getContext(ctx), resetClaims.EmailAddress)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to verify login for email address")
	}

	// If the login's password has been changed since this token was issued, then this token is no longer valid. This
	// will basically make sure that a token cannot be used twice.
	if login.PasswordResetAt != nil && !login.PasswordResetAt.Before(resetClaims.CreatedAt) {
		return c.badRequest(
			ctx,
			"Password has already been reset, you must request another password reset link.",
		)
	}

	if err = unauthenticatedRepo.ResetPassword(
		c.getContext(ctx),
		login.LoginId,
		resetPasswordRequest.Password,
	); err != nil {
		return c.wrapPgError(ctx, err, "Failed to reset password")
	}

	// Only attempt to send the "your password was changed" notification when the
	// email subsystem is actually wired up. Previously the outer
	// AllowPasswordReset gate guaranteed c.Email was non-nil here; now that the
	// endpoint accepts admin-issued tokens regardless of SMTP, we have to be
	// explicit.
	if c.Configuration.Email.Enabled {
		if err := c.Email.SendEmail(
			c.getContext(ctx),
			communication.PasswordChangedParams{
				BaseURL:      c.Configuration.Server.GetBaseURL().String(),
				Email:        login.Email,
				FirstName:    login.FirstName,
				LastName:     login.LastName,
				SupportEmail: "support@monetr.app",
			},
		); err != nil {
			return c.wrapAndReturnError(
				ctx,
				err,
				http.StatusInternalServerError,
				"Failed to send password changed notification",
			)
		}
	}

	return ctx.NoContent(http.StatusOK)
}

func (c *Controller) validateRegistration(ctx echo.Context, email, password, firstName string) error {
	if email == "" {
		return c.badRequest(ctx, "Email cannot be blank")
	}

	if len(password) < 8 {
		return c.badRequest(ctx, "Password must be at least 8 characters")
	}

	if firstName == "" {
		return c.badRequest(ctx, "First name cannot be left blank")
	}

	return nil
}

// validateProofOfWork verifies the challenge on an auth request, mapping typed
// failures to 400s (no-op when disabled). Expired and already-used get their own
// message so a real user can retry; everything else stays vague on purpose.
func (c *Controller) validateProofOfWork(ctx echo.Context, purpose powchallenge.Purpose, challenge string, nonce uint64) error {
	if !c.Configuration.ProofOfWork.Enabled {
		return nil
	}

	err := c.Challenger.Verify(c.getContext(ctx), purpose, challenge, nonce)
	switch errors.Cause(err) {
	case nil:
		return nil
	case powchallenge.ErrChallengeExpired:
		return c.badRequestError(ctx, err, "challenge expired, please retry")
	case powchallenge.ErrChallengeReplay:
		return c.badRequestError(ctx, err, "challenge already used")
	default:
		return c.badRequestError(ctx, err, "invalid proof of work")
	}
}
