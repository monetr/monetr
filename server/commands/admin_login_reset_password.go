package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/benbjohnson/clock"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/communication"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/database"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/security"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func adminLoginResetPassword(parent *cobra.Command) {
	var arguments struct {
		LoginID string
		Email   string
		Silent  bool
	}
	command := &cobra.Command{
		Use:   "login:reset-password",
		Short: "Generate a password reset link for a login (and email it when SMTP is configured).",
		Long: strings.Join([]string{
			"Mints a single-use password reset link for the target login using the",
			"same PASETO machinery as the web forgot-password flow, and prints the",
			"URL to stdout. When SMTP is configured the reset email is also sent to",
			"the user; otherwise the operator is expected to relay the printed link",
			"directly. The link's lifetime matches the configured",
			"Email.ForgotPassword.TokenLifetime (default 10 minutes), and any",
			"earlier link for the same login is invalidated as soon as the user",
			"completes a reset (single-use).",
		}, " "),
		RunE: func(cmd *cobra.Command, _ []string) error {
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

			// The link's reset token has to be signed with the same Ed25519 key the
			// running monetr server uses, otherwise the server will reject it on
			// consumption. The loadCertificates helper would otherwise silently fall
			// back to a fresh in-memory key (useful for first boot serve, useless
			// here), so require a real key path before we go any further.
			if configuration.Security.PrivateKey == "" {
				return errors.New("Security.PrivateKey must be configured: the running server's signing key is needed to mint a token it will accept")
			}
			if _, err := os.Stat(configuration.Security.PrivateKey); err != nil {
				return errors.Wrap(err, "configured signing key is not readable; tokens minted now will not validate against the running server")
			}

			publicKey, privateKey, err := loadCertificates(configuration, log, false)
			if err != nil {
				log.Error("failed to load ed25519 keypair", "err", err)
				return err
			}

			clientTokens, err := security.NewPasetoClientTokens(
				log,
				clock.New(),
				configuration.Server.GetBaseURL().String(),
				publicKey,
				privateKey,
			)
			if err != nil {
				log.Error("failed to init paseto client tokens interface", "err", err)
				return err
			}

			tokenLifetime := configuration.Email.ForgotPassword.TokenLifetime
			token, err := clientTokens.Create(
				tokenLifetime,
				security.Claims{
					Scope:        security.ResetPasswordScope,
					EmailAddress: login.Email,
					UserId:       "",
					AccountId:    "",
					LoginId:      login.LoginId.String(),
				},
			)
			if err != nil {
				log.Error("failed to mint password reset token", "err", err)
				return errors.Wrap(err, "failed to mint password reset token")
			}

			resetURL := configuration.Server.GetURL("/password/reset", map[string]string{
				"token": token,
			})

			fmt.Println()
			fmt.Println("Login:     ", login.LoginId, fmt.Sprintf("(%s)", login.Email))
			fmt.Println("Expires in:", tokenLifetime)
			fmt.Println()
			fmt.Println("PASSWORD RESET LINK:")
			fmt.Println()
			fmt.Println(resetURL)
			fmt.Println()

			// Mirror the web flow's gate so we don't email out a reset on a
			// deployment where forgot-password is intentionally disabled. The link
			// itself is always printed so the operator can hand-deliver it.
			if arguments.Silent {
				fmt.Println("Email not sent to user because --silent was specified!")
			} else if configuration.Email.AllowPasswordReset() {
				email := communication.NewEmailCommunication(log, configuration)
				if err := email.SendEmail(ctx, communication.PasswordResetParams{
					BaseURL:      configuration.Server.GetBaseURL().String(),
					Email:        login.Email,
					FirstName:    login.FirstName,
					LastName:     login.LastName,
					SupportEmail: "support@monetr.app",
					ResetURL:     resetURL,
				}); err != nil {
					log.Warn("failed to send password reset email; the printed link is still valid", "err", err)
					fmt.Println("WARNING: failed to send the reset email; relay the link above out-of-band.")
				} else {
					fmt.Println("A reset email has also been sent to the user.")
				}
			} else {
				fmt.Println("Email is not configured on this deployment; relay the link above to user directly.")
			}

			return nil
		},
	}

	command.PersistentFlags().StringVar(&arguments.LoginID, "login-id", "", "The login ID (e.g. lgn_...) of the login to reset.")
	command.PersistentFlags().StringVar(&arguments.Email, "email", "", "The email address of the login to reset.")
	command.PersistentFlags().BoolVar(&arguments.Silent, "silent", false, "Silent determines whether or not a password reset email is sent (when SMTP is configured). Defaults to: false")
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
