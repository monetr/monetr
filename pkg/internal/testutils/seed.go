package testutils

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/pkg/hash"
	"github.com/monetr/monetr/pkg/internal/myownsanity"
	"github.com/monetr/monetr/pkg/models"
	"github.com/plaid/plaid-go/plaid"
	"github.com/stretchr/testify/require"
)

type SeedAccountOption uint8

const (
	Nothing           SeedAccountOption = 0
	WithManualAccount SeedAccountOption = 1
	WithPlaidAccount  SeedAccountOption = 2
)

func SeedAccount(t *testing.T, db *pg.DB, options SeedAccountOption) (*models.User, *MockPlaidData) {
	require.NotNil(t, db, "db must not be nil")

	plaidData := &MockPlaidData{
		PlaidTokens:  map[string]models.PlaidToken{},
		PlaidLinks:   map[string]models.PlaidLink{},
		BankAccounts: map[string]map[string]plaid.AccountBase{},
	}

	var user models.User
	err := db.RunInTransaction(context.Background(), func(txn *pg.Tx) error {
		email := GetUniqueEmail(t)
		login := models.LoginWithHash{
			Login: models.Login{
				Email:           email,
				IsEnabled:       true,
				IsEmailVerified: true,
				FirstName:       gofakeit.FirstName(),
				LastName:        gofakeit.LastName(),
				Users:           nil,
			},
			PasswordHash: hash.HashPassword(email, gofakeit.Password(true, true, true, true, false, 16)),
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
			FirstName: login.FirstName,
			LastName:  login.LastName,
		}

		_, err = txn.Model(&user).Insert(&user)
		require.NoError(t, err, "failed to insert new user")

		user.Login = &login.Login
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
				Products:        []string{"transactions"},
				WebhookUrl:      "",
				InstitutionId:   "123",
				InstitutionName: "A Bank",
			}

			_, err = txn.Model(&plaidLink).Insert(&plaidLink)
			require.NoError(t, err, "failed to create plaid link")

			plaidData.PlaidLinks[plaidLink.ItemId] = plaidLink

			accessToken := gofakeit.UUID()
			plaidData.PlaidTokens[accessToken] = models.PlaidToken{
				ItemId:      plaidLink.ItemId,
				AccountId:   account.AccountId,
				AccessToken: accessToken,
			}

			withPlaidLink := models.Link{
				AccountId:       account.AccountId,
				LinkType:        models.PlaidLinkType,
				LinkStatus:      models.LinkStatusSetup,
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

			checkingAccountSubType := plaid.ACCOUNTSUBTYPE_CHECKING
			savingAccountSubType := plaid.ACCOUNTSUBTYPE_SAVINGS

			plaidData.BankAccounts[accessToken] = map[string]plaid.AccountBase{
				checkingAccountId: {
					AccountId: checkingAccountId,
					Balances: plaid.AccountBalance{
						Available:              *plaid.NewNullableFloat32(myownsanity.Float32P(float32(checkingBalance) / 100)),
						Current:                *plaid.NewNullableFloat32(myownsanity.Float32P(float32(checkingBalance) / 100)),
						IsoCurrencyCode:        *plaid.NewNullableString(myownsanity.StringP("USD")),
						UnofficialCurrencyCode: *plaid.NewNullableString(myownsanity.StringP("USD")),
					},
					Mask:         *plaid.NewNullableString(myownsanity.StringP("1234")),
					Name:         "Checking Account",
					OfficialName: *plaid.NewNullableString(myownsanity.StringP("Checking")),
					Type:         plaid.ACCOUNTTYPE_DEPOSITORY,
					Subtype:      *plaid.NewNullableAccountSubtype(&checkingAccountSubType),
				},
				savingsAccountId: {
					AccountId: savingsAccountId,
					Balances: plaid.AccountBalance{
						Available:              *plaid.NewNullableFloat32(myownsanity.Float32P(float32(savingBalance) / 100)),
						Current:                *plaid.NewNullableFloat32(myownsanity.Float32P(float32(savingBalance) / 100)),
						IsoCurrencyCode:        *plaid.NewNullableString(myownsanity.StringP("USD")),
						UnofficialCurrencyCode: *plaid.NewNullableString(myownsanity.StringP("USD")),
					},
					Mask:         *plaid.NewNullableString(myownsanity.StringP("2345")),
					Name:         "Savings Account",
					OfficialName: *plaid.NewNullableString(myownsanity.StringP("Savings")),
					Type:         plaid.ACCOUNTTYPE_DEPOSITORY,
					Subtype:      *plaid.NewNullableAccountSubtype(&savingAccountSubType),
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
