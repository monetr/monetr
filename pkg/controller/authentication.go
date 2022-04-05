package controller

import (
	"context"
	"fmt"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/form3tech-oss/jwt-go"
	"github.com/getsentry/sentry-go"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
	"github.com/monetr/monetr/pkg/build"
	"github.com/monetr/monetr/pkg/communication"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/hash"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/monetr/monetr/pkg/swag"
	"github.com/pkg/errors"
	"github.com/stripe/stripe-go/v72"
)

type MonetrClaims struct {
	LoginId   uint64 `json:"loginId"`
	UserId    uint64 `json:"userId"`
	AccountId uint64 `json:"accountId"`
	jwt.StandardClaims
}

const ClearAuthentication = ""

// updateAuthenticationCookie is used to maintain the authentication credentials that the client uses to communicate
// with the API. When this is called with a token that token will be returned to the client in the response as a
// Set-Cookie header. If a blank token is provided then the cookie is updated to expire immediately and the value of the
// cookie is set to blank.
func (c *Controller) updateAuthenticationCookie(ctx iris.Context, token string) {
	sameSite := http.SameSiteDefaultMode
	if c.configuration.Server.Cookies.SameSiteStrict {
		sameSite = http.SameSiteStrictMode
	}

	expiration := time.Now().Add(7 * 24 * time.Hour)
	if token == "" {
		expiration = time.Now().Add(-1 * time.Second)
	}

	if c.configuration.Server.Cookies.Name == "" {
		panic("authentication cookie name is blank")
	}

	ctx.SetCookie(&http.Cookie{
		Name:     c.configuration.Server.Cookies.Name,
		Value:    token,
		Path:     "/",
		Domain:   c.configuration.APIDomainName,
		Expires:  expiration,
		MaxAge:   0,
		Secure:   c.configuration.Server.Cookies.Secure,
		HttpOnly: true,
		SameSite: sameSite,
	})
}

func (c *Controller) handleAuthentication(p router.Party) {
	p.Post("/login", c.loginEndpoint)
	p.Get("/logout", c.logoutEndpoint)

	if c.configuration.AllowSignUp {
		p.Post("/register", c.registerEndpoint)
	}

	if c.configuration.Email.ShouldVerifyEmails() {
		p.Post("/verify", c.verifyEndpoint)
		p.Post("/verify/resend", c.resendVerification)
	}

	if c.configuration.Email.AllowPasswordReset() {
		p.Post("/forgot", c.sendForgotPassword)
		p.Post("/reset", c.resetPassword)
	}
}

// Login
// @Summary Login
// @id login
// @tags Authentication
// @description Authenticate a user.
// @Accept json
// @Produce json
// @Param Login body swag.LoginRequest true "User Login Request"
// @Router /authentication/login [post]
// @Success 200 {object} swag.LoginResponse
// @Failure 400 {object} swag.LoginInvalidRequestResponse Required data is missing.
// @Failure 401 {object} swag.LoginInvalidCredentialsResponse Invalid credentials.
// @Failure 428 {object} swag.LoginPreconditionRequiredResponse Login requirements are missing.
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) loginEndpoint(ctx iris.Context) {
	var loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Captcha  string `json:"captcha"`
		TOTP     string `json:"totp"`
	}
	if err := ctx.ReadJSON(&loginRequest); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed json")
		return
	}

	// This will take the captcha from the request and validate it if the API is
	// configured to do so. If it is enabled and the captcha fails then an error
	// is returned to the client.

	if err := c.validateLoginCaptcha(c.getContext(ctx), loginRequest.Captcha); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "valid ReCAPTCHA is required")
		return
	}

	loginRequest.Email = strings.ToLower(strings.TrimSpace(loginRequest.Email))
	loginRequest.Password = strings.TrimSpace(loginRequest.Password)
	loginRequest.TOTP = strings.TrimSpace(loginRequest.TOTP)

	if err := c.validateLogin(loginRequest.Email, loginRequest.Password); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "login is not valid")
		return
	}

	hashedPassword := hash.HashPassword(loginRequest.Email, loginRequest.Password)

	secureRepo := c.mustGetSecurityRepository(ctx)
	login, err := secureRepo.Login(c.getContext(ctx), loginRequest.Email, hashedPassword)
	switch errors.Cause(err) {
	case repository.ErrInvalidCredentials:
		c.returnError(ctx, http.StatusUnauthorized, "invalid email and password")
		return
	case nil:
		// If no error was returned then do nothing.
		break
	default:
		c.wrapPgError(ctx, err, "failed to authenticate")
		return
	}

	log := c.getLog(ctx).WithField("loginId", login.LoginId)

	// If we want to verify emails and the login does not have a verified email address, then return an error to the
	// user.
	if c.configuration.Email.ShouldVerifyEmails() && !login.IsEmailVerified {
		log.Debug("login email address is not verified, please verify before continuing")
		c.failure(ctx, http.StatusPreconditionRequired, EmailNotVerifiedError{})
		return
	}

	// Check if the login requires MFA in order to authenticate.
	if login.TOTP != "" && loginRequest.TOTP == "" {
		log.Debug("login requires TOTP MFA, but none was provided")
		c.failure(ctx, http.StatusPreconditionRequired, MFARequiredError{})
		return
	} else if login.TOTP != "" && loginRequest.TOTP != "" {
		// If the login does require TOTP and a code was provided in the request, then validate that the provided code is
		// correct.
		log.Trace("login requires TOTP MFA, and a code was provided; it will be verified")

		if err := login.VerifyTOTP(loginRequest.TOTP); err != nil {
			log.Trace("provided TOTP MFA code is not valid")
			c.returnError(ctx, http.StatusUnauthorized, "invalid TOTP code")
			return
		}

		log.Trace("provided TOTP MFA code is valid")
	} else if login.TOTP == "" && loginRequest.TOTP != "" {
		log.Warn("login does not require TOTP MFA, but a code was provided anyway")
	}

	switch len(login.Users) {
	case 0:
		// TODO (elliotcourant) Should we allow them to create an account?
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
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "could not generate JWT")
			return
		}

		c.updateAuthenticationCookie(ctx, token)

		ctx.JSON(map[string]interface{}{
			"users": login.Users,
		})
	}
}

