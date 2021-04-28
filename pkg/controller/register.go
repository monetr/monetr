package controller

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/kataras/iris/v12"
	"github.com/monetrapp/rest-api/pkg/hash"
	"github.com/monetrapp/rest-api/pkg/models"
	"github.com/pkg/errors"
	"github.com/stripe/stripe-go/v72"
	"net/http"
	"strings"
	"time"
)

type RegistrationClaims struct {
	RegistrationId string `json:"registrationId"`
	jwt.StandardClaims
}

func (c *Controller) registerEndpoint(ctx iris.Context) {
	var registerRequest struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Timezone  string `json:"timezone"`
		Captcha   string `json:"captcha"`
	}
	if err := ctx.ReadJSON(&registerRequest); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "invalid register JSON")
		return
	}

	// This will take the captcha from the request and validate it if the API is
	// configured to do so. If it is enabled and the captcha fails then an error
	// is returned to the client.
	if err := c.validateCaptchaMaybe(c.getContext(ctx), registerRequest.Captcha); err != nil {
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

	timezone, err := time.LoadLocation(registerRequest.Timezone)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "failed to parse timezone")
		return
	}

	// TODO (elliotcourant) Add stuff to verify email address by sending an
	//  email.

	var stripeCustomerId *string
	if c.configuration.Stripe.Enabled {
		c.log.Debug("creating stripe customer for new user")
		name := registerRequest.FirstName + " " + registerRequest.LastName
		result, err := c.stripeClient.Customers.New(&stripe.CustomerParams{
			Email: &registerRequest.Email,
			Name:  &name,
		})
		if err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create stripe customer")
			return
		}

		stripeCustomerId = &result.ID
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

	// Hash the user's password so that we can store it securely.
	hashedPassword := hash.HashPassword(
		registerRequest.Email, registerRequest.Password,
	)

	// Create the user's login record in the database, this will return the login
	// record including the new login's loginId which we will need below. If SMTP
	// is enabled and we want to verify emails then the user is disabled
	// initially.
	login, err := repository.CreateLogin(
		registerRequest.Email,
		hashedPassword,
		registerRequest.FirstName,
		registerRequest.LastName,
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
	account, err := repository.CreateAccount(timezone)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError,
			"failed to create account",
		)
		return
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

	// If SMTP is enabled and we are verifying emails then we want to create a
	// registration record and send the user a verification email.
	if c.configuration.SMTP.Enabled && c.configuration.SMTP.VerifyEmails {
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
		"nextUrl": "/setup",
		"token":   token,
		"user":    user,
	})
}

func (c *Controller) verifyEndpoint(ctx iris.Context) {
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

func (c *Controller) validateCaptchaMaybe(ctx context.Context, captcha string) error {
	if !c.configuration.ReCAPTCHA.Enabled {
		// If it is disabled then we don't need to do anything.
		return nil
	}

	if captcha == "" {
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
