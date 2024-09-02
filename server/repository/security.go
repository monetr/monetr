package repository

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"math/big"
	"strings"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/consts"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"github.com/xlzd/gotp"
	"golang.org/x/crypto/bcrypt"
)

var (
	// Deprecated: use requiresPasswordChange instead ErrRequirePasswordChange is
	// an error that is returned when a user's password must be updated. This is
	// more or less future planning, as I'd like to implement a new method of
	// storing passwords and need to make sure a path to make user's update their
	// passwords is in place. This error should not be considered a "failure"
	// state.
	ErrRequirePasswordChange = errors.New("password must be updated")
	// ErrInvalidCredentials is returned when the provided email and/or password
	// does not match the current password for an existing login.
	ErrInvalidCredentials = errors.New("invalid credentials provided")
)

type SecurityRepository interface {
	// Login recieves the user's provided email address as well as password. This
	// password is then hashed using bcrypt and is then compared to the stored
	// hash for that login. If the provided password is equivalent to the stored
	// hash then this function will return the login model for the credentials
	// provided. As well as a boolean indicating whether or not the user must
	// change their password at this time. A user MUST NOT be given credentials if
	// their password requires changing. If the provided credentials are invalid
	// then ErrInvalidCredentials is returned. NOTE: ErrRequirePasswordChange is
	// no longer returned.
	Login(ctx context.Context, email, password string) (_ *Login, requiresPasswordChange bool, err error)

	// ChangePassword accepts a login ID and the old hashed password and the new
	// hashed password. The two passwords should be hashed from the user's input.
	// Specifically, you should not retrieve the "oldHashedPassword" from the
	// database and then use it as input for this method. This way the function
	// will only succeed if the provided input is 100% valid. This method will
	// return true if the oldHashedPassword is correct and the update succeeds, it
	// will return false if the oldHashedPassword is incorrect and/or if the
	// update fail. If the oldHashedPassword provided is not valid for the login
	// ID, then ErrInvalidCredentials will be returned.
	ChangePassword(ctx context.Context, loginId ID[Login], oldHashedPassword, newHashedPassword string) error

	// SetupTOTP takes a login ID and begins the process of enabling TOTP for that
	// login. If the login already has TOTP enabled then an error will be
	// returned. Otherwise the TOTP URI and several recovery codes will be
	// returned. Calling this function does not require TOTP upon login, the TOTP
	// must be enabled by having the user confirm their code on the frontend.
	SetupTOTP(ctx context.Context, loginId ID[Login]) (secret string, recoveryCodes []string, err error)
	// EnableTOTP takes a login ID and a TOTP code. If the login does not yet have
	// TOTP enabled but is currently in the setup process then this will validate
	// the code provided against the current TOTP settings and if they are valid
	// will enable TOTP for the specified login. If they are not valid then this
	// function will return an error.
	EnableTOTP(ctx context.Context, loginId ID[Login], code string) error
}

var (
	_ SecurityRepository = &baseSecurityRepository{}
)

type baseSecurityRepository struct {
	db    pg.DBI
	clock clock.Clock
}

func NewSecurityRepository(db pg.DBI, clock clock.Clock) SecurityRepository {
	return &baseSecurityRepository{
		db:    db,
		clock: clock,
	}
}

func (b *baseSecurityRepository) Login(ctx context.Context, email, password string) (*Login, bool, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	requiresPasswordChange := false

	var login LoginWithHash
	err := b.db.ModelContext(span.Context(), &login).
		Relation("Users").
		Relation("Users.Account").
		Where(`"login_with_hash"."email" = ?`, strings.ToLower(email)).
		Limit(1).
		Select(&login)
	switch err {
	case nil:
	case pg.ErrNoRows:
		span.Status = sentry.SpanStatusNotFound
		return nil, requiresPasswordChange, errors.WithStack(ErrInvalidCredentials)
	default:
		span.Status = sentry.SpanStatusInternalError
		return nil, requiresPasswordChange, crumbs.WrapError(span.Context(), err, "failed to verify credentials")
	}

	if err = bcrypt.CompareHashAndPassword(login.Crypt, []byte(password)); err != nil {
		return nil, requiresPasswordChange, errors.WithStack(ErrInvalidCredentials)
	}

	span.Status = sentry.SpanStatusOK
	return &login.Login, requiresPasswordChange, nil
}

