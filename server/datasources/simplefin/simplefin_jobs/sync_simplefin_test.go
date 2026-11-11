package simplefin_jobs_test

import (
	"encoding/json"
	"errors"
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
	"github.com/monetr/monetr/server/similar/similar_jobs"
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
		enqueuer := mockgen.NewMockProcessor(ctrl)
		enqueuer.EXPECT().
			BulkEnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(similar_jobs.CalculateTransactionClusters),
				gomock.Any(),
				gomock.Any(),
			).
			Return(nil).
			Times(1)

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
		context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
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
		enqueuer := mockgen.NewMockProcessor(ctrl)
		enqueuer.EXPECT().
			BulkEnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(similar_jobs.CalculateTransactionClusters),
				gomock.Any(),
				gomock.Any(),
			).
			Return(nil).
			Times(1)

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
		context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
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
		enqueuer := mockgen.NewMockProcessor(ctrl)
		enqueuer.EXPECT().
			BulkEnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(similar_jobs.CalculateTransactionClusters),
				gomock.Any(),
				gomock.Any(),
			).
			Return(nil).
			Times(1)

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
		context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
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
		enqueuer := mockgen.NewMockProcessor(ctrl)
		enqueuer.EXPECT().
			BulkEnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(similar_jobs.CalculateTransactionClusters),
				gomock.Any(),
				gomock.Any(),
			).
			Return(nil).
			Times(2)

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
			context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
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
			context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
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
			"failed to retrieve accounts from SimpleFIN: SimpleFIN request failed https://bridge.simplefin.org/simplefin/accounts?start-date="+
				strconv.FormatInt(clock.Now().AddDate(0, 0, -90).Unix(), 10)+
				"&version=2 [500]: \"Internal Server Error\"\n",
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
		enqueuer := mockgen.NewMockProcessor(ctrl)
		enqueuer.EXPECT().
			BulkEnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(similar_jobs.CalculateTransactionClusters),
				gomock.Any(),
				gomock.Any(),
			).
			Return(nil).
			Times(1)

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
		context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
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
		enqueuer := mockgen.NewMockProcessor(ctrl)
		enqueuer.EXPECT().
			BulkEnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(similar_jobs.CalculateTransactionClusters),
				gomock.Any(),
				gomock.Any(),
			).
			Return(nil).
			Times(1)

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
		context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
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
		for len(statuses) < 4 {
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
			"transactions",
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

	t.Run("last successful update only moves when data changes", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		enqueuer := givenAnyClusterEnqueue(ctrl, 4)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		sync := func(account simplefin.Account) models.SimpleFinLink {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			mock_simplefin.MockFetchAccounts(t, simpleFinAccountSet(account))
			runSimpleFinSync(t, ctrl, clock, enqueuer, bankAccount)
			return testutils.MustDBRead(t, *link.SimpleFinLink)
		}

		account := simpleFinAccountFor(bankAccount)

		// The first sync always sets it so the link doesn't look like it has never
		// synced.
		firstSyncAt := clock.Now()
		simpleFinLink := sync(account)
		require.NotNil(t, simpleFinLink.LastSuccessfulUpdate, "must record the first successful update")
		assert.WithinDuration(t, firstSyncAt, *simpleFinLink.LastSuccessfulUpdate, 0, "first sync should set the last successful update")

		// Same balances and no transactions, nothing changed.
		clock.Add(time.Hour)
		simpleFinLink = sync(account)
		require.NotNil(t, simpleFinLink.LastSuccessfulUpdate, "must still have the last successful update")
		require.NotNil(t, simpleFinLink.LastAttemptedUpdate, "must record the attempted update")
		assert.WithinDuration(t, firstSyncAt, *simpleFinLink.LastSuccessfulUpdate, 0, "should not move when nothing changed")
		assert.WithinDuration(t, clock.Now(), *simpleFinLink.LastAttemptedUpdate, 0, "last attempted update should still move")

		// Only the balance changed.
		clock.Add(time.Hour)
		account.Balance = "1300.00"
		simpleFinLink = sync(account)
		require.NotNil(t, simpleFinLink.LastSuccessfulUpdate, "must still have the last successful update")
		assert.WithinDuration(t, clock.Now(), *simpleFinLink.LastSuccessfulUpdate, 0, "should move when the balance changes")

		// Only a new transaction, the balances are the same as the last sync.
		clock.Add(time.Hour)
		account.Transactions = []simplefin.Transaction{
			{Id: "TRN-1", Posted: clock.Now().Unix(), Amount: "-19.99", Description: "TACO BELL"},
		}
		simpleFinLink = sync(account)
		require.NotNil(t, simpleFinLink.LastSuccessfulUpdate, "must still have the last successful update")
		assert.WithinDuration(t, clock.Now(), *simpleFinLink.LastSuccessfulUpdate, 0, "should move when there is a new transaction")
	})
}

