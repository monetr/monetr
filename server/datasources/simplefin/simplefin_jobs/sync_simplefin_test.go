package simplefin_jobs_test

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/jarcoal/httpmock"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/datasources/simplefin"
	"github.com/monetr/monetr/server/datasources/simplefin/simplefin_jobs"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mock_http_helper"
	"github.com/monetr/monetr/server/internal/mock_simplefin"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/mockqueue"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/secrets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSyncSimpleFin(t *testing.T) {
	t.Run("simplefin not enabled", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		log := testutils.GetLog(t)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Log().Return(log).MinTimes(1)
		context.EXPECT().Configuration().Return(config.Configuration{
			SimpleFin: config.SimpleFin{
				Enabled: false,
			},
		}).MinTimes(1)

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		err := simplefin_jobs.SyncSimpleFin(
			mockqueue.NewMockContext(context),
			simplefin_jobs.SyncSimpleFinArguments{
				AccountId: "acct_bogus",
				LinkId:    "link_bogus",
			},
		)
		assert.NoError(t, err, "should return without error when simplefin is disabled")
		assert.Zero(t, httpmock.GetTotalCallCount(), "must not make any simplefin API calls when disabled")
	})

	t.Run("happy path", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		kms := secrets.NewPlaintextKMS()

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		// Use our own responder here so we can see what start date was sent.
		var startDate string
		mock_http_helper.NewHttpMockJsonResponder(
			t,
			"GET", mock_simplefin.Path(t, "/simplefin/accounts"),
			func(t *testing.T, request *http.Request) (any, int) {
				startDate = request.URL.Query().Get("start-date")
				return simplefin.AccountSet{
					Errlist: []simplefin.Error{},
					Connections: []simplefin.Connection{
						{
							ConnectionId: bankAccount.SimpleFinBankAccount.ConnectionId,
							Name:         "Lehman Brothers",
						},
					},
					Accounts: []simplefin.Account{
						{
							Id:               bankAccount.SimpleFinBankAccount.SimpleFinId,
							Name:             "Primary Checking - 1234",
							ConnectionId:     bankAccount.SimpleFinBankAccount.ConnectionId,
							Currency:         "USD",
							Balance:          "1234.56",
							AvailableBalance: new(json.Number("1200.00")),
							BalanceDate:      clock.Now().Unix(),
							Transactions:     []simplefin.Transaction{},
						},
					},
				}, http.StatusOK
			},
			nil,
		)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().Configuration().Return(config.Configuration{
			SimpleFin: config.SimpleFin{
				Enabled: true,
				AllowedHosts: []string{
					config.DefaultSimpleFinHost,
				},
			},
		}).AnyTimes()
		context.EXPECT().KMS().Return(kms).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Publisher().Return(publisher).AnyTimes()
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

		err := simplefin_jobs.SyncSimpleFin(
			mockqueue.NewMockContext(context),
			simplefin_jobs.SyncSimpleFinArguments{
				AccountId: bankAccount.AccountId,
				LinkId:    bankAccount.LinkId,
			},
		)
		assert.NoError(t, err, "must sync simplefin successfully")
		assert.Equal(
			t,
			strconv.FormatInt(clock.Now().AddDate(0, 0, -90).Unix(), 10),
			startDate,
			"should ask for 90 days of transactions",
		)
		assert.EqualValues(t, 1, httpmock.GetTotalCallCount(), "should only make one request to simplefin")

		updatedBankAccount := testutils.MustDBRead(t, bankAccount)
		assert.EqualValues(t, 123456, updatedBankAccount.CurrentBalance, "current balance should be updated")
		assert.EqualValues(t, 120000, updatedBankAccount.AvailableBalance, "available balance should be updated")

		simpleFinBankAccount := testutils.MustDBRead(t, *bankAccount.SimpleFinBankAccount)
		assert.EqualValues(t, 123456, simpleFinBankAccount.CurrentBalance, "simplefin current balance should be updated")
		assert.EqualValues(t, 120000, simpleFinBankAccount.AvailableBalance, "simplefin available balance should be updated")

		simpleFinLink := testutils.MustDBRead(t, *link.SimpleFinLink)
		assert.Equal(t, models.SimpleFinLinkStatusActive, simpleFinLink.Status, "link should still be active")
		require.NotNil(t, simpleFinLink.LastSuccessfulUpdate, "must record the successful update")
		require.NotNil(t, simpleFinLink.LastAttemptedUpdate, "must record the attempted update")
		assert.WithinDuration(t, clock.Now(), *simpleFinLink.LastSuccessfulUpdate, 0, "last successful update should be now")
		assert.WithinDuration(t, clock.Now(), *simpleFinLink.LastAttemptedUpdate, 0, "last attempted update should be now")
		assert.Empty(t, simpleFinLink.LastErrors, "should not have any errors")
	})

	t.Run("available balance falls back", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		kms := secrets.NewPlaintextKMS()

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		// SimpleFIN doesn't have to send an available balance, when it doesn't
		// we should just use the balance for both.
		mock_simplefin.MockFetchAccounts(t, simplefin.AccountSet{
			Errlist: []simplefin.Error{},
			Connections: []simplefin.Connection{
				{
					ConnectionId: bankAccount.SimpleFinBankAccount.ConnectionId,
					Name:         "Lehman Brothers",
				},
			},
			Accounts: []simplefin.Account{
				{
					Id:               bankAccount.SimpleFinBankAccount.SimpleFinId,
					Name:             "Primary Checking - 1234",
					ConnectionId:     bankAccount.SimpleFinBankAccount.ConnectionId,
					Currency:         "USD",
					Balance:          "1234.56",
					AvailableBalance: nil,
					BalanceDate:      clock.Now().Unix(),
					Transactions:     []simplefin.Transaction{},
				},
			},
		})

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().Configuration().Return(config.Configuration{
			SimpleFin: config.SimpleFin{
				Enabled: true,
				AllowedHosts: []string{
					config.DefaultSimpleFinHost,
				},
			},
		}).AnyTimes()
		context.EXPECT().KMS().Return(kms).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Publisher().Return(publisher).AnyTimes()
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

		err := simplefin_jobs.SyncSimpleFin(
			mockqueue.NewMockContext(context),
			simplefin_jobs.SyncSimpleFinArguments{
				AccountId: bankAccount.AccountId,
				LinkId:    bankAccount.LinkId,
			},
		)
		assert.NoError(t, err, "must sync simplefin successfully")

		updatedBankAccount := testutils.MustDBRead(t, bankAccount)
		assert.EqualValues(t, 123456, updatedBankAccount.CurrentBalance, "current balance should be updated")
		assert.EqualValues(t, 123456, updatedBankAccount.AvailableBalance, "available balance should use the balance")

		simpleFinBankAccount := testutils.MustDBRead(t, *bankAccount.SimpleFinBankAccount)
		assert.EqualValues(t, 123456, simpleFinBankAccount.CurrentBalance, "simplefin current balance should be updated")
		assert.EqualValues(t, 123456, simpleFinBankAccount.AvailableBalance, "simplefin available balance should use the balance")
	})

	t.Run("missing account is skipped", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		kms := secrets.NewPlaintextKMS()

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		firstBankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)
		secondBankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		// Only the first account comes back, like when SimpleFIN can't reach the
		// institution for the second one.
		mock_simplefin.MockFetchAccounts(t, simplefin.AccountSet{
			Errlist: []simplefin.Error{},
			Connections: []simplefin.Connection{
				{
					ConnectionId: firstBankAccount.SimpleFinBankAccount.ConnectionId,
					Name:         "Lehman Brothers",
				},
			},
			Accounts: []simplefin.Account{
				{
					Id:               firstBankAccount.SimpleFinBankAccount.SimpleFinId,
					Name:             "Primary Checking - 1234",
					ConnectionId:     firstBankAccount.SimpleFinBankAccount.ConnectionId,
					Currency:         "USD",
					Balance:          "1234.56",
					AvailableBalance: new(json.Number("1200.00")),
					BalanceDate:      clock.Now().Unix(),
					Transactions:     []simplefin.Transaction{},
				},
			},
		})

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().Configuration().Return(config.Configuration{
			SimpleFin: config.SimpleFin{
				Enabled: true,
				AllowedHosts: []string{
					config.DefaultSimpleFinHost,
				},
			},
		}).AnyTimes()
		context.EXPECT().KMS().Return(kms).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Publisher().Return(publisher).AnyTimes()
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

		err := simplefin_jobs.SyncSimpleFin(
			mockqueue.NewMockContext(context),
			simplefin_jobs.SyncSimpleFinArguments{
				AccountId: firstBankAccount.AccountId,
				LinkId:    firstBankAccount.LinkId,
			},
		)
		assert.NoError(t, err, "must sync simplefin successfully")

		first := testutils.MustDBRead(t, firstBankAccount)
		assert.EqualValues(t, 123456, first.CurrentBalance, "first current balance should be updated")
		assert.EqualValues(t, 120000, first.AvailableBalance, "first available balance should be updated")

		second := testutils.MustDBRead(t, secondBankAccount)
		assert.Equal(t, secondBankAccount.CurrentBalance, second.CurrentBalance, "second current balance should not change")
		assert.Equal(t, secondBankAccount.AvailableBalance, second.AvailableBalance, "second available balance should not change")

		simpleFinLink := testutils.MustDBRead(t, *link.SimpleFinLink)
		require.NotNil(t, simpleFinLink.LastSuccessfulUpdate, "must still record the successful update")
	})

	t.Run("inactive bank account", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		kms := secrets.NewPlaintextKMS()

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		// Turn the account off, it should not be touched by the sync even though
		// SimpleFIN still sends it.
		bankAccount.SimpleFinBankAccount.Status = models.SimpleFinBankAccountStatusInactive
		testutils.MustDBUpdate(t, bankAccount.SimpleFinBankAccount)

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		mock_simplefin.MockFetchAccounts(t, simplefin.AccountSet{
			Errlist: []simplefin.Error{},
			Connections: []simplefin.Connection{
				{
					ConnectionId: bankAccount.SimpleFinBankAccount.ConnectionId,
					Name:         "Lehman Brothers",
				},
			},
			Accounts: []simplefin.Account{
				{
					Id:               bankAccount.SimpleFinBankAccount.SimpleFinId,
					Name:             "Primary Checking - 1234",
					ConnectionId:     bankAccount.SimpleFinBankAccount.ConnectionId,
					Currency:         "USD",
					Balance:          "1234.56",
					AvailableBalance: new(json.Number("1200.00")),
					BalanceDate:      clock.Now().Unix(),
					Transactions:     []simplefin.Transaction{},
				},
			},
		})

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().Configuration().Return(config.Configuration{
			SimpleFin: config.SimpleFin{
				Enabled: true,
				AllowedHosts: []string{
					config.DefaultSimpleFinHost,
				},
			},
		}).AnyTimes()
		context.EXPECT().KMS().Return(kms).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Publisher().Return(publisher).AnyTimes()
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

		err := simplefin_jobs.SyncSimpleFin(
			mockqueue.NewMockContext(context),
			simplefin_jobs.SyncSimpleFinArguments{
				AccountId: bankAccount.AccountId,
				LinkId:    bankAccount.LinkId,
			},
		)
		assert.NoError(t, err, "must sync simplefin successfully")

		updatedBankAccount := testutils.MustDBRead(t, bankAccount)
		assert.Equal(t, bankAccount.CurrentBalance, updatedBankAccount.CurrentBalance, "current balance should not change")
		assert.Equal(t, bankAccount.AvailableBalance, updatedBankAccount.AvailableBalance, "available balance should not change")

		simpleFinBankAccount := testutils.MustDBRead(t, *bankAccount.SimpleFinBankAccount)
		assert.Equal(t, bankAccount.SimpleFinBankAccount.CurrentBalance, simpleFinBankAccount.CurrentBalance, "simplefin current balance should not change")
		assert.Equal(t, bankAccount.SimpleFinBankAccount.AvailableBalance, simpleFinBankAccount.AvailableBalance, "simplefin available balance should not change")
	})

	t.Run("errlist stored and cleared", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		kms := secrets.NewPlaintextKMS()

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		// Do each sync in an anonymous function so the http mocks reset properly
		// between them.
		func() {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			mock_simplefin.MockFetchAccounts(t, simplefin.AccountSet{
				Errlist: []simplefin.Error{
					{
						Code:         "con.auth",
						Message:      "Lehman Brothers needs your attention",
						ConnectionId: new(bankAccount.SimpleFinBankAccount.ConnectionId),
						AccountId:    nil,
					},
				},
				Connections: []simplefin.Connection{
					{
						ConnectionId: bankAccount.SimpleFinBankAccount.ConnectionId,
						Name:         "Lehman Brothers",
					},
				},
				Accounts: []simplefin.Account{
					{
						Id:               bankAccount.SimpleFinBankAccount.SimpleFinId,
						Name:             "Primary Checking - 1234",
						ConnectionId:     bankAccount.SimpleFinBankAccount.ConnectionId,
						Currency:         "USD",
						Balance:          "1234.56",
						AvailableBalance: new(json.Number("1200.00")),
						BalanceDate:      clock.Now().Unix(),
						Transactions:     []simplefin.Transaction{},
					},
				},
			})

			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().Configuration().Return(config.Configuration{
				SimpleFin: config.SimpleFin{
					Enabled: true,
					AllowedHosts: []string{
						config.DefaultSimpleFinHost,
					},
				},
			}).AnyTimes()
			context.EXPECT().KMS().Return(kms).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().Publisher().Return(publisher).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := simplefin_jobs.SyncSimpleFin(
				mockqueue.NewMockContext(context),
				simplefin_jobs.SyncSimpleFinArguments{
					AccountId: bankAccount.AccountId,
					LinkId:    bankAccount.LinkId,
				},
			)
			assert.NoError(t, err, "must sync simplefin successfully")

			simpleFinLink := testutils.MustDBRead(t, *link.SimpleFinLink)
			assert.Equal(t, []models.SimpleFinLinkError{
				{
					Code:         "con.auth",
					Message:      "Lehman Brothers needs your attention",
					ConnectionId: new(bankAccount.SimpleFinBankAccount.ConnectionId),
					SimpleFinId:  nil,
				},
			}, simpleFinLink.LastErrors, "should store the errlist on the link")
		}()

		func() {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			clock.Add(time.Hour)

			mock_simplefin.MockFetchAccounts(t, simplefin.AccountSet{
				Errlist: []simplefin.Error{},
				Connections: []simplefin.Connection{
					{
						ConnectionId: bankAccount.SimpleFinBankAccount.ConnectionId,
						Name:         "Lehman Brothers",
					},
				},
				Accounts: []simplefin.Account{
					{
						Id:               bankAccount.SimpleFinBankAccount.SimpleFinId,
						Name:             "Primary Checking - 1234",
						ConnectionId:     bankAccount.SimpleFinBankAccount.ConnectionId,
						Currency:         "USD",
						Balance:          "1234.56",
						AvailableBalance: new(json.Number("1200.00")),
						BalanceDate:      clock.Now().Unix(),
						Transactions:     []simplefin.Transaction{},
					},
				},
			})

			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().Configuration().Return(config.Configuration{
				SimpleFin: config.SimpleFin{
					Enabled: true,
					AllowedHosts: []string{
						config.DefaultSimpleFinHost,
					},
				},
			}).AnyTimes()
			context.EXPECT().KMS().Return(kms).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().Publisher().Return(publisher).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := simplefin_jobs.SyncSimpleFin(
				mockqueue.NewMockContext(context),
				simplefin_jobs.SyncSimpleFinArguments{
					AccountId: bankAccount.AccountId,
					LinkId:    bankAccount.LinkId,
				},
			)
			assert.NoError(t, err, "must sync simplefin successfully")

			simpleFinLink := testutils.MustDBRead(t, *link.SimpleFinLink)
			assert.Empty(t, simpleFinLink.LastErrors, "old errors should be cleared")
		}()
	})

	t.Run("payment required", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		kms := secrets.NewPlaintextKMS()

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		mock_simplefin.MockFetchAccountsError(t, http.StatusPaymentRequired)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().Configuration().Return(config.Configuration{
			SimpleFin: config.SimpleFin{
				Enabled: true,
				AllowedHosts: []string{
					config.DefaultSimpleFinHost,
				},
			},
		}).AnyTimes()
		context.EXPECT().KMS().Return(kms).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Publisher().Return(publisher).AnyTimes()
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(0)

		err := simplefin_jobs.SyncSimpleFin(
			mockqueue.NewMockContext(context),
			simplefin_jobs.SyncSimpleFinArguments{
				AccountId: bankAccount.AccountId,
				LinkId:    bankAccount.LinkId,
			},
		)
		assert.NoError(t, err, "should not return an error so the job is not retried")
		assert.EqualValues(t, 1, httpmock.GetTotalCallCount(), "should only make one request to simplefin")

		simpleFinLink := testutils.MustDBRead(t, *link.SimpleFinLink)
		assert.Equal(t, models.SimpleFinLinkStatusError, simpleFinLink.Status, "link should be in the error status")
		assert.Equal(t, []models.SimpleFinLinkError{
			{
				Code:         "payment_required",
				Message:      "SimpleFIN requires payment, check your SimpleFIN account",
				ConnectionId: nil,
				SimpleFinId:  nil,
			},
		}, simpleFinLink.LastErrors, "should store why the link is in the error status")
		assert.Nil(t, simpleFinLink.LastSuccessfulUpdate, "must not record a successful update")
		require.NotNil(t, simpleFinLink.LastAttemptedUpdate, "must still record the attempted update")
	})

	t.Run("forbidden", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		kms := secrets.NewPlaintextKMS()

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		mock_simplefin.MockFetchAccountsError(t, http.StatusForbidden)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().Configuration().Return(config.Configuration{
			SimpleFin: config.SimpleFin{
				Enabled: true,
				AllowedHosts: []string{
					config.DefaultSimpleFinHost,
				},
			},
		}).AnyTimes()
		context.EXPECT().KMS().Return(kms).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Publisher().Return(publisher).AnyTimes()
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(0)

		err := simplefin_jobs.SyncSimpleFin(
			mockqueue.NewMockContext(context),
			simplefin_jobs.SyncSimpleFinArguments{
				AccountId: bankAccount.AccountId,
				LinkId:    bankAccount.LinkId,
			},
		)
		assert.NoError(t, err, "should not return an error so the job is not retried")
		assert.EqualValues(t, 1, httpmock.GetTotalCallCount(), "should only make one request to simplefin")

		simpleFinLink := testutils.MustDBRead(t, *link.SimpleFinLink)
		assert.Equal(t, models.SimpleFinLinkStatusError, simpleFinLink.Status, "link should be in the error status")
		assert.Equal(t, []models.SimpleFinLinkError{
			{
				Code:         "forbidden",
				Message:      "SimpleFIN access was denied, the token may have been revoked",
				ConnectionId: nil,
				SimpleFinId:  nil,
			},
		}, simpleFinLink.LastErrors, "should store why the link is in the error status")
		assert.Nil(t, simpleFinLink.LastSuccessfulUpdate, "must not record a successful update")
		require.NotNil(t, simpleFinLink.LastAttemptedUpdate, "must still record the attempted update")
	})

	t.Run("upstream error", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		kms := secrets.NewPlaintextKMS()

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		mock_simplefin.MockFetchAccountsError(t, http.StatusInternalServerError)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().Configuration().Return(config.Configuration{
			SimpleFin: config.SimpleFin{
				Enabled: true,
				AllowedHosts: []string{
					config.DefaultSimpleFinHost,
				},
			},
		}).AnyTimes()
		context.EXPECT().KMS().Return(kms).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Publisher().Return(publisher).AnyTimes()
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(0)

		// FailWithoutRetry panics and the queue catches it, that's how the job
		// gets marked as failed without being retried.
		assert.PanicsWithError(
			t,
			"failed to retrieve accounts from SimpleFIN: SimpleFIN request failed https://bridge.simplefin.org/simplefin/accounts?start-date=-7776000&version=2 [500]: \"Internal Server Error\"\n",
			func() {
				_ = simplefin_jobs.SyncSimpleFin(
					mockqueue.NewMockContext(context),
					simplefin_jobs.SyncSimpleFinArguments{
						AccountId: bankAccount.AccountId,
						LinkId:    bankAccount.LinkId,
					},
				)
			},
			"should fail without retrying",
		)
		assert.EqualValues(t, 1, httpmock.GetTotalCallCount(), "should only make one request to simplefin")

		simpleFinLink := testutils.MustDBRead(t, *link.SimpleFinLink)
		assert.Equal(t, models.SimpleFinLinkStatusActive, simpleFinLink.Status, "link status should not change")
		assert.Nil(t, simpleFinLink.LastSuccessfulUpdate, "must not record a successful update")
		require.NotNil(t, simpleFinLink.LastAttemptedUpdate, "must still record the attempted update")
	})

	t.Run("error link goes active", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		kms := secrets.NewPlaintextKMS()

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		// Put the link in the error status like a 403 would, then the user fixes
		// their token and the next sync works.
		link.SimpleFinLink.Status = models.SimpleFinLinkStatusError
		link.SimpleFinLink.LastErrors = []models.SimpleFinLinkError{
			{
				Code:         "forbidden",
				Message:      "SimpleFIN access was denied, the token may have been revoked",
				ConnectionId: nil,
				SimpleFinId:  nil,
			},
		}
		testutils.MustDBUpdate(t, link.SimpleFinLink)

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		mock_simplefin.MockFetchAccounts(t, simplefin.AccountSet{
			Errlist: []simplefin.Error{},
			Connections: []simplefin.Connection{
				{
					ConnectionId: bankAccount.SimpleFinBankAccount.ConnectionId,
					Name:         "Lehman Brothers",
				},
			},
			Accounts: []simplefin.Account{
				{
					Id:               bankAccount.SimpleFinBankAccount.SimpleFinId,
					Name:             "Primary Checking - 1234",
					ConnectionId:     bankAccount.SimpleFinBankAccount.ConnectionId,
					Currency:         "USD",
					Balance:          "1234.56",
					AvailableBalance: new(json.Number("1200.00")),
					BalanceDate:      clock.Now().Unix(),
					Transactions:     []simplefin.Transaction{},
				},
			},
		})

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().Configuration().Return(config.Configuration{
			SimpleFin: config.SimpleFin{
				Enabled: true,
				AllowedHosts: []string{
					config.DefaultSimpleFinHost,
				},
			},
		}).AnyTimes()
		context.EXPECT().KMS().Return(kms).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Publisher().Return(publisher).AnyTimes()
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

		err := simplefin_jobs.SyncSimpleFin(
			mockqueue.NewMockContext(context),
			simplefin_jobs.SyncSimpleFinArguments{
				AccountId: bankAccount.AccountId,
				LinkId:    bankAccount.LinkId,
			},
		)
		assert.NoError(t, err, "must sync simplefin successfully")

		simpleFinLink := testutils.MustDBRead(t, *link.SimpleFinLink)
		assert.Equal(t, models.SimpleFinLinkStatusActive, simpleFinLink.Status, "link should be active again")
		assert.Empty(t, simpleFinLink.LastErrors, "old errors should be cleared")
		require.NotNil(t, simpleFinLink.LastSuccessfulUpdate, "must record the successful update")
	})

	t.Run("pending link", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		kms := secrets.NewPlaintextKMS()

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		link.SimpleFinLink.Status = models.SimpleFinLinkStatusPending
		testutils.MustDBUpdate(t, link.SimpleFinLink)

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().Configuration().Return(config.Configuration{
			SimpleFin: config.SimpleFin{
				Enabled: true,
				AllowedHosts: []string{
					config.DefaultSimpleFinHost,
				},
			},
		}).AnyTimes()
		context.EXPECT().KMS().Return(kms).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Publisher().Return(publisher).AnyTimes()
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(0)

		err := simplefin_jobs.SyncSimpleFin(
			mockqueue.NewMockContext(context),
			simplefin_jobs.SyncSimpleFinArguments{
				AccountId: bankAccount.AccountId,
				LinkId:    bankAccount.LinkId,
			},
		)
		assert.NoError(t, err, "should skip the pending link without an error")
		assert.Zero(t, httpmock.GetTotalCallCount(), "must not make any simplefin API calls for a pending link")

		simpleFinLink := testutils.MustDBRead(t, *link.SimpleFinLink)
		assert.Equal(t, models.SimpleFinLinkStatusPending, simpleFinLink.Status, "link should still be pending")
		assert.Nil(t, simpleFinLink.LastSuccessfulUpdate, "must not record a successful update")
	})

	t.Run("progress messages", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		kms := secrets.NewPlaintextKMS()

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		listener, err := publisher.Subscribe(
			t.Context(),
			bankAccount.AccountId,
			simplefin_jobs.SimpleFinSyncNotificationChannel(bankAccount.AccountId, bankAccount.LinkId),
		)
		require.NoError(t, err, "must be able to subscribe to sync progress")
		defer listener.Close()

		// The listener drops messages if nothing is reading from it, so we need to
		// be reading while the sync runs.
		payloads := make(chan string, 10)
		go func() {
			for notification := range listener.Channel() {
				payloads <- notification.Payload()
			}
		}()

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		mock_simplefin.MockFetchAccounts(t, simplefin.AccountSet{
			Errlist: []simplefin.Error{},
			Connections: []simplefin.Connection{
				{
					ConnectionId: bankAccount.SimpleFinBankAccount.ConnectionId,
					Name:         "Lehman Brothers",
				},
			},
			Accounts: []simplefin.Account{
				{
					Id:               bankAccount.SimpleFinBankAccount.SimpleFinId,
					Name:             "Primary Checking - 1234",
					ConnectionId:     bankAccount.SimpleFinBankAccount.ConnectionId,
					Currency:         "USD",
					Balance:          "1234.56",
					AvailableBalance: new(json.Number("1200.00")),
					BalanceDate:      clock.Now().Unix(),
					Transactions:     []simplefin.Transaction{},
				},
			},
		})

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().Configuration().Return(config.Configuration{
			SimpleFin: config.SimpleFin{
				Enabled: true,
				AllowedHosts: []string{
					config.DefaultSimpleFinHost,
				},
			},
		}).AnyTimes()
		context.EXPECT().KMS().Return(kms).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Publisher().Return(publisher).AnyTimes()
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

		err = simplefin_jobs.SyncSimpleFin(
			mockqueue.NewMockContext(context),
			simplefin_jobs.SyncSimpleFinArguments{
				AccountId: bankAccount.AccountId,
				LinkId:    bankAccount.LinkId,
			},
		)
		assert.NoError(t, err, "must sync simplefin successfully")

		statuses := make([]string, 0)
		deadline := time.After(10 * time.Second)
	ReadLoop:
		for len(statuses) < 3 {
			select {
			case payload := <-payloads:
				var message map[string]string
				require.NoError(t, json.Unmarshal([]byte(payload), &message), "must be able to decode progress message")
				assert.Equal(t, bankAccount.LinkId.String(), message["linkId"], "progress message should have the link Id")
				statuses = append(statuses, message["status"])
			case <-deadline:
				break ReadLoop
			}
		}

		assert.Equal(t, []string{
			"begin",
			"balances",
			"complete",
		}, statuses, "should announce the sync progress in order")
	})

	t.Run("progress messages when denied", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)
		kms := secrets.NewPlaintextKMS()

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		listener, err := publisher.Subscribe(
			t.Context(),
			bankAccount.AccountId,
			simplefin_jobs.SimpleFinSyncNotificationChannel(bankAccount.AccountId, bankAccount.LinkId),
		)
		require.NoError(t, err, "must be able to subscribe to sync progress")
		defer listener.Close()

		// The listener drops messages if nothing is reading from it, so we need to
		// be reading while the sync runs.
		payloads := make(chan string, 10)
		go func() {
			for notification := range listener.Channel() {
				payloads <- notification.Payload()
			}
		}()

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		mock_simplefin.MockFetchAccountsError(t, http.StatusForbidden)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().Configuration().Return(config.Configuration{
			SimpleFin: config.SimpleFin{
				Enabled: true,
				AllowedHosts: []string{
					config.DefaultSimpleFinHost,
				},
			},
		}).AnyTimes()
		context.EXPECT().KMS().Return(kms).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Publisher().Return(publisher).AnyTimes()
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(0)

		err = simplefin_jobs.SyncSimpleFin(
			mockqueue.NewMockContext(context),
			simplefin_jobs.SyncSimpleFinArguments{
				AccountId: bankAccount.AccountId,
				LinkId:    bankAccount.LinkId,
			},
		)
		assert.NoError(t, err, "should not return an error so the job is not retried")

		statuses := make([]string, 0)
		deadline := time.After(10 * time.Second)
	ReadLoop:
		for len(statuses) < 2 {
			select {
			case payload := <-payloads:
				var message map[string]string
				require.NoError(t, json.Unmarshal([]byte(payload), &message), "must be able to decode progress message")
				statuses = append(statuses, message["status"])
			case <-deadline:
				break ReadLoop
			}
		}

		assert.Equal(t, []string{
			"begin",
			"error",
		}, statuses, "should announce the error so the client stops waiting")
	})
}
