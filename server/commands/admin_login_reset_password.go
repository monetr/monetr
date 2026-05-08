package commands

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/benbjohnson/clock"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/database"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func adminLoginResetPassword(parent *cobra.Command) {
	var arguments struct {
		LoginID string
		Email   string
	}
	command := &cobra.Command{
		Use:   "login:reset-password",
		Short: "Reset a login's password to a randomly-generated value.",
		Long: strings.Join([]string{
			"Generates a fresh 16-character password for the target login and prints",
			"it to stdout so the operator can communicate it to the user directly.",
			"TOTP and active sessions are not affected; the operator is responsible",
			"for any follow-up that monetr would normally do via the email-based",
			"reset flow.",
		}, " "),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			configuration := config.LoadConfiguration()
			log := logging.NewLoggerWithConfig(configuration.Logging)
			if configFileName := configuration.GetConfigFileName(); configFileName != "" {
				log.Info("config file loaded", "config", configFileName)
			}
			db, err := database.GetDatabase(log, configuration, nil)
			if err != nil {
				log.Error("failed to establish database connection", "err", err)
				return err
			}
			defer db.Close()

			repo := repository.NewUnauthenticatedRepository(clock.New(), db)

			login, err := resolveLoginForReset(ctx, db, repo, arguments.LoginID, arguments.Email)
			if err != nil {
				return err
			}

			if !login.IsEnabled {
				fmt.Println("WARNING: this login is currently disabled.")
			}
			if login.TOTPEnabledAt != nil {
				fmt.Println("WARNING: TOTP is still enabled; the user will need their authenticator to sign in.")
			}
			if !login.IsEmailVerified {
				fmt.Println("WARNING: this login's email is not verified; sign-in may be blocked depending on configuration.")
			}

			password, err := generateAdminPassword()
			if err != nil {
				log.Error("failed to generate password", "err", err)
				return err
			}

			if err := repo.ResetPassword(ctx, login.LoginId, password); err != nil {
				log.Error("failed to reset login password", "err", err)
				return errors.Wrap(err, "failed to reset login password")
			}

			fmt.Println()
			fmt.Println("Login:   ", login.LoginId, fmt.Sprintf("(%s)", login.Email))
			fmt.Println()
			fmt.Println("NEW PASSWORD:")
			fmt.Println()
			fmt.Println(password)
			fmt.Println()
			fmt.Println("Communicate this password to the user out-of-band.")

			return nil
		},
	}

	command.PersistentFlags().StringVar(&arguments.LoginID, "login-id", "", "The login ID (e.g. lgn_...) of the login to reset.")
	command.PersistentFlags().StringVar(&arguments.Email, "email", "", "The email address of the login to reset.")
	command.MarkFlagsMutuallyExclusive("login-id", "email")
	command.MarkFlagsOneRequired("login-id", "email")

	parent.AddCommand(command)
}

func resolveLoginForReset(
	ctx context.Context,
	db pg.DBI,
	repo repository.UnauthenticatedRepository,
	loginIDArg, emailArg string,
) (*models.Login, error) {
	if emailArg != "" {
		login, err := repo.GetLoginForEmail(ctx, emailArg)
		if err != nil {
			if errors.Is(err, pg.ErrNoRows) {
				return nil, errors.Errorf("no login found for email %q", emailArg)
			}
			return nil, errors.Wrap(err, "failed to find login by email")
		}
		return login, nil
	}

	loginId, err := models.ParseID[models.Login](loginIDArg)
	if err != nil {
		return nil, errors.Wrap(err, "invalid login id")
	}

	var login models.Login
	err = db.ModelContext(ctx, &login).
		Where(`"login"."login_id" = ?`, loginId).
		Limit(1).
		Select(&login)
	if err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			return nil, errors.Errorf("no login found for login id %q", loginIDArg)
		}
		return nil, errors.Wrap(err, "failed to find login by id")
	}
	return &login, nil
}

// generateAdminPassword returns a 16-character password drawn from a shell-safe
// pool of letters, digits and symbols, with at least one character from each
// class. The four "anchor" characters are placed at the front and then shuffled
// so the position of each class is not predictable.
func generateAdminPassword() (string, error) {
	const (
		lower  = "abcdefghijklmnopqrstuvwxyz"
		upper  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		digits = "0123456789"
		// Intentionally omits quote characters, backtick, backslash, dollar,
		// semicolon, pipe and whitespace so the password can be pasted into a shell
		// or chat client without escaping.
		special = "!@#%^&*()-_=+[]{}<>?"
	)
	classes := []string{lower, upper, digits, special}
	pool := strings.Join(classes, "")
	out := make([]byte, 16)

	for i, class := range classes {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(class))))
		if err != nil {
			return "", errors.Wrap(err, "failed to read random data")
		}
		out[i] = class[n.Int64()]
	}
	for i := len(classes); i < len(out); i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(pool))))
		if err != nil {
			return "", errors.Wrap(err, "failed to read random data")
		}
		out[i] = pool[n.Int64()]
	}
	for i := len(out) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", errors.Wrap(err, "failed to shuffle password")
		}
		out[i], out[j.Int64()] = out[j.Int64()], out[i]
	}
	return string(out), nil
}
