import { useFormikContext } from 'formik';

import { waitFor } from '@testing-library/react';

import MForm from '@monetr/interface/components/MForm';
import SelectFrequency from '@monetr/interface/components/SelectFrequency';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderer from '@monetr/interface/testutils/renderer';

interface FrequencyValues {
  date: Date | null;
  ruleset: string | null;
}

describe('select frequency', () => {
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

  it('will keep a stored rule that matches none of the options as custom', async () => {
    // Every third Thursday is not one of the generated options for this date.
    const ruleset = 'FREQ=WEEKLY;INTERVAL=3;BYDAY=TH';
    const world = testRenderer(
      <MForm<FrequencyValues>
        initialValues={{
          date: new Date(2023, 2, 30),
          ruleset,
        }}
        onSubmit={() => {}}
      >
        <SelectFrequency<FrequencyValues> dateFrom='date' label='Frequency' name='ruleset' />
        <RulesetValue />
      </MForm>,
    );

    await waitFor(() => expect(world.getByDisplayValue('Custom (every 3 weeks on Thursday)')).toBeVisible());
    expect(world.getByTestId('ruleset-value')).toHaveTextContent(ruleset);
  });
});

function RulesetValue(): React.JSX.Element {
  const { values } = useFormikContext<FrequencyValues>();
  return <span data-testid='ruleset-value'>{values.ruleset}</span>;
}
