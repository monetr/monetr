import { waitFor } from '@testing-library/react';

import Transactions from '@monetr/interface/pages/transactions';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import apiSampleResponses from '@monetr/interface/testutils/fixtures/apiSampleResponses';
import testRenderer from '@monetr/interface/testutils/renderer';

describe('transactions page', () => {
  let mockFetch: FetchMock;
  // Only restore the spies made here, restoring every mock would also undo the matchMedia mock from the test setup.
  const spies: Array<{ mockRestore: () => void }> = [];

  beforeEach(() => {
    mockFetch = new FetchMock();
    apiSampleResponses(mockFetch);
    // jsdom doesn't do any layout so every day in the virtualized transaction list would measure as 0 tall, which makes
    // it impossible for the virtualizer to tell which days are actually on screen. Pretend each day has a real height.
    spies.push(
      rs.spyOn(HTMLElement.prototype, 'offsetHeight', 'get').mockImplementation(function (this: HTMLElement) {
        return this.matches('li[data-index]') ? 100 : 0;
      }),
    );
    // The virtualizer scrolls the window itself, jsdom doesn't implement that and complains loudly.
    spies.push(rs.spyOn(window, 'scrollTo').mockImplementation(() => {}));
  });
  afterEach(() => {
    mockFetch.reset();
    for (const spy of spies.splice(0)) {
      spy.mockRestore();
    }
    window.history.replaceState(null, '');
  });
  afterAll(() => {
    mockFetch.restore();
  });

  it('will render the days at the saved scroll offset when going back', async () => {
    // Going back to the list lands on a history entry that remembers where they had scrolled to.
    window.history.replaceState({ transactionsScrollOffset: 1500 }, '');

    const world = testRenderer(<Transactions />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions',
    });

    // The days at the bottom of the list should be rendered right away, not the ones at the top of the page.
    await waitFor(() => expect(world.container.querySelector('li[data-index="18"]')).toBeInTheDocument());
    expect(world.container.querySelector('li[data-index="0"]')).not.toBeInTheDocument();
    expect(window.scrollTo).toHaveBeenCalledWith({ top: 1500, behavior: undefined });
  });

  it('will remember the scroll offset on the history entry', async () => {
    const world = testRenderer(<Transactions />, {
      initialRoute: '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/transactions',
    });

    await waitFor(() => expect(world.container.querySelector('li[data-index="0"]')).toBeInTheDocument());

    spies.push(rs.spyOn(window, 'scrollY', 'get').mockReturnValue(800));
    window.dispatchEvent(new Event('scroll'));

    await waitFor(() => expect(window.history.state?.transactionsScrollOffset).toBe(800));
  });
});
