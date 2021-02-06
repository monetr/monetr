package controller

import (
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"strings"
	"time"

	"github.com/kataras/iris/v12/context"
	"github.com/pkg/errors"
)

type RegistrationClaims struct {
	RegistrationId string `json:"registrationId"`
	jwt.StandardClaims
}

func (c *Controller) registerEndpoint(ctx *context.Context) {
	var registerRequest struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Captcha   string `json:"captcha"`
	}
	if err := ctx.ReadJSON(&registerRequest); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "invalid register JSON")
		return
	}

	// This will take the captcha from the request and validate it if the API is
	// configured to do so. If it is enabled and the captcha fails then an error
	// is returned to the client.
	if err := c.validateCaptchaMaybe(registerRequest.Captcha); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "valid ReCAPTCHA is required")
		return
	}

	registerRequest.Email = strings.TrimSpace(registerRequest.Email)
	registerRequest.Password = strings.TrimSpace(registerRequest.Password)
	registerRequest.FirstName = strings.TrimSpace(registerRequest.FirstName)

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

	// TODO (elliotcourant) Add stuff to verify email address by sending an
	//  email.

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

	// Hash the user's password so that we can store it securely.
	hashedPassword := c.hashPassword(
		registerRequest.Email, registerRequest.Password,
	)

	// Create the user's login record in the database, this will return the login
	// record including the new login's loginId which we will need below. If SMTP
	// is enabled and we want to verify emails then the user is disabled
	// initially.
	login, err := repository.CreateLogin(
		registerRequest.Email,
		hashedPassword,
		!(c.configuration.SMTP.Enabled && c.configuration.SMTP.VerifyEmails),
	)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError,
			"failed to create login",
		)
		return
	}

	// Now that the login exists we can create the account, at the time of
	// writing this we are only using the local time zone of the server, but in
	// the future I want to have it somehow use the user's timezone.
	account, err := repository.CreateAccount(time.Local)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError,
			"failed to create account",
		)
		return
	}

	// Now that we have an accountId we can create the user object which will
	// bind the login and the account together.
	user, err := repository.CreateUser(
		login.LoginId,
		account.AccountId,
		registerRequest.FirstName,
		registerRequest.LastName,
	)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError,
			"failed to create user",
		)
		return
	}

	// If SMTP is enabled and we are verifying emails then we want to create a
	// registration record and send the user a verification email.
	if c.configuration.SMTP.Enabled && c.configuration.SMTP.VerifyEmails {
		registration, err := repository.CreateRegistration(login.LoginId)
		if err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError,
				"failed to create registration",
			)
			return
		}

		// Once we have the registrationId create a token specifically for it. This
		// token is used in a link that we send the user in an email.
		registrationToken, err := c.generateRegistrationToken(registration.RegistrationId)
		if err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError,
				"failed to create registration token",
			)
			return
		}

		// With the email and the token send the user a message asking them to
		// activate their account.
		if err := c.sendEmailVerification(
			registerRequest.Email, registrationToken,
		); err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError,
				"failed to send activation email",
			)
			return
		}

		ctx.JSON(map[string]interface{}{
			"needsVerification": true,
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
	user.Account = account

	ctx.JSON(map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

func (c *Controller) verifyEndpoint(ctx *context.Context) {
	var verifyRequest struct {
		Token string `json:"token"`
	}
	if err := ctx.ReadJSON(&verifyRequest); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed json")
		return
	}

	var claims RegistrationClaims
	result, err := jwt.ParseWithClaims(verifyRequest.Token, &claims, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(c.configuration.JWT.RegistrationJwtSecret), nil
	})
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusForbidden, "unauthorized")
		return
	}

	if !result.Valid {
		c.returnError(ctx, http.StatusForbidden, "unauthorized")
		return
	}

	repo, err := c.getUnauthenticatedRepository(ctx)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to get repository")
		return
	}

	user, err := repo.VerifyRegistration(claims.RegistrationId)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to verify registration")
		return
	}

	token, err := c.generateToken(user.LoginId, user.UserId, user.AccountId)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError,
			"failed to create JWT",
		)
		return
	}

	ctx.JSON(map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

func (c *Controller) validateRegistration(email, password, firstName string) error {
	if len(email) == 0 {
		return errors.Errorf("email cannot be blank")
	}

	if len(password) < 8 {
		return errors.Errorf("password must be at least 8 characters")
	}

	if len(firstName) == 0 {
		return errors.Errorf("first name cannot be blank")
	}

	return nil
}

func (c *Controller) validateCaptchaMaybe(captcha string) error {
	if !c.configuration.ReCAPTCHA.Enabled {
		// If it is disabled then we don't need to do anything.
		return nil
	}

	if len(captcha) == 0 {
		return errors.Errorf("captcha is not valid")
	}

	return c.captcha.Verify(captcha)
}

func (c *Controller) generateRegistrationToken(registrationId string) (string, error) {
	now := time.Now()
	claims := &RegistrationClaims{
		RegistrationId: registrationId,
		StandardClaims: jwt.StandardClaims{
			Audience:  c.configuration.APIDomainName,
			ExpiresAt: now.Add(7 * 24 * time.Hour).Unix(),
			Id:        "",
			IssuedAt:  now.Unix(),
			Issuer:    c.configuration.APIDomainName,
			NotBefore: now.Unix(),
			Subject:   "harderThanItNeedsToBe - registration",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(c.configuration.JWT.RegistrationJwtSecret))
	if err != nil {
		return "", errors.Wrap(err, "failed to sign registration JWT")
	}

	return signedToken, nil
}
