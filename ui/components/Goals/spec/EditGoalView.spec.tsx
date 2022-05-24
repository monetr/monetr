import React from 'react';
import { EditGoalView } from 'components/Goals/EditGoalView';
import FundingSchedule from 'models/FundingSchedule';
import Spending from 'models/Spending';
import moment from 'moment';
import testRenderer from 'testutils/renderer';
import { Map } from 'immutable';
import { screen } from '@testing-library/react';

describe('EditGoalsView', () => {
  it('will probably not render', () => {
    const fundingSchedules = Map<number, FundingSchedule>().set(1, new FundingSchedule({
      fundingScheduleId: 1,
      bankAccountId: 1,
      name: 'Payday',
      nextOccurrence: moment(),
    }));
    const goal = new Spending({
      spendingId: 1,
      bankAccountId: 1,
      fundingScheduleId: 1,
      name: 'Test Goal',
      currentAmount: 0,
      usedAmount: 0,
      targetAmount: 100,
      nextRecurrence: moment().add(1, 'day'),
    });
    expect(goal.getGoalIsInProgress()).toBeTruthy(); /// Required to render view.

    testRenderer(<EditGoalView
      goal={ goal }
      fundingSchedules={ fundingSchedules }
      updateSpending={ (spending) => Promise.resolve() }
      hideView={ () => {
      } }
    />);
    expect(screen.getByTestId('funding-schedule-selector')).toBeInTheDocument();
  });
});