// Logout
// @Summary Logout
// @id logout
// @tags Authentication
// @description Removes the `HttpOnly` authentication cookie for the user. This request does require the authentication
// @description cookie is present. It will return a `403` if the cookie is not present.
// @Router /authentication/logout [get]
// @Security ApiKeyAuth
// @Success 200
// @Failure 401 {object} ApiError Cookie was not present.
func (c *Controller) logoutEndpoint(ctx iris.Context) {
	if cookie := ctx.GetCookie(c.configuration.Server.Cookies.Name); cookie == "" {
		c.returnError(ctx, http.StatusUnauthorized, "authentication required")
		return
	}

	c.updateAuthenticationCookie(ctx, ClearAuthentication)
}

// Register
// @Summary Register
// @id register
// @tags Authentication
// @description Register creates a new login, user and account. Logins are used for authentication, users tie authentication to an account, and accounts hold budgeting data.
// @Produce json
// @Accept json
// @Param Registration body swag.RegisterRequest true "New User Registration"
// @Router /authentication/register [post]
// @Success 200 {object} swag.RegisterResponse
// @Failure 400 {object} ApiError Required data is missing.
// @Failure 401 {object} ApiError Invalid credentials.
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) registerEndpoint(ctx iris.Context) {
	var registerRequest struct {
		Email     string  `json:"email"`
		Password  string  `json:"password"`
		FirstName string  `json:"firstName"`
		LastName  string  `json:"lastName"`
		Timezone  string  `json:"timezone"`
		Captcha   string  `json:"captcha"`
		BetaCode  *string `json:"betaCode"`
		Agree     bool    `json:"agree"`
	}
	if err := ctx.ReadJSON(&registerRequest); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "invalid register JSON")
		return
	}

	// This will take the captcha from the request and validate it if the API is
	// configured to do so. If it is enabled and the captcha fails then an error
	// is returned to the client.
	if err := c.validateRegistrationCaptcha(c.getContext(ctx), registerRequest.Captcha); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "valid ReCAPTCHA is required")
		return
	}

	registerRequest.Email = strings.TrimSpace(registerRequest.Email)
	registerRequest.Password = strings.TrimSpace(registerRequest.Password)
	registerRequest.FirstName = strings.TrimSpace(registerRequest.FirstName)
	if registerRequest.BetaCode != nil {
		*registerRequest.BetaCode = strings.TrimSpace(*registerRequest.BetaCode)
	}

	if err := c.validateRegistration(
		registerRequest.Email,
		registerRequest.Password,
		registerRequest.FirstName,
	); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest,
			"invalid registration",
		)
		return
	}

	timezone, err := time.LoadLocation(registerRequest.Timezone)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "failed to parse timezone")
		return
	}

	// If the registration details provided look good then we want to create an
	// unauthenticated repository. This will give us some basic database access
	// without being able to access user information directly. It is essentially
	// a write only interface to the database.
	repository, err := c.getUnauthenticatedRepository(ctx)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError,
			"cannot register user",
		)
		return
	}

	var beta *models.Beta
	if c.configuration.Beta.EnableBetaCodes {
		if registerRequest.BetaCode == nil || *registerRequest.BetaCode == "" {
			c.badRequest(ctx, "beta code required for registration")
			return
		}

		beta, err = repository.ValidateBetaCode(c.getContext(ctx), *registerRequest.BetaCode)
		if err != nil {
			c.wrapPgError(ctx, err, "could not verify beta code")
			return
		}
	}

	// Hash the user's password so that we can store it securely.
	hashedPassword := hash.HashPassword(
		registerRequest.Email, registerRequest.Password,
	)

	// Create the user's login record in the database, this will return the login
	// record including the new login's loginId which we will need below.
	login, err := repository.CreateLogin(
		c.getContext(ctx),
		registerRequest.Email,
		hashedPassword,
		registerRequest.FirstName,
		registerRequest.LastName,
	)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError,
			"failed to create login",
		)
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
	if err = repository.CreateAccountV2(c.getContext(ctx), &account); err != nil {
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
	err = repository.CreateUser(
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
		if err = repository.UseBetaCode(c.getContext(ctx), beta.BetaID, user.UserId); err != nil {
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
			Login: *login,
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

	user.Login = login
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

// Verify Email
// @Summary Verify Email
// @id verify-email
// @tags Authentication
// @description Consumes a verification token to confirm that an email does belong to a user. Verification tokens cannot
// @description be retrieved from the API, they are generated when a user signs up; and a link including the token is
// @description sent to their email address.
// @Produce json
// @Accept json
// @Param Token body swag.VerifyRequest true "Verify Token"
// @Router /authentication/verify [post]
// @Success 200 {object} swag.VerifyResponse
// @Failure 400 {object} ApiError Required data is missing. The token is invalid or expired. Or the email has already been verified.
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) verifyEndpoint(ctx iris.Context) {
	var verifyRequest struct {
		Token string `json:"token"`
	}
	if err := ctx.ReadJSON(&verifyRequest); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed JSON")
		return
	}

	if strings.TrimSpace(verifyRequest.Token) == "" {
		c.badRequest(ctx, "token cannot be blank")
		return
	}

	if err := c.emailVerification.UseEmailVerificationToken(c.getContext(ctx), verifyRequest.Token); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "invalid email verification")
		return
	}

	ctx.JSON(map[string]interface{}{
		"nextUrl": "/login",
		"message": "Your email is now verified. Please login.",
	})
}