func TestSyncSimpleFinCron(t *testing.T) {
	t.Run("no links to sync", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		enqueuer := mockgen.NewMockProcessor(ctrl)
		enqueuer.EXPECT().
			BulkEnqueueAt(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).
			Return(nil).
			Times(0)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().Configuration().Return(config.Configuration{
			SimpleFin: config.SimpleFin{
				Enabled: true,
				AllowedHosts: []string{
					config.DefaultSimpleFinHost,
				},
			},
		}).MinTimes(1)
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Enqueuer().Return(enqueuer).Times(0)
		context.EXPECT().Log().Return(log).AnyTimes()

		err := simplefin_jobs.SyncSimpleFinCron(
			mockqueue.NewMockContext(context),
		)
		assert.NoError(t, err, "sync simplefin cron should succeed")
	})

	t.Run("enqueues one job per link", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		firstLink := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &firstLink)
		fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &firstLink)
		secondLink := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &secondLink)

		enqueuer := mockgen.NewMockProcessor(ctrl)
		enqueuer.EXPECT().
			BulkEnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(simplefin_jobs.SyncSimpleFin),
				gomock.Any(),
				gomock.InAnyOrder([]any{
					simplefin_jobs.SyncSimpleFinArguments{
						AccountId: firstLink.AccountId,
						LinkId:    firstLink.LinkId,
					},
					simplefin_jobs.SyncSimpleFinArguments{
						AccountId: secondLink.AccountId,
						LinkId:    secondLink.LinkId,
					},
				}),
			).
			Return(nil).
			Times(1)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().Configuration().Return(config.Configuration{
			SimpleFin: config.SimpleFin{
				Enabled: true,
				AllowedHosts: []string{
					config.DefaultSimpleFinHost,
				},
			},
		}).MinTimes(1)
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()

		err := simplefin_jobs.SyncSimpleFinCron(
			mockqueue.NewMockContext(context),
		)
		assert.NoError(t, err, "sync simplefin cron should succeed")
	})

	t.Run("returns an error when the enqueue fails", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		enqueuer := mockgen.NewMockProcessor(ctrl)
		enqueuer.EXPECT().
			BulkEnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(simplefin_jobs.SyncSimpleFin),
				gomock.Any(),
				gomock.Any(),
			).
			Return(errors.New("database is on fire")).
			Times(1)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().Configuration().Return(config.Configuration{
			SimpleFin: config.SimpleFin{
				Enabled: true,
				AllowedHosts: []string{
					config.DefaultSimpleFinHost,
				},
			},
		}).MinTimes(1)
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()

		err := simplefin_jobs.SyncSimpleFinCron(
			mockqueue.NewMockContext(context),
		)
		assert.EqualError(t, err, "database is on fire")
	})
}