func (b *baseSecurityRepository) ChangePassword(ctx context.Context, loginId ID[Login], oldPassword, newPassword string) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var login LoginWithHash
	err := b.db.ModelContext(span.Context(), &login).
		Where(`"login_id" = ?`, loginId).
		Limit(1).
		Select(&login)
	if err != nil {
		return crumbs.WrapError(span.Context(), err, "failed to find login record to change password")
	}

	if err = bcrypt.CompareHashAndPassword(login.Crypt, []byte(oldPassword)); err != nil {
		return errors.WithStack(ErrInvalidCredentials)
	}

	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), consts.BcryptCost)
	if err != nil {
		return crumbs.WrapError(span.Context(), err, "failed to encrypt new password for change")
	}

	_, err = b.db.ModelContext(span.Context(), &login).
		Set(`"crypt" = ?`, newPasswordHash).
		Where(`"login_id" = ?`, loginId).
		Update(&login)
	if err != nil {
		return crumbs.WrapError(span.Context(), err, "failed to update password")
	}

	return nil
}

func (b *baseSecurityRepository) SetupTOTP(
	ctx context.Context,
	loginId ID[Login],
) (uri string, recoveryCodes []string, err error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var login Login
	err = b.db.ModelContext(span.Context(), &login).
		Where(`"login"."login_id" = ?`, loginId).
		Select(&login)
	if err != nil {
		return "", nil, crumbs.WrapError(span.Context(), err, "failed to retrieve login details")
	}

	if login.TOTPEnabledAt != nil {
		return "", nil, errors.Errorf("login already has TOTP enabled")
	}

	randBytes := make([]byte, 64)
	if _, err := rand.Read(randBytes); err != nil {
		return "", nil, errors.Wrap(err, "failed to generate TOTP secret")
	}

	secret := base32.StdEncoding.EncodeToString(randBytes)
	login.TOTP = secret

	// Generate 10, 6 digit recovery codes
	digits := 6
	recoveryCount := 10
	login.TOTPRecoveryCodes = make([]string, recoveryCount)
	// TODO Technically it should be rare but possible for this to generate
	// duplicate recovery codes. It would be super rare, but possible.
	for i := range login.TOTPRecoveryCodes {
		for x := 0; x < digits; x++ {
			digit, err := rand.Int(rand.Reader, big.NewInt(9))
			if err != nil {
				return "", nil, errors.Wrap(err, "failed to generate recovery codes")
			}
			login.TOTPRecoveryCodes[i] += digit.String()
		}
	}
	// Make sure this is nil
	login.TOTPEnabledAt = nil

	// Store this data on the login itself, that way when the user confirms it we
	// can simply set the enabled at date.
	_, err = b.db.ModelContext(span.Context(), &login).
		Set(`"totp" = ?`, login.TOTP).
		Set(`"totp_recovery_codes" = ?`, pg.Array(login.TOTPRecoveryCodes)).
		Set(`"totp_enabled_at" = ?`, login.TOTPEnabledAt).
		WherePK().
		Update(&login)
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to save TOTP settings")
	}

	// Generate the URI to return to the caller. This is what the client needs for
	// the QR code.
	uri = gotp.BuildUri(
		"totp",      // Kind
		secret,      // Secret
		login.Email, // Account namem
		"monetr",    // Issuer
		"",          // Algorithm
		0,           // Initial count
		6,           // Digits
		30,          // Period in seconds
	)

	return uri, login.TOTPRecoveryCodes, nil
}

func (b *baseSecurityRepository) EnableTOTP(
	ctx context.Context,
	loginId ID[Login],
	code string,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	var login Login
	err := b.db.ModelContext(span.Context(), &login).
		Where(`"login"."login_id" = ?`, loginId).
		Limit(1).
		Select(&login)
	if err != nil {
		return errors.Wrap(err, "failed to retrive login to enable TOTP")
	}

	if len(login.TOTPRecoveryCodes) == 0 || login.TOTP == "" {
		return errors.New("TOTP has not been setup on this login")
	}

	if login.TOTPEnabledAt != nil {
		return errors.New("TOTP is already enabled on this login")
	}

	err = login.VerifyTOTP(code, b.clock.Now())
	if err != nil {
		return err
	}

	login.TOTPEnabledAt = myownsanity.TimeP(b.clock.Now())

	_, err = b.db.ModelContext(span.Context(), &login).
		Set(`"totp_enabled_at" = ?`, login.TOTPEnabledAt).
		WherePK().
		Update(&login)
	if err != nil {
		return errors.Wrap(err, "failed to enable TOTP")
	}

	return nil
}