// Resend Verification Email
// @Summary Resend Verification Email
// @id resend-verification-email
// @tags Authentication
// @description This endpoint is used to generate a new verification token and email it to an address to verify that the
// @description email address is owned by a user. This endpoint only works with addresses that are associated with a
// @description login, and will return a successful status code **if** the provided email is associated with a login and
// @description the email is not already verified. All other situations will return a bad request, even if the email is
// @description valid or if the email is not associated. This is to prevent someone from being able to have relatively
// @description easy access to an endpoint that would let them see what email addresses are associated with active users.
// @Produce json
// @Accept json
// @Param Token body swag.ResendVerificationRequest true "Resend Verification Request"
// @Router /authentication/verify/resend [post]
// @Success 200
// @Failure 400 {object} ApiError Cannot resend verification link.
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) resendVerification(ctx iris.Context) {
	var request swag.ResendVerificationRequest
	if err := ctx.ReadJSON(&request); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed JSON")
		return
	}

	request.Email = strings.TrimSpace(strings.ToLower(request.Email))

	if request.Email == "" {
		c.badRequest(ctx, "email must be provided to resend verification link")
		return
	}

	if c.configuration.ReCAPTCHA.Enabled {
		if request.Captcha == nil {
			c.badRequest(ctx, "must provide ReCAPTCHA")
			return
		}

		if err := c.validateCaptchaMaybe(c.getContext(ctx), *request.Captcha); err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "invalid ReCAPTCHA provided")
			return
		}
	}

	// Make sure that no matter what, after this point we always return successful.
	ctx.StatusCode(http.StatusOK)

	login, verificationToken, err := c.emailVerification.RegenerateEmailVerificationToken(c.getContext(ctx), request.Email)
	if err != nil {
		c.reportWrappedError(ctx, err, "failed to regenerate email verification token")
		return
	}

	if err = c.communication.SendVerificationEmail(c.getContext(ctx), communication.VerifyEmailParams{
		Login: *login,
		VerifyURL: fmt.Sprintf("%s/verify/email?token=%s",
			c.configuration.GetUIURL(),
			url.QueryEscape(verificationToken),
		),
	}); err != nil {
		c.reportWrappedError(ctx, err, "failed to send (re-send) verification email")
	}

	return
}

