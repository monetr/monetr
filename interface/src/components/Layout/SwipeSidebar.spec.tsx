import { createPortal } from 'react-dom';
import { Link, useLocation } from 'wouter';

import { act, fireEvent, type RenderResult } from '@testing-library/react';

import SwipeSidebar from '@monetr/interface/components/Layout/SwipeSidebar';
import MSidebarToggle from '@monetr/interface/components/MSidebarToggle';
import { installMatchMedia, type MatchMediaControl } from '@monetr/interface/testutils/matchMedia';
import { installPointerEvents } from '@monetr/interface/testutils/pointerEvents';
import testRenderer from '@monetr/interface/testutils/renderer';
import { isSidebarGestureActive, setSidebarGestureActive } from '@monetr/interface/util/sidebarGesture';

const HOME = '/bank/bac_123/transactions';

function Panel(): React.JSX.Element {
  return (
    <div data-sidebar-secondary data-testid='panel-content'>
      <MSidebarToggle />
      {/* A real wouter Link so the same-route case behaves like the app: a pushState with an identical pathname and
      no re-render. */}
      <Link data-testid='panel-link' to={HOME}>
        Transactions
      </Link>
      {/* The account picker and link menu portal to the body while staying React children of the panel. */}
      {createPortal(
        <div data-testid='portalled'>
          <button type='button'>Pick an account</button>
        </div>,
        document.body,
      )}
    </div>
  );
}

// Lets a test navigate the way a <Redirect> or an action button would, without a click the drag suppression would
// swallow.
let navigateFromPage: ((to: string) => void) | null = null;

function Page(): React.JSX.Element {
  const [, navigate] = useLocation();
  navigateFromPage = navigate;
  return (
    <div>
      <MSidebarToggle className='page-toggle' />
      <button data-testid='page-button' type='button'>
        Page button
      </button>
      {/* Stands in for a transaction row, an anchor covering the whole thing. */}
      <a data-testid='page-row' href='/bank/bac_123/transactions/txn_1/details'>
        A row
      </a>
      <input data-testid='page-input' type='text' />
    </div>
  );
}

// jsdom reports every element as 0x0, which would make the panel measure 1px and the rubber band eat every drag.
const PANEL_WIDTH = 320;
// What paint writes at rest closed. The panel always carries a transform, so "never moved" means still this.
const CLOSED = `translate3d(-${PANEL_WIDTH}px,0,0)`;

function stubPanelWidth(): () => void {
  const original = Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'offsetWidth');
  Object.defineProperty(HTMLElement.prototype, 'offsetWidth', {
    configurable: true,
    get(this: HTMLElement) {
      return this.getAttribute('role') === 'dialog' ? PANEL_WIDTH : 0;
    },
  });
  return () => {
    if (original) {
      Object.defineProperty(HTMLElement.prototype, 'offsetWidth', original);
    }
  };
}

interface DragStep {
  x: number;
  y?: number;
}

// A drag is a pointerdown, some moves and a pointerup, all on the same pointerId. Nothing here waits for a frame: the
// commit decision reads the integrator's target, which pointermove sets synchronously.
//
// `msPerStep` is the fake gap between samples and decides which commit path gets exercised. The default is slow
// enough that velocity stays well under the 420px/s fling threshold, so the decision falls through to position.
function drag(
  element: Element,
  from: DragStep,
  steps: Array<DragStep>,
  options: { up?: boolean; msPerStep?: number } = {},
) {
  const base = { pointerId: 7, isPrimary: true, button: 0, pointerType: 'touch' };
  const gap = options.msPerStep ?? 500;
  let time = 1000;
  fireEvent.pointerDown(element, { ...base, clientX: from.x, clientY: from.y ?? 400, timeStamp: time });
  for (const step of steps) {
    time += gap;
    fireEvent.pointerMove(element, { ...base, clientX: step.x, clientY: step.y ?? from.y ?? 400, timeStamp: time });
  }
  if (options.up !== false) {
    const last = steps.at(-1) ?? from;
    fireEvent.pointerUp(element, { ...base, clientX: last.x, clientY: last.y ?? from.y ?? 400, timeStamp: time });
  }
}

