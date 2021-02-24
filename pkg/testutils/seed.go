package testutils

import (
	"context"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-pg/pg/v10"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/hash"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/plaid/plaid-go/plaid"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

type SeedAccountOption uint8

const (
	Nothing           SeedAccountOption = 0
	WithManualAccount SeedAccountOption = 1
	WithPlaidAccount  SeedAccountOption = 2
)

func GivenIHaveAnEmail(t *testing.T) string {
	return fmt.Sprintf("%s@testing.harderthanitneedstobe.com", strings.ReplaceAll(gofakeit.UUID(), "-", ""))
}

func SeedAccount(t *testing.T, db *pg.DB, options SeedAccountOption) (*models.User, *MockPlaidData) {
	require.NotNil(t, db, "db must not be nil")

	plaidData := &MockPlaidData{
		BankAccounts: map[string]map[string]plaid.Account{},
	}

	var user models.User
	err := db.RunInTransaction(context.Background(), func(txn *pg.Tx) error {
		email := GivenIHaveAnEmail(t)
		login := models.Login{
			Email:              email,
			PasswordHash:       hash.HashPassword(email, gofakeit.Password(true, true, true, true, false, 16)),
			IsEnabled:          true,
			IsEmailVerified:    true,
			Users:              nil,
			EmailVerifications: nil,
			PhoneVerifications: nil,
		}

		_, err := txn.Model(&login).Insert(&login)
		require.NoError(t, err, "must insert new login")

		account := models.Account{
			Timezone: "UTC",
		}

		_, err = txn.Model(&account).Insert(&account)
		require.NoError(t, err, "failed to insert new account")

		user = models.User{
			LoginId:   login.LoginId,
			AccountId: account.AccountId,
			FirstName: gofakeit.FirstName(),
			LastName:  gofakeit.LastName(),
		}

		_, err = txn.Model(&user).Insert(&user)
		require.NoError(t, err, "failed to insert new user")

		user.Login = &login
		user.Account = &account

		now := time.Now().UTC()
		if options&WithManualAccount > 0 {
			manualLink := models.Link{
				AccountId:       account.AccountId,
				LinkType:        models.ManualLinkType,
				InstitutionName: gofakeit.Company() + " Bank",
				CreatedAt:       now,
				CreatedByUserId: user.UserId,
				UpdatedAt:       now,
				UpdatedByUserId: &user.UserId,
			}

			_, err := txn.Model(&manualLink).Insert(&manualLink)
			require.NoError(t, err, "failed to create manual link")

			checkingBalance := int64(gofakeit.Float32Range(0.00, 1000.00) * 100)
			savingBalance := int64(gofakeit.Float32Range(0.00, 1000.00) * 100)
			bankAccounts := []models.BankAccount{
				{
					AccountId:         account.AccountId,
					LinkId:            manualLink.LinkId,
					AvailableBalance:  checkingBalance,
					CurrentBalance:    checkingBalance,
					Mask:              "1234",
					Name:              "Checking Account",
					PlaidName:         "Checking Account",
					PlaidOfficialName: "Checking",
					Type:              "depository",
					SubType:           "checking",
				},
				{
					AccountId:         account.AccountId,
					LinkId:            manualLink.LinkId,
					AvailableBalance:  savingBalance,
					CurrentBalance:    savingBalance,
					Mask:              "2345",
					Name:              "Savings Account",
					PlaidName:         "Savings Account",
					PlaidOfficialName: "Savings",
					Type:              "depository",
					SubType:           "saving",
				},
			}

			_, err = txn.Model(&bankAccounts).Insert(&bankAccounts)
			require.NoError(t, err, "failed to create bank accounts")
		}

		if options&WithPlaidAccount > 0 {
			plaidLink := models.PlaidLink{
				ItemId:          gofakeit.UUID(),
				AccessToken:     gofakeit.UUID(),
				Products:        []string{"transactions"},
				WebhookUrl:      "",
				InstitutionId:   "123",
				InstitutionName: "A Bank",
			}

			_, err = txn.Model(&plaidLink).Insert(&plaidLink)
			require.NoError(t, err, "failed to create plaid link")

			withPlaidLink := models.Link{
				AccountId:       account.AccountId,
				LinkType:        models.PlaidLinkType,
				PlaidLinkId:     &plaidLink.PlaidLinkID,
				InstitutionName: gofakeit.Company() + " Bank",
				CreatedAt:       now,
				CreatedByUserId: user.UserId,
				UpdatedAt:       now,
				UpdatedByUserId: &user.UserId,
			}

			_, err = txn.Model(&withPlaidLink).Insert(&withPlaidLink)
			require.NoError(t, err, "failed to create link")

			checkingBalance := int64(gofakeit.Float32Range(0.00, 1000.00) * 100)
			savingBalance := int64(gofakeit.Float32Range(0.00, 1000.00) * 100)
			checkingAccountId := gofakeit.UUID()
			savingsAccountId := gofakeit.UUID()
			bankAccounts := []models.BankAccount{
				{
					AccountId:         account.AccountId,
					LinkId:            withPlaidLink.LinkId,
					PlaidAccountId:    checkingAccountId,
					AvailableBalance:  checkingBalance,
					CurrentBalance:    checkingBalance,
					Mask:              "1234",
					Name:              "Checking Account",
					PlaidName:         "Checking Account",
					PlaidOfficialName: "Checking",
					Type:              "depository",
					SubType:           "checking",
				},
				{
					AccountId:         account.AccountId,
					LinkId:            withPlaidLink.LinkId,
					PlaidAccountId:    savingsAccountId,
					AvailableBalance:  savingBalance,
					CurrentBalance:    savingBalance,
					Mask:              "2345",
					Name:              "Savings Account",
					PlaidName:         "Savings Account",
					PlaidOfficialName: "Savings",
					Type:              "depository",
					SubType:           "saving",
				},
			}

			plaidData.BankAccounts[plaidLink.AccessToken] = map[string]plaid.Account{
				checkingAccountId: {
					AccountID: checkingAccountId,
					Balances: plaid.AccountBalances{
						Available:              float64(checkingBalance) / 100,
						Current:                float64(checkingBalance) / 100,
						Limit:                  0,
						ISOCurrencyCode:        "USD",
						UnofficialCurrencyCode: "",
					},
					Mask:               "1234",
					Name:               "Checking Account",
					OfficialName:       "Checking",
					Subtype:            "depository",
					Type:               "checking",
					VerificationStatus: "",
				},
				savingsAccountId: {
					AccountID: savingsAccountId,
					Balances: plaid.AccountBalances{
						Available:              float64(savingBalance) / 100,
						Current:                float64(savingBalance) / 100,
						Limit:                  0,
						ISOCurrencyCode:        "USD",
						UnofficialCurrencyCode: "",
					},
					Mask:               "2345",
					Name:               "Savings Account",
					OfficialName:       "Savings",
					Subtype:            "depository",
					Type:               "saving",
					VerificationStatus: "",
				},
			}

			_, err = txn.Model(&bankAccounts).Insert(&bankAccounts)
			require.NoError(t, err, "failed to create bank accounts")
		}

		return nil
	})
	require.NoError(t, err, "should seed account")

	return &user, plaidData
}