// Send Password Reset Link
// @Summary Send Password Reset Link
// @id send-password-reset-link
// @tags Authentication
// @description This endpoint should be used to send password reset links to clients who have forgotten their password.
// @Produce json
// @Accept json
// @Param Token body swag.ForgotPasswordRequest true "Forgot Password Request"
// @Router /authentication/forgot [post]
// @Success 200
// @Failure 400 {object} swag.ForgotPasswordBadRequest
// @Failure 428 {object} swag.ForgotPasswordEmailNotVerifiedError Email verification required.
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) sendForgotPassword(ctx iris.Context) {
	var sendForgotPasswordRequest struct {
		Email     string `json:"email"`
		ReCAPTCHA string `json:"captcha"`
	}
	if err := ctx.ReadJSON(&sendForgotPasswordRequest); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed JSON")
		return
	}

	// Clean up some of the input provided just in case.
	sendForgotPasswordRequest.Email = strings.TrimSpace(strings.ToLower(sendForgotPasswordRequest.Email))
	sendForgotPasswordRequest.ReCAPTCHA = strings.TrimSpace(sendForgotPasswordRequest.ReCAPTCHA)

	if sendForgotPasswordRequest.Email == "" {
		c.badRequest(ctx, "Must provide an email address.")
		return
	}

	// If we require ReCAPTCHA then make sure they provide it.
	if c.configuration.ReCAPTCHA.ShouldVerifyForgotPassword() {
		if sendForgotPasswordRequest.ReCAPTCHA == "" {
			c.badRequest(ctx, "Must provide a valid ReCAPTCHA.")
			return
		}

		if err := c.validateCaptchaMaybe(c.getContext(ctx), sendForgotPasswordRequest.ReCAPTCHA); err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "Valid ReCAPTCHA is required")
			return
		}
	}

	// We need to retrieve the login record for the email in order to verify that the login actually exists, as well as
	// make sure the login email address has been verified.
	login, err := c.mustGetUnauthenticatedRepository(ctx).GetLoginForEmail(
		c.getContext(ctx),
		sendForgotPasswordRequest.Email,
	)
	if err != nil {
		crumbs.Debug(c.getContext(ctx), "No password reset email will be sent, login for email not found", map[string]interface{}{
			"error": err.Error(),
		})
		// Don't return an error to the client, don't want them to know if it failed to send.
		ctx.StatusCode(http.StatusOK)
		return
	}

	// If the login's email is not verified then return an error to the client.
	if c.configuration.Email.ShouldVerifyEmails() && !login.GetEmailIsVerified() {
		c.returnError(ctx, http.StatusPreconditionRequired, "You must verify your email before you can send forgot password requests.")
		return
	}

	// Generate the password reset token.
	// TODO When we allow for email address to be changed, we will also need to verify that a token being used is for
	//  the same login that it was generated for. Otherwise someone could generate a token for an email, then change the
	//  email of the login -> change the email of another login to the first email; then reset the password for a
	//  different login. I don't think this is a security risk as the actor would need to have access to both logins to
	//  begin with. But it might cause some goofy issues and is definitely not desired behavior.
	passwordResetToken, err := c.passwordResetTokens.GenerateToken(
		c.getContext(ctx),
		sendForgotPasswordRequest.Email,
		c.configuration.Email.ForgotPassword.TokenLifetime,
	)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "Failed to generate a password reset token")
		return
	}

	if err = c.communication.SendPasswordResetEmail(c.getContext(ctx), communication.ForgotPasswordParams{
		Login: *login,
		ResetURL: fmt.Sprintf("%s/password/reset?token=%s",
			c.configuration.GetUIURL(),
			url.QueryEscape(passwordResetToken),
		),
	}); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "Failed to send password reset email")
		return
	}

	ctx.StatusCode(http.StatusOK)
	return
}

