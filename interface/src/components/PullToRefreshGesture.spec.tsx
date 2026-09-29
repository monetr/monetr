import type React from 'react';
import { useRef } from 'react';
import { rs } from '@rstest/core';
import { QueryClient } from '@tanstack/react-query';

import { fireEvent } from '@testing-library/react';

import PullToRefreshGesture from '@monetr/interface/components/PullToRefreshGesture';
import testRenderer from '@monetr/interface/testutils/renderer';

// jsdom 26 makes window.location non-configurable. To mock location.reload, we spy on jsdom's internal implementation
// via its symbol property.
const implSymbol = Reflect.ownKeys(window.location).find(i => typeof i === 'symbol');
if (!implSymbol) {
  throw new Error('jsdom implementation symbol not found on window.location');
}

// The real app renders PullToRefreshGesture through PullToRefresh which waits for the page to finish loading first, so
// in the tests we just render it directly with our own root.
function PullHarness(): React.JSX.Element {
  const rootRef = useRef<HTMLDivElement>(null);
  return (
    <div ref={rootRef}>
      <div data-testid='page'>page content</div>
      <div data-ptr-ignore data-testid='ignored'>
        ignored content
      </div>
      <PullToRefreshGesture rootRef={rootRef} />
    </div>
  );
}

// These distances are how far the finger moves, not how far the page moves. The rubber band makes the page move less
// than the finger so these need to be a good bit bigger than the thresholds in PullToRefreshGesture.
const SHORT_PULL = 50; // Page moves about 38px, not enough to do anything.
const REFRESH_PULL = 200; // Page moves about 89px, past the refresh threshold but not the reload one.
const RELOAD_PULL = 400; // Page moves about 114px, past the reload threshold.

const START_Y = 100;

function touchAt(y: number) {
  const touch = { identifier: 0, clientX: 0, clientY: y };
  return { touches: [touch], changedTouches: [touch] };
}

function pull(element: Element, distance: number) {
  fireEvent.touchStart(element, touchAt(START_Y));
  fireEvent.touchMove(element, touchAt(START_Y + distance));
}

function release(element: Element, distance: number) {
  fireEvent.touchEnd(element, { touches: [], changedTouches: [touchAt(START_Y + distance).changedTouches[0]] });
}

describe('pull to refresh', () => {
  let refetchSpy: ReturnType<typeof rs.spyOn>;
  let reloadSpy: ReturnType<typeof rs.spyOn>;

  beforeEach(() => {
    refetchSpy = rs.spyOn(QueryClient.prototype, 'refetchQueries').mockResolvedValue(undefined);
    // biome-ignore lint/suspicious/noExplicitAny: :middlefinder:
    reloadSpy = rs.spyOn((window.location as any)[implSymbol], 'reload').mockImplementation(() => {});
  });

  afterEach(() => {
    refetchSpy.mockRestore();
    reloadSpy.mockRestore();
  });

  it('will refresh when pulled past the threshold', () => {
    const world = testRenderer(<PullHarness />);
    const page = world.getByTestId('page');

    pull(page, REFRESH_PULL);
    expect(world.getByText('Release to refresh')).toBeInTheDocument();

    release(page, REFRESH_PULL);
    expect(refetchSpy).toHaveBeenCalledTimes(1);
    expect(reloadSpy).not.toHaveBeenCalled();
  });

  it('will reload the whole page when pulled really far', () => {
    const world = testRenderer(<PullHarness />);
    const page = world.getByTestId('page');

    pull(page, RELOAD_PULL);
    expect(world.getByText('Release to reload')).toBeInTheDocument();

    release(page, RELOAD_PULL);
    expect(reloadSpy).toHaveBeenCalledTimes(1);
    expect(refetchSpy).not.toHaveBeenCalled();
  });

  it('will not do anything for a short pull', () => {
    const world = testRenderer(<PullHarness />);
    const page = world.getByTestId('page');

    pull(page, SHORT_PULL);
    expect(world.getByText('Pull to refresh')).toBeInTheDocument();

    release(page, SHORT_PULL);
    expect(refetchSpy).not.toHaveBeenCalled();
    expect(reloadSpy).not.toHaveBeenCalled();
  });

  it('will cancel the pull if they drag back up past where they started', () => {
    const world = testRenderer(<PullHarness />);
    const page = world.getByTestId('page');

    pull(page, REFRESH_PULL);
    // Drag back up above where the finger started.
    fireEvent.touchMove(page, touchAt(START_Y - 10));
    release(page, -10);
    expect(refetchSpy).not.toHaveBeenCalled();
    expect(reloadSpy).not.toHaveBeenCalled();
  });

  it('will not start a pull when the page is scrolled down', () => {
    const world = testRenderer(<PullHarness />);
    const page = world.getByTestId('page');

    Object.defineProperty(window, 'scrollY', { value: 100, configurable: true });
    try {
      pull(page, REFRESH_PULL);
      release(page, REFRESH_PULL);
    } finally {
      Object.defineProperty(window, 'scrollY', { value: 0, configurable: true });
    }
    expect(refetchSpy).not.toHaveBeenCalled();
  });

  it('will not start a pull while a dialog is open', () => {
    const world = testRenderer(<PullHarness />);
    const page = world.getByTestId('page');

    const dialog = document.createElement('div');
    dialog.setAttribute('role', 'dialog');
    document.body.appendChild(dialog);
    try {
      pull(page, REFRESH_PULL);
      release(page, REFRESH_PULL);
    } finally {
      dialog.remove();
    }
    expect(refetchSpy).not.toHaveBeenCalled();
  });

  it('will not start a pull from inside data-ptr-ignore', () => {
    const world = testRenderer(<PullHarness />);
    const ignored = world.getByTestId('ignored');

    pull(ignored, REFRESH_PULL);
    release(ignored, REFRESH_PULL);
    expect(refetchSpy).not.toHaveBeenCalled();
  });

  // Touch events keep getting sent to the element the finger first touched even after its removed from the page, and at
  // that point they dont bubble up anymore. This used to leave the page stuck pulled down forever.
  it('will still finish the pull if the touched element gets removed', () => {
    const world = testRenderer(<PullHarness />);
    const page = world.getByTestId('page');

    pull(page, REFRESH_PULL);
    page.remove();
    release(page, REFRESH_PULL);
    expect(refetchSpy).toHaveBeenCalledTimes(1);
  });

  it('will still let them pull again after the touched element gets removed', () => {
    const world = testRenderer(<PullHarness />);
    const page = world.getByTestId('page');
    const ignored = world.getByTestId('ignored');
    // Take the ignore off so we can pull from this one after the first element is gone.
    ignored.removeAttribute('data-ptr-ignore');

    pull(page, SHORT_PULL);
    page.remove();
    release(page, SHORT_PULL);
    expect(refetchSpy).not.toHaveBeenCalled();

    pull(ignored, REFRESH_PULL);
    release(ignored, REFRESH_PULL);
    expect(refetchSpy).toHaveBeenCalledTimes(1);
  });
});
