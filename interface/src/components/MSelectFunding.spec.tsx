import { useFormikContext } from 'formik';

import { waitFor } from '@testing-library/react';

import MForm from '@monetr/interface/components/MForm';
import MSelectFunding from '@monetr/interface/components/MSelectFunding';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderer from '@monetr/interface/testutils/renderer';

describe('select funding schedule', () => {
  let mockFetch: FetchMock;

  beforeEach(() => {
    mockFetch = new FetchMock();
  });
  afterEach(() => {
    mockFetch.reset();
  });
  afterAll(() => {
    mockFetch.restore();
  });

  it('will pick the only funding schedule when nothing is picked', async () => {
    mockFetch.onGet('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/funding_schedules').reply(200, [
      {
        bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
        dateStarted: '2023-02-28T06:00:00Z',
        description: '15th and last day of every month',
        estimatedDeposit: null,
        excludeWeekends: false,
        fundingScheduleId: 'fund_01hy4re7c1xc2v44cf6kx302jx',
        lastRecurrence: '2023-09-29T05:00:00Z',
        name: 'Payday',
        nextRecurrence: '2023-10-13T05:00:00Z',
        nextRecurrenceOriginal: '2023-10-15T05:00:00Z',
        ruleset: 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
        waitForDeposit: false,
      },
    ]);

    const world = testRenderer(
      <MForm
        initialValues={{
          fundingScheduleId: '',
        }}
        onSubmit={() => {}}
      >
        <MSelectFunding label='Funding' name='fundingScheduleId' />
        <FundingValue />
      </MForm>,
      { initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/expenses' },
    );

    await waitFor(() =>
      expect(world.getByTestId('funding-value')).toHaveTextContent('fund_01hy4re7c1xc2v44cf6kx302jx'),
    );
  });

  it('will not replace a funding schedule that is already picked', async () => {
    mockFetch.onGet('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/funding_schedules').reply(200, [
      {
        bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
        dateStarted: '2023-02-28T06:00:00Z',
        description: '15th and last day of every month',
        estimatedDeposit: null,
        excludeWeekends: false,
        fundingScheduleId: 'fund_01hy4re7c1xc2v44cf6kx302jx',
        lastRecurrence: '2023-09-29T05:00:00Z',
        name: 'Payday',
        nextRecurrence: '2023-10-13T05:00:00Z',
        nextRecurrenceOriginal: '2023-10-15T05:00:00Z',
        ruleset: 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
        waitForDeposit: false,
      },
    ]);

    const world = testRenderer(
      <MForm
        initialValues={{
          fundingScheduleId: 'fund_01hy4rhqmy4wjy0vtrmqsc5c1m',
        }}
        onSubmit={() => {}}
      >
        <MSelectFunding label='Funding' name='fundingScheduleId' />
        <FundingValue />
      </MForm>,
      { initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/expenses' },
    );

    // Wait for the funding schedules to load so the effect has had its chance to run
    await waitFor(() => expect(world.getByPlaceholderText('Select a funding schedule...')).toBeInTheDocument());
    expect(world.getByTestId('funding-value')).toHaveTextContent('fund_01hy4rhqmy4wjy0vtrmqsc5c1m');
  });

  it('will not pick anything when there are a few to choose from', async () => {
    mockFetch.onGet('/api/bank_accounts/bac_01hy4rcmadc01d2kzv7vynbxxx/funding_schedules').reply(200, [
      {
        bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
        dateStarted: '2023-02-28T06:00:00Z',
        description: '15th and last day of every month',
        estimatedDeposit: null,
        excludeWeekends: false,
        fundingScheduleId: 'fund_01hy4re7c1xc2v44cf6kx302jx',
        lastRecurrence: '2023-09-29T05:00:00Z',
        name: 'Payday',
        nextRecurrence: '2023-10-13T05:00:00Z',
        nextRecurrenceOriginal: '2023-10-15T05:00:00Z',
        ruleset: 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
        waitForDeposit: false,
      },
      {
        bankAccountId: 'bac_01hy4rcmadc01d2kzv7vynbxxx',
        dateStarted: '2023-02-28T06:00:00Z',
        description: '15th and last day of every month',
        estimatedDeposit: null,
        excludeWeekends: false,
        fundingScheduleId: 'fund_01hy4rhqmy4wjy0vtrmqsc5c1m',
        lastRecurrence: '2023-09-29T05:00:00Z',
        name: 'Side gig',
        nextRecurrence: '2023-10-13T05:00:00Z',
        nextRecurrenceOriginal: '2023-10-15T05:00:00Z',
        ruleset: 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1',
        waitForDeposit: false,
      },
    ]);

    const world = testRenderer(
      <MForm
        initialValues={{
          fundingScheduleId: '',
        }}
        onSubmit={() => {}}
      >
        <MSelectFunding label='Funding' name='fundingScheduleId' />
        <FundingValue />
      </MForm>,
      { initialRoute: '/bank/bac_01hy4rcmadc01d2kzv7vynbxxx/expenses' },
    );

    await waitFor(() => expect(world.getByPlaceholderText('Select a funding schedule...')).toBeInTheDocument());
    expect(world.getByTestId('funding-value')).toBeEmptyDOMElement();
  });
});

// Just prints whatever is in the form so we can see what got picked
function FundingValue(): React.JSX.Element {
  const { values } = useFormikContext<{ fundingScheduleId: string }>();
  return <span data-testid='funding-value'>{values.fundingScheduleId}</span>;
}