function renderSidebar(): RenderResult {
  return testRenderer(
    <SwipeSidebar panel={<Panel />}>
      <Page />
    </SwipeSidebar>,
    { initialRoute: HOME },
  );
}

// The toggle in the page is the one a user actually presses; the one inside the panel is only reachable once it's
// already open.
function pageToggle(world: RenderResult): HTMLElement {
  const toggle = world.container.querySelector('.page-toggle');
  if (!(toggle instanceof HTMLElement)) {
    throw new Error('page toggle not rendered');
  }
  return toggle;
}

function panelElement(world: RenderResult): HTMLElement {
  const panel = world.getByTestId('panel-content').parentElement;
  if (!panel) {
    throw new Error('panel not rendered');
  }
  return panel;
}

describe('SwipeSidebar', () => {
  let media: MatchMediaControl;
  let restorePointerEvents: () => void;

  beforeEach(() => {
    restorePointerEvents = installPointerEvents();
  });

  afterEach(() => {
    media?.restore();
    restorePointerEvents();
    navigateFromPage = null;
    document.body.style.overflow = '';
    document.documentElement.removeAttribute('data-sidebar-past');
  });

  describe('on desktop', () => {
    beforeEach(() => {
      media = installMatchMedia({ mobile: false });
    });

    it('will not present the panel as a dialog', () => {
      const world = renderSidebar();
      const panel = panelElement(world);

      // display: contents keeps the element in the accessibility tree, so a permanent dialog role would hide the
      // real sidebars from screen readers.
      expect(panel).not.toHaveAttribute('role');
      expect(panel).not.toHaveAttribute('aria-hidden');
      expect(panel).not.toHaveAttribute('aria-modal');
    });

    it('will not write a single inline style', () => {
      const world = renderSidebar();

      for (const element of Array.from(world.container.querySelectorAll('*'))) {
        expect(element).not.toHaveAttribute('style');
      }
    });

    it('will not lock scrolling or mark the page inert', () => {
      const world = renderSidebar();

      expect(document.body.style.overflow).toBe('');
      expect(world.getByTestId('page-button').closest('[inert]')).toBeNull();
    });
  });

  describe('on mobile', () => {
    beforeEach(() => {
      media = installMatchMedia({ mobile: true });
    });

    it('will start closed and hidden from assistive technology', () => {
      const world = renderSidebar();
      const panel = panelElement(world);

      expect(panel).toHaveAttribute('role', 'dialog');
      expect(panel).toHaveAttribute('aria-hidden', 'true');
      expect(panel).not.toHaveAttribute('aria-modal');
      expect(document.body.style.overflow).toBe('');
    });

    it('will open from the toggle, and lock and inert the page behind it', () => {
      const world = renderSidebar();

      fireEvent.click(pageToggle(world));

      const panel = panelElement(world);
      expect(panel).toHaveAttribute('aria-hidden', 'false');
      expect(panel).toHaveAttribute('aria-modal', 'true');
      expect(document.body.style.overflow).toBe('hidden');
      expect(world.getByTestId('page-button').closest('[inert]')).not.toBeNull();
    });

    it('will close again from the same toggle', () => {
      const world = renderSidebar();

      fireEvent.click(pageToggle(world));
      fireEvent.click(pageToggle(world));

      expect(panelElement(world)).toHaveAttribute('aria-hidden', 'true');
      expect(document.body.style.overflow).toBe('');
      expect(world.getByTestId('page-button').closest('[inert]')).toBeNull();
    });

    it('will move focus into the panel and hand it back on close', () => {
      const world = renderSidebar();
      const toggle = pageToggle(world);
      toggle.focus();
      expect(document.activeElement).toBe(toggle);

      fireEvent.click(toggle);
      expect(document.activeElement).toBe(panelElement(world));

      fireEvent.click(pageToggle(world));
      // Grabbed when open() ran, not in an effect; React commits `inert` before effects and that blurs this.
      expect(document.activeElement).toBe(toggle);
    });

    it('will close on escape', () => {
      const world = renderSidebar();
      fireEvent.click(pageToggle(world));

      fireEvent.keyDown(panelElement(world), { key: 'Escape' });

      expect(panelElement(world)).toHaveAttribute('aria-hidden', 'true');
    });

    it('will leave escape alone when it belongs to a layer outside the panel', () => {
      const world = renderSidebar();
      fireEvent.click(pageToggle(world));

      // A dropdown or bottom drawer opened from inside the panel portals to the body and owns escape while it's up.
      fireEvent.keyDown(document.body, { key: 'Escape' });

      expect(panelElement(world)).toHaveAttribute('aria-hidden', 'false');
    });

    it('will close when a link in the panel is tapped', () => {
      const world = renderSidebar();
      fireEvent.click(pageToggle(world));

      // Same route, so the pathname doesn't change and the close-on-navigation effect never runs.
      fireEvent.click(world.getByTestId('panel-link'));

      expect(panelElement(world)).toHaveAttribute('aria-hidden', 'true');
    });

    it('will publish its position for the toggle icon', () => {
      const world = renderSidebar();
      expect(document.documentElement.dataset.sidebarPast).toBe('false');

      fireEvent.click(pageToggle(world));
      expect(document.documentElement.dataset.sidebarPast).toBe('true');
    });
  });

  describe('the gesture', () => {
    let restoreWidth: () => void;

    beforeEach(() => {
      media = installMatchMedia({ mobile: true });
      restoreWidth = stubPanelWidth();
    });

    afterEach(() => {
      restoreWidth();
      setSidebarGestureActive(false);
    });

    it('will open on a horizontal drag from the middle of the page', () => {
      const world = renderSidebar();
      // Nowhere near the left edge; the gesture is armed across the whole page.
      drag(world.getByTestId('page-row'), { x: 200 }, [{ x: 260 }, { x: 340 }]);

      expect(panelElement(world)).toHaveAttribute('aria-hidden', 'false');
    });

    it('will revert a slow drag that does not reach the commit point', () => {
      const world = renderSidebar();
      // 60px of 320 is 0.19, under the 0.42 threshold, and slow enough that velocity gets no say.
      drag(world.getByTestId('page-row'), { x: 200 }, [{ x: 230 }, { x: 260 }]);

      expect(panelElement(world)).toHaveAttribute('aria-hidden', 'true');
    });

    it('will open on a flick that never reaches the commit point', () => {
      const world = renderSidebar();
      // The same 60px but thrown: 30px per 16ms is ~1875px/s, well past the fling threshold, so direction wins.
      drag(world.getByTestId('page-row'), { x: 200 }, [{ x: 230 }, { x: 260 }], { msPerStep: 16 });

      expect(panelElement(world)).toHaveAttribute('aria-hidden', 'false');
    });

    it('will leave a vertical drag alone', () => {
      const world = renderSidebar();
      const panel = panelElement(world);
      // Enough sideways movement to open if the axis resolved wrong, but far more downward.
      drag(world.getByTestId('page-row'), { x: 200, y: 700 }, [
        { x: 260, y: 500 },
        { x: 340, y: 200 },
      ]);

      expect(panel).toHaveAttribute('aria-hidden', 'true');
      // The panel never moved at all. The gesture flag after pointerup proves nothing, releaseDrag clears it
      // whatever the axis was.
      expect(panel.style.transform).toBe(CLOSED);
    });

    it('will refuse a rightward drag that is not clearly horizontal', () => {
      const world = renderSidebar();
      // 120 across and 120 down, outside the cone.
      drag(world.getByTestId('page-row'), { x: 200, y: 400 }, [{ x: 320, y: 520 }]);

      expect(panelElement(world)).toHaveAttribute('aria-hidden', 'true');
    });

    it('will leave leftward drags in the page unclaimed', () => {
      const world = renderSidebar();
      const panel = panelElement(world);
      drag(world.getByTestId('page-row'), { x: 300 }, [{ x: 200 }, { x: 100 }], { up: false });

      // Checked while the finger is still down; after pointerup an unclaimed and a claimed-then-settled drag look
      // identical.
      expect(isSidebarGestureActive()).toBe(false);
      expect(panel.style.transform).toBe(CLOSED);
    });

    it('will close on a leftward drag once open', () => {
      const world = renderSidebar();
      fireEvent.click(pageToggle(world));
      expect(panelElement(world)).toHaveAttribute('aria-hidden', 'false');

      // Being open puts every surface in loose mode, so leftward closes it from anywhere.
      drag(panelElement(world), { x: 300 }, [{ x: 200 }, { x: 60 }]);

      expect(panelElement(world)).toHaveAttribute('aria-hidden', 'true');
    });

    it('will put the panel back out of sight after a press that barely moved', () => {
      const world = renderSidebar();
      const panel = panelElement(world);
      expect(panel.style.visibility).toBe('hidden');

      // One pixel is under the slop, so no axis resolves. The press still promoted layers and made the panel
      // visible, and stopping the loop without a render would leave it that way.
      drag(world.getByTestId('page-row'), { x: 200 }, [{ x: 201 }]);

      expect(panel.style.visibility).toBe('hidden');
      expect(panel).toHaveAttribute('aria-hidden', 'true');
    });

    it('will carry on to the nearest end when a press interrupts an animation', async () => {
      // Reduced motion has to be off or the open is instant and there is nothing in flight to interrupt.
      media.restore();
      media = installMatchMedia({ mobile: true, reducedMotion: false });
      const world = renderSidebar();
      const panel = panelElement(world);

      fireEvent.click(pageToggle(world));
      // Partway through the spring.
      await new Promise(resolve => setTimeout(resolve, 60));
      const midFlight = panel.style.transform;
      expect(midFlight).not.toBe('translate3d(0px,0,0)');

      // pointerdown freezes the motion where it is; a press that never becomes a drag must not strand it there.
      await act(async () => {
        drag(panel, { x: 200 }, [{ x: 201 }]);
        await new Promise(resolve => setTimeout(resolve, 800));
      });

      expect(panel.style.transform).toBe('translate3d(0px,0,0)');
      expect(panel).toHaveAttribute('aria-hidden', 'false');
    });

    it('will not arm over a form control', () => {
      const world = renderSidebar();
      drag(world.getByTestId('page-input'), { x: 200 }, [{ x: 300 }, { x: 380 }]);

      expect(panelElement(world)).toHaveAttribute('aria-hidden', 'true');
    });

    it('will suppress the click that ends a drag, without letting the browser navigate', () => {
      const world = renderSidebar();
      const row = world.getByTestId('page-row');
      drag(row, { x: 200 }, [{ x: 260 }, { x: 340 }]);

      const click = new MouseEvent('click', { bubbles: true, cancelable: true, detail: 1 });
      row.dispatchEvent(click);

      // preventDefault is the half that matters. wouter's Link calls preventDefault in its own onClick, so only
      // stopping propagation would let the browser do a native navigation, a full reload.
      expect(click.defaultPrevented).toBe(true);
    });

    it('will not eat a keyboard activation left over from a touch drag', () => {
      const world = renderSidebar();
      // A touch drag preventDefaults the touchmove, so no click is synthesised and the suppression stays armed.
      drag(world.getByTestId('page-row'), { x: 200 }, [{ x: 260 }, { x: 340 }]);

      // Keyboard activations arrive with detail 0. Dispatched on a button rather than a link, since wouter's Link
      // calls preventDefault itself.
      const activation = new MouseEvent('click', { bubbles: true, cancelable: true, detail: 0 });
      world.getByTestId('page-button').dispatchEvent(activation);

      expect(activation.defaultPrevented).toBe(false);
    });

    it('will let a tap through untouched', () => {
      const world = renderSidebar();
      const row = world.getByTestId('page-row');
      // A press and release with no movement isn't a drag.
      drag(row, { x: 200 }, []);

      const click = new MouseEvent('click', { bubbles: true, cancelable: true, detail: 1 });
      row.dispatchEvent(click);

      expect(click.defaultPrevented).toBe(false);
    });

    it('will ignore pointers from portalled content that is not a DOM descendant', () => {
      const world = renderSidebar();
      fireEvent.click(pageToggle(world));
      const panel = panelElement(world);
      const portalled = world.getByTestId('portalled');

      // A real portal, so React routes these events to the panel's handlers even though the node lives under
      // document.body. Flicking through the account list must not drag the sidebar underneath it.
      expect(panel.contains(portalled)).toBe(false);
      drag(portalled, { x: 300 }, [{ x: 200 }, { x: 60 }]);

      expect(panel).toHaveAttribute('aria-hidden', 'false');
    });

    it('will ignore a second pointer while a drag is live', () => {
      const world = renderSidebar();
      const row = world.getByTestId('page-row');
      const base = { isPrimary: true, button: 0, pointerType: 'touch', clientY: 400 };

      fireEvent.pointerDown(row, { ...base, pointerId: 1, clientX: 200 });
      fireEvent.pointerMove(row, { ...base, pointerId: 1, clientX: 340 });
      // A second finger must not reset the drag origin.
      fireEvent.pointerDown(row, { ...base, pointerId: 2, clientX: 100 });
      fireEvent.pointerMove(row, { ...base, pointerId: 2, clientX: 110 });
      fireEvent.pointerUp(row, { ...base, pointerId: 1, clientX: 340 });

      expect(panelElement(world)).toHaveAttribute('aria-hidden', 'false');
    });

    it('will release a live drag when the route changes underneath it', () => {
      const world = renderSidebar();
      const row = world.getByTestId('page-row');
      drag(row, { x: 200 }, [{ x: 300 }], { up: false });
      expect(isSidebarGestureActive()).toBe(true);

      // A <Redirect> or an action button can navigate with a finger still down. Left live, the drag either gets the
      // settle spring swapped in mid gesture or carries on from a stale origin after a snap.
      act(() => navigateFromPage?.('/bank/bac_123/expenses'));

      expect(isSidebarGestureActive()).toBe(false);
    });

    it('will clear the gesture flag when unmounted mid drag', () => {
      const world = renderSidebar();
      drag(world.getByTestId('page-row'), { x: 200 }, [{ x: 300 }], { up: false });
      expect(isSidebarGestureActive()).toBe(true);

      world.unmount();

      // Left set, this kills pull to refresh for the rest of the session.
      expect(isSidebarGestureActive()).toBe(false);
      expect(document.body.style.overflow).toBe('');
      expect(document.body.style.userSelect).toBe('');
    });
  });

  describe('crossing the breakpoint', () => {
    beforeEach(() => {
      media = installMatchMedia({ mobile: true });
    });

    it('will clear everything it wrote when the viewport grows', () => {
      const world = renderSidebar();
      fireEvent.click(pageToggle(world));
      expect(panelElement(world)).toHaveAttribute('style');

      act(() => media.setMobile(false));

      // visibility inherits, so a leftover `hidden` here would blank out the real desktop sidebars.
      const panel = panelElement(world);
      expect(panel.style.visibility).toBe('');
      expect(panel.style.transform).toBe('');
      expect(document.body.style.overflow).toBe('');
      expect(document.documentElement).not.toHaveAttribute('data-sidebar-past');
      expect(panel).not.toHaveAttribute('role');
    });

    it('will not tear down the page when the viewport grows', () => {
      const world = renderSidebar();
      const before = world.getByTestId('page-button');

      act(() => media.setMobile(false));

      // Same DOM node, so React didn't unmount the routed page. Swapping components here would reset every page's
      // state on a tablet rotation or a window resize.
      expect(world.getByTestId('page-button')).toBe(before);
    });
  });
});