// Reset Password
// @Summary Reset Password
// @id reset-password
// @tags Authentication
// @description This endpoint handles resetting passwords for users who have forgotten theirs. It requires a `token` be
// @description provided that comes from the email the `/authentication/forgot` endpoint sends to the login's email
// @description address.
// @Produce json
// @Accept json
// @Param Token body swag.ResetPasswordRequest true "Reset Password Request"
// @Router /authentication/reset [post]
// @Success 200
// @Failure 400 {object} swag.ResetPasswordBadRequest
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) resetPassword(ctx iris.Context) {
	var resetPasswordRequest struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := ctx.ReadJSON(&resetPasswordRequest); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed JSON")
		return
	}

	resetPasswordRequest.Token = strings.TrimSpace(resetPasswordRequest.Token)
	resetPasswordRequest.Password = strings.TrimSpace(resetPasswordRequest.Password)

	// The token is what verifies that the user is who they say they are even without a password. The token is emailed
	// to their verified email address.
	if resetPasswordRequest.Token == "" {
		c.badRequest(ctx, "Token must be provided to reset password.")
		return
	}

	if len(resetPasswordRequest.Password) < 8 {
		c.badRequest(ctx, "Password must be at least 8 characters long.")
		return
	}

	validation, err := c.passwordResetTokens.ValidateTokenEx(c.getContext(ctx), resetPasswordRequest.Token)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "Failed to validate password reset token")
		return
	}

	unauthenticatedRepo := c.mustGetUnauthenticatedRepository(ctx)

	// Retrieve the login for the email address in the token.
	login, err := unauthenticatedRepo.GetLoginForEmail(c.getContext(ctx), validation.Email)
	if err != nil {
		c.wrapPgError(ctx, err, "Failed to verify login for email address")
		return
	}

	// If the login's password has been changed since this token was issued, then this token is no longer valid. This
	// will basically make sure that a token cannot be used twice.
	if login.PasswordResetAt != nil && login.PasswordResetAt.After(validation.CreatedAt) {
		c.badRequest(ctx, "Password has already been reset, you must request another password reset link.")
		return
	}

	hashedPassword := hash.HashPassword(login.Email, resetPasswordRequest.Password)

	if err = unauthenticatedRepo.ResetPassword(c.getContext(ctx), login.LoginId, hashedPassword); err != nil {
		c.wrapPgError(ctx, err, "Failed to reset password")
		return
	}

	ctx.StatusCode(http.StatusOK)
}

func (c *Controller) validateRegistration(email, password, firstName string) error {
	if email == "" {
		return errors.Errorf("email cannot be blank")
	}

	if len(password) < 8 {
		return errors.Errorf("password must be at least 8 characters")
	}

	if firstName == "" {
		return errors.Errorf("first name cannot be blank")
	}

	return nil
}

func (c *Controller) validateLoginCaptcha(ctx context.Context, captcha string) error {
	if !c.configuration.ReCAPTCHA.ShouldVerifyLogin() {
		// If it is disabled then we don't need to do anything.
		return nil
	}

	return c.validateCaptchaMaybe(ctx, captcha)
}

func (c *Controller) validateRegistrationCaptcha(ctx context.Context, captcha string) error {
	if !c.configuration.ReCAPTCHA.ShouldVerifyRegistration() {
		// If it is disabled then we don't need to do anything.
		return nil
	}

	return c.validateCaptchaMaybe(ctx, captcha)
}

func (c *Controller) validateCaptchaMaybe(ctx context.Context, captcha string) error {
	if captcha == "" {
		return errors.Errorf("captcha is not valid")
	}

	span := sentry.StartSpan(ctx, "ReCAPTCHA")
	defer span.Finish()

	return c.captcha.VerifyCaptcha(ctx, captcha)
}

func (c *Controller) validateLogin(email, password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	address, err := mail.ParseAddress(email)
	if err != nil {
		return errors.New("email address provided is not valid")
	}

	if strings.ToLower(address.Address) != strings.ToLower(email) {
		return errors.New("email address provided is not valid")
	}

	return nil
}

func (c *Controller) generateToken(loginId, userId, accountId uint64) (string, error) {
	now := time.Now()
	claims := &MonetrClaims{
		LoginId:   loginId,
		UserId:    userId,
		AccountId: accountId,
		StandardClaims: jwt.StandardClaims{
			Audience: []string{
				c.configuration.APIDomainName,
			},
			ExpiresAt: now.Add(31 * 24 * time.Hour).Unix(),
			Id:        "",
			IssuedAt:  now.Unix(),
			Issuer:    c.configuration.APIDomainName,
			NotBefore: now.Unix(),
			Subject:   "monetr",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(c.configuration.JWT.LoginJwtSecret))
	if err != nil {
		return "", errors.Wrap(err, "failed to sign JWT")
	}

	return signedToken, nil
}
