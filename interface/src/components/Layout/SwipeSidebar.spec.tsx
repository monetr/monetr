import { createPortal } from 'react-dom';
import { Link, useLocation } from 'wouter';

import { act, fireEvent, type RenderResult } from '@testing-library/react';

import SwipeSidebar from '@monetr/interface/components/Layout/SwipeSidebar';
import MSidebarToggle from '@monetr/interface/components/MSidebarToggle';
import { installMatchMedia, type MatchMediaControl } from '@monetr/interface/testutils/matchMedia';
import { installPointerEvents } from '@monetr/interface/testutils/pointerEvents';
import testRenderer from '@monetr/interface/testutils/renderer';
import { isSidebarGestureActive, setSidebarGestureActive } from '@monetr/interface/util/sidebarGesture';

const homeRoute = '/bank/bac_123/transactions';

function Panel(): React.JSX.Element {
  return (
    <div data-sidebar-secondary data-testid='panel-content'>
      <MSidebarToggle />
      <Link data-testid='panel-link' to={homeRoute}>
        Transactions
      </Link>
      {createPortal(
        <div data-testid='portalled'>
          <button type='button'>Pick an account</button>
        </div>,
        document.body,
      )}
    </div>
  );
}

// This lets a test navigate like a <Redirect> would, without a click.
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
      <a data-testid='page-row' href='/bank/bac_123/transactions/txn_1/details'>
        A row
      </a>
      <input data-testid='page-input' type='text' />
    </div>
  );
}

// jsdom reports every element as 0x0, which would make the panel 1px wide.
const panelWidth = 320;
// What paint writes when the drawer is closed.
const closedTransform = `translate3d(-${panelWidth}px,0,0)`;

function stubPanelWidth(): () => void {
  const original = Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'offsetWidth');
  Object.defineProperty(HTMLElement.prototype, 'offsetWidth', {
    configurable: true,
    get(this: HTMLElement) {
      return this.getAttribute('role') === 'dialog' ? panelWidth : 0;
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

// The default msPerStep is slow enough that it never counts as a fling, so the drawer opens or closes based on how
// far it was dragged.
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
    { initialRoute: homeRoute },
  );
}

// The toggle in the page, not the one in the panel.
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
      // This has to be remembered in open(), by the time an effect runs inert has already blurred it.
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

      // A dropdown opened from the panel portals to the body and handles its own escape.
      fireEvent.keyDown(document.body, { key: 'Escape' });

      expect(panelElement(world)).toHaveAttribute('aria-hidden', 'false');
    });

    it('will close when a link in the panel is tapped', () => {
      const world = renderSidebar();
      fireEvent.click(pageToggle(world));

      // Same route, so the pathname doesn't change.
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
      // Nowhere near the left edge.
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
      // The same 60px but thrown. 30px per 16ms is about 1875px/s, well past the fling threshold.
      drag(world.getByTestId('page-row'), { x: 200 }, [{ x: 230 }, { x: 260 }], { msPerStep: 16 });

      expect(panelElement(world)).toHaveAttribute('aria-hidden', 'false');
    });

    it('will leave a vertical drag alone', () => {
      const world = renderSidebar();
      const panel = panelElement(world);
      drag(world.getByTestId('page-row'), { x: 200, y: 700 }, [
        { x: 260, y: 500 },
        { x: 340, y: 200 },
      ]);

      expect(panel).toHaveAttribute('aria-hidden', 'true');
      // Check the panel and not the gesture flag, the flag is always cleared on pointerup.
      expect(panel.style.transform).toBe(closedTransform);
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

      // Check while the finger is still down, after pointerup the flag is always cleared.
      expect(isSidebarGestureActive()).toBeFalsy();
      expect(panel.style.transform).toBe(closedTransform);
    });

    it('will close on a leftward drag once open', () => {
      const world = renderSidebar();
      fireEvent.click(pageToggle(world));
      expect(panelElement(world)).toHaveAttribute('aria-hidden', 'false');

      drag(panelElement(world), { x: 300 }, [{ x: 200 }, { x: 60 }]);

      expect(panelElement(world)).toHaveAttribute('aria-hidden', 'true');
    });

    it('will put the panel back out of sight after a press that barely moved', () => {
      const world = renderSidebar();
      const panel = panelElement(world);
      expect(panel.style.visibility).toBe('hidden');

      // One pixel is under the slop. But the press still made the panel visible, so it needs to be hidden again.
      drag(world.getByTestId('page-row'), { x: 200 }, [{ x: 201 }]);

      expect(panel.style.visibility).toBe('hidden');
      expect(panel).toHaveAttribute('aria-hidden', 'true');
    });

    it('will carry on to the nearest end when a press interrupts an animation', async () => {
      // Reduced motion needs to be off, otherwise there is no animation to interrupt.
      media.restore();
      media = installMatchMedia({ mobile: true, reducedMotion: false });
      const world = renderSidebar();
      const panel = panelElement(world);

      fireEvent.click(pageToggle(world));
      // Partway through the spring.
      await new Promise(resolve => setTimeout(resolve, 60));
      const midFlight = panel.style.transform;
      expect(midFlight).not.toBe('translate3d(0px,0,0)');

      // pointerdown freezes the animation, a press that isn't a drag should let it finish.
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

      // If this isn't prevented then the browser follows the link and reloads the page.
      expect(click.defaultPrevented).toBeTruthy();
    });

    it('will not eat a keyboard activation left over from a touch drag', () => {
      const world = renderSidebar();
      // A touch drag doesn't make a click, so the suppression is still set after it.
      drag(world.getByTestId('page-row'), { x: 200 }, [{ x: 260 }, { x: 340 }]);

      // Keyboard clicks have a detail of 0. This uses a button since wouter's Link prevents default on its own.
      const activation = new MouseEvent('click', { bubbles: true, cancelable: true, detail: 0 });
      world.getByTestId('page-button').dispatchEvent(activation);

      expect(activation.defaultPrevented).toBeFalsy();
    });

    it('will let a tap through untouched', () => {
      const world = renderSidebar();
      const row = world.getByTestId('page-row');
      drag(row, { x: 200 }, []);

      const click = new MouseEvent('click', { bubbles: true, cancelable: true, detail: 1 });
      row.dispatchEvent(click);

      expect(click.defaultPrevented).toBeFalsy();
    });

    it('will ignore pointers from portalled content that is not a DOM descendant', () => {
      const world = renderSidebar();
      fireEvent.click(pageToggle(world));
      const panel = panelElement(world);
      const portalled = world.getByTestId('portalled');

      // React sends these to the panel's handlers even though the portal is under document.body.
      expect(panel.contains(portalled)).toBeFalsy();
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
      expect(isSidebarGestureActive()).toBeTruthy();

      // A <Redirect> can navigate while a finger is still down.
      act(() => navigateFromPage?.('/bank/bac_123/expenses'));

      expect(isSidebarGestureActive()).toBeFalsy();
    });

    it('will clear the gesture flag when unmounted mid drag', () => {
      const world = renderSidebar();
      drag(world.getByTestId('page-row'), { x: 200 }, [{ x: 300 }], { up: false });
      expect(isSidebarGestureActive()).toBeTruthy();

      world.unmount();

      // If this stayed set then pull to refresh would be broken until a reload.
      expect(isSidebarGestureActive()).toBeFalsy();
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

      // A leftover visibility: hidden would hide the desktop sidebars.
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

      // Same DOM node, so the page wasn't remounted. Otherwise rotating a tablet would reset the page.
      expect(world.getByTestId('page-button')).toBe(before);
    });
  });
});
