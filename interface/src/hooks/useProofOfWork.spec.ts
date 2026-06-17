import { rs } from '@rstest/core';

import { useProofOfWork } from '@monetr/interface/hooks/useProofOfWork';
import FetchMock from '@monetr/interface/testutils/fetchMock';
import testRenderHook from '@monetr/interface/testutils/hooks';

describe('use proof of work', () => {
  let mockFetch: FetchMock;
  // We drive Date.now ourselves so we can fast forward past a challenge's ttl
  // without actually waiting.
  let now: number;

  beforeEach(() => {
    mockFetch = new FetchMock();
    now = 1_000_000;
    rs.spyOn(Date, 'now').mockImplementation(() => now);
  });
  afterEach(() => {
    mockFetch.reset();
    rs.restoreAllMocks();
  });
  afterAll(() => {
    mockFetch.restore();
  });

  function challengeRequests(): number {
    return (mockFetch.history.post ?? []).filter(entry => entry.url === '/api/authentication/challenge').length;
  }

  it('will not fetch a challenge when disabled', async () => {
    const world = testRenderHook(() => useProofOfWork('login', false), { initialRoute: '/login' });

    const solution = await world.result.current.getSolution();
    expect(solution).toBeNull();
    expect(challengeRequests()).toBe(0);
  });

  it('will not fetch a challenge on mount', async () => {
    // The whole point of the lazy hook. Nothing is requested until we warm up or actually ask for a solution. The login
    // page can remount for a few milliseconds during the post-login navigation and we must not burn a challenge on that.
    mockFetch.onPost('/api/authentication/challenge').reply(200, { challenge: 'x', difficulty: 0, ttl: 60 });

    const world = testRenderHook(() => useProofOfWork('login', true), { initialRoute: '/login' });

    expect(challengeRequests()).toBe(0);

    // But it still works the moment we actually ask for one.
    await world.result.current.getSolution();
    expect(challengeRequests()).toBe(1);
  });

  it('will warm up and solve only once for repeated warmups', async () => {
    // warmup is wired to the form's onInput so it fires on every keystroke. It must only ever line up one challenge
    // while the current one is still fresh.
    mockFetch.onPost('/api/authentication/challenge').reply(200, { challenge: 'x', difficulty: 0, ttl: 60 });

    const world = testRenderHook(() => useProofOfWork('login', true), { initialRoute: '/login' });

    world.result.current.warmup();
    world.result.current.warmup();
    world.result.current.warmup();

    const solution = await world.result.current.getSolution();
    expect(solution).toMatchObject({ challenge: 'x', nonce: 0 });
    expect(challengeRequests()).toBe(1);
  });

  it('will fetch and solve on first use and reuse it while it is still fresh', async () => {
    // Difficulty 0 means the solver returns immediately.
    mockFetch.onPost('/api/authentication/challenge').reply(200, { challenge: 'x', difficulty: 0, ttl: 60 });

    const world = testRenderHook(() => useProofOfWork('login', true), { initialRoute: '/login' });

    const solution = await world.result.current.getSolution();
    expect(solution).toMatchObject({ challenge: 'x', nonce: 0 });
    expect(challengeRequests()).toBe(1);

    // Still well within the 60 second ttl, so a second call must reuse the already
    // solved challenge rather than fetching another.
    await world.result.current.getSolution();
    expect(challengeRequests()).toBe(1);
  });

  it('will fetch a fresh challenge once the solved one goes stale', async () => {
    mockFetch.onPost('/api/authentication/challenge').reply(200, { challenge: 'x', difficulty: 0, ttl: 60 });

    const world = testRenderHook(() => useProofOfWork('login', true), { initialRoute: '/login' });

    await world.result.current.getSolution();
    expect(challengeRequests()).toBe(1);

    // The user sat idle past the 60 second ttl. The next solution request should
    // grab and solve a brand new challenge instead of handing back the stale one.
    now += 61 * 1000;
    const solution = await world.result.current.getSolution();
    expect(solution).toMatchObject({ challenge: 'x', nonce: 0 });
    expect(challengeRequests()).toBe(2);
  });
});