func TestSyncSimpleFinTransactions(t *testing.T) {
	// 2026-10-15 03:00 UTC is still 2026-10-14 in Chicago, so the dates in these
	// tests only line up if the account's timezone is actually used.
	now := time.Date(2026, 10, 15, 3, 0, 0, 0, time.UTC)

	t.Run("creates transactions", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(now)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		enqueuer := givenAnyClusterEnqueue(ctrl, 1)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		user.Account.Timezone = "America/Chicago"
		testutils.MustDBUpdate(t, user.Account)
		timezone, err := user.Account.GetTimezone()
		require.NoError(t, err, "must be able to load the account timezone")
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		transactedAt := now.Add(-26 * time.Hour).Unix()
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		mock_simplefin.MockFetchAccounts(t, simpleFinAccountSet(
			simpleFinAccountFor(bankAccount,
				simplefin.Transaction{
					Id:           "TRN-1",
					Posted:       now.Unix(),
					Amount:       "-19.99",
					Description:  "TACO BELL",
					TransactedAt: &transactedAt,
				},
				simplefin.Transaction{
					Id:          "TRN-2",
					Posted:      now.Add(-33 * time.Hour).Unix(),
					Amount:      "2000.00",
					Description: "PAYROLL",
				},
			),
		))

		runSimpleFinSync(t, ctrl, clock, enqueuer, bankAccount)

		transactions := readSimpleFinTransactions(t, bankAccount)
		require.Len(t, transactions, 2, "should create both transactions")

		spent := transactions[0]
		assert.EqualValues(t, 1999, spent.Amount, "money spent should be positive in monetr")
		assert.Equal(t, time.Date(2026, 10, 14, 0, 0, 0, 0, timezone), spent.Date.In(timezone), "date should be midnight of the posted day in the account timezone")
		assert.Equal(t, "TACO BELL", spent.Name, "name should be the description")
		assert.Equal(t, "TACO BELL", spent.OriginalName, "original name should be the description")
		assert.Equal(t, models.TransactionSourceSimpleFin, spent.Source, "source should be simplefin")
		assert.False(t, spent.IsPending, "should not be pending")
		require.NotNil(t, spent.SimpleFinTransaction, "must be linked to the simplefin transaction")
		assert.Equal(t, "TRN-1", spent.SimpleFinTransaction.SimpleFinId, "should store the simplefin id")
		assert.Equal(t, *bankAccount.SimpleFinBankAccountId, spent.SimpleFinTransaction.SimpleFinBankAccountId, "should be on the simplefin bank account")
		assert.WithinDuration(t, now, spent.SimpleFinTransaction.Posted, 0, "should store the posted timestamp")
		require.NotNil(t, spent.SimpleFinTransaction.TransactedAt, "should store transacted at when it is sent")
		assert.WithinDuration(t, time.Unix(transactedAt, 0), *spent.SimpleFinTransaction.TransactedAt, 0, "should store transacted at")

		deposit := transactions[1]
		assert.EqualValues(t, -200000, deposit.Amount, "deposits should be negative in monetr")
		assert.Equal(t, time.Date(2026, 10, 13, 0, 0, 0, 0, timezone), deposit.Date.In(timezone), "date should be midnight of the posted day in the account timezone")
		require.NotNil(t, deposit.SimpleFinTransaction, "must be linked to the simplefin transaction")
		assert.Nil(t, deposit.SimpleFinTransaction.TransactedAt, "transacted at should be empty when it isn't sent")
	})

	t.Run("same response twice creates nothing new", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(now)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		enqueuer := givenAnyClusterEnqueue(ctrl, 2)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		accountSet := simpleFinAccountSet(
			simpleFinAccountFor(bankAccount,
				simplefin.Transaction{Id: "TRN-1", Posted: now.Unix(), Amount: "-19.99", Description: "TACO BELL"},
				simplefin.Transaction{Id: "TRN-2", Posted: now.Unix(), Amount: "-5.00", Description: "COFFEE"},
			),
		)

		for range 2 {
			func() {
				httpmock.Activate()
				defer httpmock.DeactivateAndReset()
				mock_simplefin.MockFetchAccounts(t, accountSet)
				runSimpleFinSync(t, ctrl, clock, enqueuer, bankAccount)
			}()
			clock.Add(time.Hour)
		}

		assert.EqualValues(t, 2, fixtures.CountAllTransactions(t, user.AccountId), "second sync should not create anything")
	})

	t.Run("second sync creates only the new transaction", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(now)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		enqueuer := givenAnyClusterEnqueue(ctrl, 2)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		first := simplefin.Transaction{Id: "TRN-1", Posted: now.Unix(), Amount: "-19.99", Description: "TACO BELL"}
		second := simplefin.Transaction{Id: "TRN-2", Posted: now.Unix(), Amount: "-5.00", Description: "COFFEE"}

		func() {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			mock_simplefin.MockFetchAccounts(t, simpleFinAccountSet(simpleFinAccountFor(bankAccount, first)))
			runSimpleFinSync(t, ctrl, clock, enqueuer, bankAccount)
		}()
		original := readSimpleFinTransactions(t, bankAccount)
		require.Len(t, original, 1, "first sync should create one transaction")

		clock.Add(time.Hour)
		func() {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			mock_simplefin.MockFetchAccounts(t, simpleFinAccountSet(simpleFinAccountFor(bankAccount, second, first)))
			runSimpleFinSync(t, ctrl, clock, enqueuer, bankAccount)
		}()

		transactions := readSimpleFinTransactions(t, bankAccount)
		require.Len(t, transactions, 2, "second sync should only add the new transaction")
		assert.Equal(t, original[0].TransactionId, transactions[0].TransactionId, "existing transaction should be left alone")
		assert.Equal(t, "TRN-2", transactions[1].SimpleFinTransaction.SimpleFinId, "new transaction should be created")
	})

	t.Run("two connections land on the right accounts", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(now)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		enqueuer := givenAnyClusterEnqueue(ctrl, 1)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		// The fixture gives each bank account its own connection.
		firstBankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)
		secondBankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)
		require.NotEqual(t, firstBankAccount.SimpleFinBankAccount.ConnectionId, secondBankAccount.SimpleFinBankAccount.ConnectionId, "bank accounts should be on different connections")

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		mock_simplefin.MockFetchAccounts(t, simpleFinAccountSet(
			simpleFinAccountFor(firstBankAccount,
				simplefin.Transaction{Id: "TRN-A1", Posted: now.Unix(), Amount: "-1.00", Description: "FIRST ONE"},
				simplefin.Transaction{Id: "TRN-A2", Posted: now.Unix(), Amount: "-2.00", Description: "FIRST TWO"},
			),
			simpleFinAccountFor(secondBankAccount,
				simplefin.Transaction{Id: "TRN-B1", Posted: now.Unix(), Amount: "-3.00", Description: "SECOND ONE"},
			),
		))

		runSimpleFinSync(t, ctrl, clock, enqueuer, firstBankAccount)

		first := readSimpleFinTransactions(t, firstBankAccount)
		require.Len(t, first, 2, "first bank account should have its two transactions")
		assert.Equal(t, "TRN-A1", first[0].SimpleFinTransaction.SimpleFinId)
		assert.Equal(t, "TRN-A2", first[1].SimpleFinTransaction.SimpleFinId)

		second := readSimpleFinTransactions(t, secondBankAccount)
		require.Len(t, second, 1, "second bank account should have its one transaction")
		assert.Equal(t, "TRN-B1", second[0].SimpleFinTransaction.SimpleFinId)
		assert.Equal(t, *secondBankAccount.SimpleFinBankAccountId, second[0].SimpleFinTransaction.SimpleFinBankAccountId)
	})

	t.Run("same account id on different connections", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(now)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		enqueuer := givenAnyClusterEnqueue(ctrl, 1)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		firstBankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)
		secondBankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)
		// SimpleFIN only promises an account id is unique within its connection.
		secondBankAccount.SimpleFinBankAccount.SimpleFinId = firstBankAccount.SimpleFinBankAccount.SimpleFinId
		testutils.MustDBUpdate(t, secondBankAccount.SimpleFinBankAccount)

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		// The transaction ids are the same too since those are only unique within
		// an account.
		mock_simplefin.MockFetchAccounts(t, simpleFinAccountSet(
			simpleFinAccountFor(firstBankAccount,
				simplefin.Transaction{Id: "TRN-1", Posted: now.Unix(), Amount: "-1.00", Description: "FIRST"},
			),
			simpleFinAccountFor(secondBankAccount,
				simplefin.Transaction{Id: "TRN-1", Posted: now.Unix(), Amount: "-2.00", Description: "SECOND"},
			),
		))

		runSimpleFinSync(t, ctrl, clock, enqueuer, firstBankAccount)

		first := readSimpleFinTransactions(t, firstBankAccount)
		require.Len(t, first, 1, "first bank account should have one transaction")
		assert.EqualValues(t, 100, first[0].Amount, "first bank account should get its own transaction")
		assert.Equal(t, "FIRST", first[0].Name)

		second := readSimpleFinTransactions(t, secondBankAccount)
		require.Len(t, second, 1, "second bank account should have one transaction")
		assert.EqualValues(t, 200, second[0].Amount, "second bank account should get its own transaction")
		assert.Equal(t, "SECOND", second[0].Name)
	})

	t.Run("deleted transaction does not come back", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(now)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		enqueuer := givenAnyClusterEnqueue(ctrl, 2)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		accountSet := simpleFinAccountSet(
			simpleFinAccountFor(bankAccount,
				simplefin.Transaction{Id: "TRN-1", Posted: now.Unix(), Amount: "-19.99", Description: "TACO BELL"},
			),
		)

		func() {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			mock_simplefin.MockFetchAccounts(t, accountSet)
			runSimpleFinSync(t, ctrl, clock, enqueuer, bankAccount)
		}()

		transactions := readSimpleFinTransactions(t, bankAccount)
		require.Len(t, transactions, 1, "first sync should create the transaction")
		transaction := transactions[0]
		transaction.DeletedAt = new(clock.Now())
		testutils.MustDBUpdate(t, &transaction)
		require.EqualValues(t, 0, fixtures.CountNonDeletedTransactions(t, user.AccountId), "transaction should be deleted")

		clock.Add(time.Hour)
		func() {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			mock_simplefin.MockFetchAccounts(t, accountSet)
			runSimpleFinSync(t, ctrl, clock, enqueuer, bankAccount)
		}()

		assert.EqualValues(t, 0, fixtures.CountNonDeletedTransactions(t, user.AccountId), "deleted transaction should not come back")
		assert.EqualValues(t, 1, fixtures.CountAllTransactions(t, user.AccountId), "should not create a new copy of the transaction")
	})

	t.Run("empty description uses the fallback name", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(now)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		enqueuer := givenAnyClusterEnqueue(ctrl, 1)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		user.Account.Timezone = "America/Chicago"
		testutils.MustDBUpdate(t, user.Account)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		mock_simplefin.MockFetchAccounts(t, simpleFinAccountSet(
			simpleFinAccountFor(bankAccount,
				simplefin.Transaction{Id: "TRN-1", Posted: now.Unix(), Amount: "-12.34", Description: ""},
			),
		))

		runSimpleFinSync(t, ctrl, clock, enqueuer, bankAccount)

		transactions := readSimpleFinTransactions(t, bankAccount)
		require.Len(t, transactions, 1, "should create the transaction")
		// The date is the local date, not the UTC one.
		assert.Equal(t, "Transaction 2026-10-14 -12.34 USD", transactions[0].Name, "should use the fallback name")
		assert.Equal(t, "Transaction 2026-10-14 -12.34 USD", transactions[0].OriginalName, "should use the fallback original name")
		assert.Equal(t, "Transaction 2026-10-14 -12.34 USD", transactions[0].SimpleFinTransaction.Description, "should use the fallback description")
	})

	t.Run("JPY has no decimals", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(now)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		enqueuer := givenAnyClusterEnqueue(ctrl, 1)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)
		bankAccount.Currency = "JPY"
		testutils.MustDBUpdate(t, &bankAccount)
		bankAccount.SimpleFinBankAccount.Currency = "JPY"
		testutils.MustDBUpdate(t, bankAccount.SimpleFinBankAccount)

		account := simpleFinAccountFor(bankAccount,
			simplefin.Transaction{Id: "TRN-1", Posted: now.Unix(), Amount: "-1500", Description: "RAMEN"},
		)
		account.Currency = "JPY"
		account.Balance = "150000"
		account.AvailableBalance = nil

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		mock_simplefin.MockFetchAccounts(t, simpleFinAccountSet(account))

		runSimpleFinSync(t, ctrl, clock, enqueuer, bankAccount)

		transactions := readSimpleFinTransactions(t, bankAccount)
		require.Len(t, transactions, 1, "should create the transaction")
		assert.EqualValues(t, 1500, transactions[0].Amount, "JPY amounts should not be scaled")

		updatedBankAccount := testutils.MustDBRead(t, bankAccount)
		assert.EqualValues(t, 150000, updatedBankAccount.CurrentBalance, "JPY balance should not be scaled")
	})

	t.Run("pending transaction is skipped", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(now)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		enqueuer := givenAnyClusterEnqueue(ctrl, 1)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		mock_simplefin.MockFetchAccounts(t, simpleFinAccountSet(
			simpleFinAccountFor(bankAccount,
				simplefin.Transaction{Id: "TRN-1", Posted: now.Unix(), Amount: "-19.99", Description: "TACO BELL"},
				simplefin.Transaction{Id: "TRN-2", Posted: now.Unix(), Amount: "-5.00", Description: "COFFEE", Pending: true},
			),
		))

		runSimpleFinSync(t, ctrl, clock, enqueuer, bankAccount)

		transactions := readSimpleFinTransactions(t, bankAccount)
		require.Len(t, transactions, 1, "should only create the posted transaction")
		assert.Equal(t, "TRN-1", transactions[0].SimpleFinTransaction.SimpleFinId, "pending transaction should be skipped")
	})

	t.Run("enqueues clusters once per bank account", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(now)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		firstBankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)
		secondBankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		// Both bank accounts should go out in one bulk enqueue.
		enqueuer := mockgen.NewMockProcessor(ctrl)
		enqueuer.EXPECT().
			BulkEnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(similar_jobs.CalculateTransactionClusters),
				gomock.Any(),
				gomock.InAnyOrder([]any{
					similar_jobs.CalculateTransactionClustersArguments{
						AccountId:     firstBankAccount.AccountId,
						BankAccountId: firstBankAccount.BankAccountId,
					},
					similar_jobs.CalculateTransactionClustersArguments{
						AccountId:     secondBankAccount.AccountId,
						BankAccountId: secondBankAccount.BankAccountId,
					},
				}),
			).
			Return(nil).
			Times(1)

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		mock_simplefin.MockFetchAccounts(t, simpleFinAccountSet(
			simpleFinAccountFor(firstBankAccount,
				simplefin.Transaction{Id: "TRN-1", Posted: now.Unix(), Amount: "-1.00", Description: "FIRST"},
			),
			simpleFinAccountFor(secondBankAccount,
				simplefin.Transaction{Id: "TRN-2", Posted: now.Unix(), Amount: "-2.00", Description: "SECOND"},
			),
		))

		runSimpleFinSync(t, ctrl, clock, enqueuer, firstBankAccount)
	})
}

