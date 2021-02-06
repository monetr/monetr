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
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "invalid registration")
		return
	}

	// TODO (elliotcourant) Add stuff to verify email address by sending an
	//  email.

	repository, err := c.getUnauthenticatedRepository(ctx)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "cannot register user")
		return
	}

	hashedPassword := c.hashPassword(registerRequest.Email, registerRequest.Password)

	login, err := repository.CreateLogin(registerRequest.Email, hashedPassword)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create login")
		return
	}

	account, err := repository.CreateAccount(time.Local)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create account")
		return
	}

	user, err := repository.CreateUser(login.LoginId, account.AccountId, registerRequest.FirstName, registerRequest.LastName)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create user")
		return
	}

	if c.configuration.SMTP.Enabled && c.configuration.SMTP.VerifyEmails {
		registration, err := repository.CreateRegistration(login.LoginId)
		if err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create registration")
			return
		}

		registrationToken, err := c.generateRegistrationToken(registration.RegistrationId)
		if err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create registration token")
			return
		}

		if err := c.sendEmailVerification(registerRequest.Email, registrationToken); err != nil {

		}
	}

	token, err := c.generateToken(login.LoginId, user.UserId, account.AccountId)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create JWT")
		return
	}

	user.Login = login
	user.Account = account

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
