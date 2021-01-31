package controller

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/go-pg/pg/v10"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/pkg/errors"

	"github.com/kataras/iris/v12/context"
)

type HarderClaims struct {
	LoginId   uint64 `json:"loginId"`
	UserId    uint64 `json:"userId"`
	AccountId uint64 `json:"accountId"`
	jwt.StandardClaims
}

func (c *Controller) loginEndpoint(ctx *context.Context) {
	var loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := ctx.ReadJSON(&loginRequest); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "failed to decode login request")
		return
	}
	loginRequest.Email = strings.TrimSpace(loginRequest.Email)
	loginRequest.Password = strings.TrimSpace(loginRequest.Password)

	if err := c.validateLogin(loginRequest.Email, loginRequest.Password); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "login is not valid")
		return
	}

	hashedPassword := c.hashPassword(loginRequest.Email, loginRequest.Password)
	var login models.Login
	if err := c.db.RunInTransaction(ctx.Request().Context(), func(txn *pg.Tx) error {
		return txn.Model(&login).
			Relation("Users").
			Relation("Users.Account").
			Where(`"login"."email" = ? AND "login"."password_hash" = ?`, loginRequest.Email, hashedPassword).
			Limit(1).
			Select(&login)
	}); err != nil {
		if err == pg.ErrNoRows {
			c.returnError(ctx, http.StatusForbidden, "invalid email and password")
			return
		}

		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to authenticate")
		return
	}

	switch len(login.Users) {
	case 0:
		// TODO (elliotcourant) Should we allow them to create an account?
		c.returnError(ctx, http.StatusForbidden, "user has no accounts")
		return
	case 1:
		user := login.Users[0]
		token, err := c.generateToken(login.LoginId, user.UserId, user.AccountId)
		if err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "could not generate JWT")
			return
		}
		// Return their account token.
		ctx.JSON(map[string]interface{}{
			"token": token,
		})
		return
	default:
		// If the login has more than one user then we want to generate a temp
		// JWT that will only grant them access to API endpoints not specific to
		// an account.
		token, err := c.generateToken(login.LoginId, 0, 0)
		if err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "could not generate JWT")
			return
		}

		ctx.JSON(map[string]interface{}{
			"token": token,
			"users": login.Users,
		})
	}
}

func (c *Controller) validateLogin(email, password string) error {
	// TODO (elliotcourant) Add some email format validation here.
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	return nil
}

func (c *Controller) hashPassword(email, password string) string {
	email = strings.ToLower(email)
	hash := sha256.New()
	hash.Write([]byte(email))
	hash.Write([]byte(password))
	return fmt.Sprintf("%X", hash.Sum(nil))
}

func (c *Controller) generateToken(loginId, userId, accountId uint64) (string, error) {
	now := time.Now()
	claims := &HarderClaims{
		LoginId:   loginId,
		UserId:    userId,
		AccountId: accountId,
		StandardClaims: jwt.StandardClaims{
			Audience:  c.configuration.APIDomainName,
			ExpiresAt: now.Add(31 * 24 * time.Hour).Unix(),
			Id:        "",
			IssuedAt:  now.Unix(),
			Issuer:    c.configuration.APIDomainName,
			NotBefore: now.Unix(),
			Subject:   "harderThanItNeedsToBe",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(c.configuration.JWTSecret))
	if err != nil {
		return "", errors.Wrap(err, "failed to sign JWT")
	}

	return signedToken, nil
}