// givenAnyClusterEnqueue is for the tests that don't care which bank accounts
// the cluster jobs are queued for. Each successful sync with at least one bank
// account does one bulk enqueue, so times is the number of syncs.
func givenAnyClusterEnqueue(ctrl *gomock.Controller, times int) *mockgen.MockProcessor {
	enqueuer := mockgen.NewMockProcessor(ctrl)
	enqueuer.EXPECT().
		BulkEnqueueAt(
			gomock.Any(),
			mockqueue.EqQueue(similar_jobs.CalculateTransactionClusters),
			gomock.Any(),
			gomock.Any(),
		).
		Return(nil).
		Times(times)
	return enqueuer
}

// runSimpleFinSync runs one successful sync for the bank account's link against
// whatever is mocked in httpmock.
func runSimpleFinSync(
	t *testing.T,
	ctrl *gomock.Controller,
	clock *clock.Mock,
	enqueuer *mockgen.MockProcessor,
	bankAccount models.BankAccount,
) {
	log := testutils.GetLog(t)
	db := testutils.GetPgDatabase(t)

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
	context.EXPECT().KMS().Return(secrets.NewPlaintextKMS()).AnyTimes()
	context.EXPECT().DB().Return(db).AnyTimes()
	context.EXPECT().Log().Return(log).AnyTimes()
	context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
	context.EXPECT().Publisher().Return(pubsub.NewPostgresPubSub(log, db)).AnyTimes()
	context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

	err := simplefin_jobs.SyncSimpleFin(
		mockqueue.NewMockContext(context),
		simplefin_jobs.SyncSimpleFinArguments{
			AccountId: bankAccount.AccountId,
			LinkId:    bankAccount.LinkId,
		},
	)
	require.NoError(t, err, "must sync simplefin successfully")
}

