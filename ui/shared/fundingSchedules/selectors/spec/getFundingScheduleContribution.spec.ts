import Spending from 'models/Spending';
import { CHANGE_BANK_ACCOUNT } from 'shared/bankAccounts/actions';
import { getFundingScheduleContribution } from 'shared/fundingSchedules/selectors/getFundingScheduleContribution';
import { CreateSpending } from 'shared/spending/actions';
import { createTestStore } from 'testutils/store';

describe('getFundingScheduleContribution', () => {
  it('will return the total next contribution', () => {
    const store = createTestStore();

    // Set the current working bank account.
    store.dispatch({
      type: CHANGE_BANK_ACCOUNT,
      payload: 10,
    });

    // Include two spending objects that will be included in the total.
    store.dispatch({
      type: CreateSpending.Success,
      payload: new Spending({
        spendingId: 1000,
        fundingScheduleId: 100,
        bankAccountId: 10,
        nextContributionAmount: 1525, // $15.25
      }),
    });
    store.dispatch({
      type: CreateSpending.Success,
      payload: new Spending({
        spendingId: 1001,
        fundingScheduleId: 100,
        bankAccountId: 10,
        nextContributionAmount: 375, // $3.75
      }),
    });

    // Then include a spending object that will not be included because of its funding schedule.
    store.dispatch({
      type: CreateSpending.Success,
      payload: new Spending({
        spendingId: 1002,
        fundingScheduleId: 101,
        bankAccountId: 10,
        nextContributionAmount: 1254, // $12.54
      }),
    });

    // Then include a spending object that will not be included because it is paused.
    store.dispatch({
      type: CreateSpending.Success,
      payload: new Spending({
        spendingId: 1003,
        fundingScheduleId: 100,
        bankAccountId: 10,
        nextContributionAmount: 375, // $3.75
        isPaused: true, // Will cause this to be excluded.
      }),
    });

    const total = getFundingScheduleContribution(100)(store.getState());
    expect(total).toBe(1900);
  });
});