func simpleFinAccountSet(accounts ...simplefin.Account) simplefin.AccountSet {
	connections := make([]simplefin.Connection, len(accounts))
	for i := range accounts {
		connections[i] = simplefin.Connection{
			ConnectionId: accounts[i].ConnectionId,
			Name:         "Lehman Brothers",
		}
	}
	return simplefin.AccountSet{
		Errlist:     []simplefin.Error{},
		Connections: connections,
		Accounts:    accounts,
	}
}

func simpleFinAccountFor(
	bankAccount models.BankAccount,
	transactions ...simplefin.Transaction,
) simplefin.Account {
	return simplefin.Account{
		Id:               bankAccount.SimpleFinBankAccount.SimpleFinId,
		Name:             "Primary Checking - 1234",
		ConnectionId:     bankAccount.SimpleFinBankAccount.ConnectionId,
		Currency:         "USD",
		Balance:          "1234.56",
		AvailableBalance: new(json.Number("1200.00")),
		BalanceDate:      time.Now().Unix(),
		Transactions:     transactions,
	}
}

// readSimpleFinTransactions returns every transaction on the bank account,
// deleted or not, ordered by their SimpleFIN Id.
func readSimpleFinTransactions(t *testing.T, bankAccount models.BankAccount) []models.Transaction {
	db := testutils.GetPgDatabase(t)
	transactions := make([]models.Transaction, 0)
	err := db.NewSelect().
		Model(&transactions).
		Relation("SimpleFinTransaction").
		Where(`"transaction"."account_id" = ?`, bankAccount.AccountId).
		Where(`"transaction"."bank_account_id" = ?`, bankAccount.BankAccountId).
		Order(`simple_fin_transaction.simplefin_id ASC`).
		Scan(t.Context())
	require.NoError(t, err, "must be able to read transactions")
	return transactions
}
