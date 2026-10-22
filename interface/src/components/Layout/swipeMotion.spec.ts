import { rs } from '@rstest/core';

import {
  clamp,
  commitFraction,
  commitMinimum,
  commitThreshold,
  createMotion,
  type Motion,
  type MotionCallbacks,
  resolveAxis,
  rubber,
  type Sample,
  slopEdge,
  slopPage,
  swipeBlocked,
  timeStep,
  velocityOf,
  velocityWindow,
} from '@monetr/interface/components/Layout/swipeMotion';

// jsdom runs requestAnimationFrame off of real time. This swaps it for a queue so tests can control every frame.
interface FrameHarness {
  motion: Motion;
  rendered: Array<number>;
  restCount: () => number;
  errors: Array<unknown>;
  // The first frame after start() only sets the clock and doesn't step. Tests counting steps need count + 1.
  advance: (count: number, delta?: number) => void;
  pending: () => boolean;
}

function harness(overrides: Partial<MotionCallbacks> = {}): FrameHarness {
  const queue: Array<FrameRequestCallback> = [];
  let now = 0;
  let rests = 0;
  const rendered: Array<number> = [];
  const errors: Array<unknown> = [];

  globalThis.requestAnimationFrame = ((callback: FrameRequestCallback) => {
    queue.push(callback);
    return queue.length;
  }) as typeof globalThis.requestAnimationFrame;
  globalThis.cancelAnimationFrame = (() => {
    queue.length = 0;
  }) as typeof globalThis.cancelAnimationFrame;

  const motion = createMotion({
    render: (position: number) => rendered.push(position),
    onRest: () => {
      rests += 1;
    },
    onError: (error: unknown) => errors.push(error),
    ...overrides,
  });

  return {
    motion,
    rendered,
    errors,
    restCount: () => rests,
    pending: () => queue.length > 0,
    advance: (count: number, delta: number = 1000 / 60) => {
      for (let i = 0; i < count; i++) {
        const callback = queue.shift();
        if (!callback) {
          return;
        }
        now += delta;
        callback(now);
      }
    },
  };
}

describe('swipeMotion', () => {
  describe('clamp', () => {
    it('will hold a value between its bounds', () => {
      expect(clamp(0.5, 0, 1)).toBe(0.5);
      expect(clamp(-3, 0, 1)).toBe(0);
      expect(clamp(9, 0, 1)).toBe(1);
    });
  });

  describe('rubber', () => {
    it('will never exceed the dimension it is resisting against', () => {
      for (const over of [1, 10, 100, 1000, 100000]) {
        expect(rubber(over, 320)).toBeLessThan(320);
      }
    });

    it('will increase monotonically but with diminishing returns', () => {
      const first = rubber(10, 320);
      const second = rubber(20, 320);
      const third = rubber(30, 320);
      expect(second).toBeGreaterThan(first);
      expect(third).toBeGreaterThan(second);
      // Every extra 10px of drag moves the panel less than the last one did.
      expect(third - second).toBeLessThan(second - first);
    });

    it('will not move at all at the boundary', () => {
      expect(rubber(0, 320)).toBe(0);
    });
  });

  describe('commitThreshold', () => {
    it('will use the plain fraction on a full width panel', () => {
      // 0.42 * 335 is 141px, well above the floor.
      expect(commitThreshold(335)).toBeCloseTo(0.42, 5);
    });

    it('will apply the pixel floor on a mid width panel', () => {
      // 0.42 * 100 is 42px which is below the floor. The floor wins.
      expect(commitThreshold(100) * 100).toBeCloseTo(commitMinimum, 5);
    });

    it('will cap the floor rather than make a narrow panel unopenable', () => {
      // On the 64px rail the floor would be 56 of 64px, the cap brings that back to about 38px.
      const threshold = commitThreshold(64);
      expect(threshold).toBeCloseTo(0.6, 5);
      expect(threshold * 64).toBeCloseTo(38.4, 5);
      // Travel past the slop goes from 14.9px to 26.4px.
      expect(commitFraction * 64 - slopPage).toBeCloseTo(14.88, 2);
      expect(threshold * 64 - slopPage).toBeCloseTo(26.4, 2);
    });

    it('will not divide by zero on an unmeasured panel', () => {
      expect(commitThreshold(0)).toBe(0.42);
    });
  });

  describe('resolveAxis', () => {
    it('will wait until movement passes the slop', () => {
      expect(resolveAxis('loose', slopEdge - 1, 0)).toBeNull();
      expect(resolveAxis('page', slopPage - 1, 0)).toBeNull();
    });

    it('will use a tighter slop at the edge than out in the page', () => {
      // The same movement arms at the edge but is still undecided out in the page.
      const dx = (slopEdge + slopPage) / 2;
      expect(resolveAxis('loose', dx, 0)).toBe('x');
      expect(resolveAxis('page', dx, 0)).toBeNull();
    });

    it('will only claim rightward drags out in the page', () => {
      expect(resolveAxis('page', 40, 0)).toBe('x');
      // Leftward stays with the page so swipe to dismiss rows can have it later.
      expect(resolveAxis('page', -40, 0)).toBe('y');
    });

    it('will claim leftward drags in loose mode, which is how an open drawer closes', () => {
      expect(resolveAxis('loose', -40, 0)).toBe('x');
    });

    it('will refuse a rightward drag in the page that is not clearly horizontal', () => {
      // 40 across and 35 down, past the slop but outside the cone.
      expect(resolveAxis('page', 40, 35)).toBe('y');
      expect(resolveAxis('page', 40, 5)).toBe('x');
    });

    it('will give a vertical drag straight back to the page', () => {
      expect(resolveAxis('loose', 0, 40)).toBe('y');
      expect(resolveAxis('page', 0, 40)).toBe('y');
    });
  });

  describe('velocityOf', () => {
    it('will report nothing without at least two samples', () => {
      expect(velocityOf([])).toBe(0);
      expect(velocityOf([{ x: 0, t: 0 }])).toBe(0);
    });

    it('will report pixels per second', () => {
      const samples: Array<Sample> = [
        { x: 0, t: 0 },
        { x: 50, t: 50 },
      ];
      expect(velocityOf(samples)).toBe(1000);
    });

    it('will ignore samples older than the window', () => {
      // The first sample is way outside the window and would halve the result if we counted it.
      const samples: Array<Sample> = [
        { x: 0, t: 0 },
        { x: 0, t: velocityWindow * 4 },
        { x: 50, t: velocityWindow * 4 + 50 },
      ];
      expect(velocityOf(samples)).toBe(1000);
    });

    it('will not be dominated by a single trailing sample', () => {
      // A finger moving steadily that happens to emit one tiny event right at the end.
      const steady: Array<Sample> = [
        { x: 0, t: 0 },
        { x: 30, t: 30 },
        { x: 60, t: 60 },
        { x: 61, t: 62 },
      ];
      // Two sample velocity would say 500 here. Over the window its close to 1000.
      expect(velocityOf(steady)).toBeGreaterThan(900);
    });

    it('will not divide by zero when every sample shares a timestamp', () => {
      expect(
        velocityOf([
          { x: 0, t: 5 },
          { x: 20, t: 5 },
        ]),
      ).toBe(0);
    });
  });

  describe('swipeBlocked', () => {
    afterEach(() => {
      document.body.replaceChildren();
    });

    function mount(html: string): Element {
      const root = document.createElement('div');
      root.innerHTML = html;
      document.body.append(root);
      return root;
    }

    it('will not block a plain anchor', () => {
      const root = mount('<a href="/somewhere" id="link">row</a>');
      const link = root.querySelector('#link');
      expect(link).not.toBeNull();
      expect((link as HTMLAnchorElement).draggable).toBeTruthy();
      expect(swipeBlocked(link, root)).toBeFalsy();
    });

    it('will block an element explicitly marked draggable', () => {
      const root = mount('<div draggable="true" id="target">handle</div>');
      expect(swipeBlocked(root.querySelector('#target'), root)).toBeTruthy();
    });

    it('will block anything under data-no-swipe', () => {
      const root = mount('<div data-no-swipe><span id="target">nope</span></div>');
      expect(swipeBlocked(root.querySelector('#target'), root)).toBeTruthy();
    });

    it('will block form controls', () => {
      const root = mount('<input id="text" type="text" /><textarea id="area"></textarea><select id="pick"></select>');
      expect(swipeBlocked(root.querySelector('#text'), root)).toBeTruthy();
      expect(swipeBlocked(root.querySelector('#area'), root)).toBeTruthy();
      expect(swipeBlocked(root.querySelector('#pick'), root)).toBeTruthy();
    });

    it('will block a horizontal scroller', () => {
      // jsdom reports every element as 0x0, we need to stub the size.
      const root = mount('<div id="scroller" style="overflow-x: auto"><span id="target">wide</span></div>');
      const scroller = root.querySelector('#scroller');
      expect(scroller).not.toBeNull();
      Object.defineProperty(scroller, 'scrollWidth', { configurable: true, value: 900 });
      Object.defineProperty(scroller, 'clientWidth', { configurable: true, value: 300 });
      expect(swipeBlocked(root.querySelector('#target'), root)).toBeTruthy();
    });

    it('will not block an element that merely clips its overflow', () => {
      // Truncated text overflows without being a scroller, Typography does this on most rows.
      const root = mount('<div id="clipped" style="overflow: hidden"><span id="target">ellipsized</span></div>');
      const clipped = root.querySelector('#clipped');
      Object.defineProperty(clipped, 'scrollWidth', { configurable: true, value: 900 });
      Object.defineProperty(clipped, 'clientWidth', { configurable: true, value: 300 });
      expect(swipeBlocked(root.querySelector('#target'), root)).toBeFalsy();
    });

    it('will stop walking at the root it was given', () => {
      const outer = document.createElement('div');
      outer.setAttribute('data-no-swipe', '');
      const root = document.createElement('div');
      const target = document.createElement('span');
      root.append(target);
      outer.append(root);
      document.body.append(outer);
      // The attribute is above the root and should not be checked.
      expect(swipeBlocked(target, root)).toBeFalsy();
    });
  });

  describe('createMotion', () => {
    // harness() replaces requestAnimationFrame. Put it back after each test.
    let originalRaf: typeof globalThis.requestAnimationFrame;
    let originalCancel: typeof globalThis.cancelAnimationFrame;

    beforeEach(() => {
      originalRaf = globalThis.requestAnimationFrame;
      originalCancel = globalThis.cancelAnimationFrame;
    });

    afterEach(() => {
      globalThis.requestAnimationFrame = originalRaf;
      globalThis.cancelAnimationFrame = originalCancel;
    });

    it('will start at rest and do nothing until asked', () => {
      const world = harness();
      expect(world.motion.location()).toBe(0);
      expect(world.motion.isRunning()).toBeFalsy();
      expect(world.pending()).toBeFalsy();
    });

    it('will converge on the target and land exactly', () => {
      const world = harness();
      world.motion.release(1);
      expect(world.motion.isRunning()).toBeTruthy();

      world.advance(200);

      expect(world.motion.location()).toBe(1);
      expect(world.motion.target()).toBe(1);
      expect(world.motion.isRunning()).toBeFalsy();
      expect(world.restCount()).toBe(1);
      expect(world.rendered.at(-1)).toBe(1);
    });

    it('will run the same physics regardless of frame cadence', () => {
      // Compare at the same simulated time, not the same number of frames. We use 10.5 steps instead of a whole
      // number, this way floating point error at a step boundary can't change how many steps each one ran.
      const simulated = 10.5;
      const positions = [0.25, 0.5, 0.75, 1.5].map(multiple => {
        const world = harness();
        world.motion.release(1);
        world.advance(simulated / multiple + 1, timeStep * multiple);
        return world.motion.location();
      });

      for (const position of positions) {
        expect(position).toBeCloseTo(positions[0] ?? 0, 6);
      }
      // If it already landed then this test wouldn't prove anything.
      expect(positions[0]).toBeGreaterThan(0.5);
      expect(positions[0]).toBeLessThan(0.999);
    });

    it('will interpolate between physics steps rather than jumping', () => {
      const world = harness();
      world.motion.release(1);
      // The clock frame, then a frame worth 1.5 steps. One step runs and half a step is left over. We should paint
      // halfway between the last two steps.
      world.advance(1);
      world.advance(1, timeStep * 1.5);

      const painted = world.rendered.at(-1) ?? 0;
      const physics = world.motion.location();
      expect(physics).toBeGreaterThan(0);
      expect(painted).toBeCloseTo(physics / 2, 10);
      expect(painted).toBeLessThan(physics);
    });

    it('will not owe several seconds of physics after a long stall', () => {
      const stalled = harness();
      stalled.motion.release(1);
      // A clock frame, then one frame ten seconds later. That would be 600 steps without the cap.
      stalled.advance(2, 10_000);

      const normal = harness();
      normal.motion.release(1);
      // A clock frame plus 5 normal ones.
      normal.advance(6);

      // This isn't an exact match. Floating point error can leave the stalled frame one step short.
      expect(stalled.motion.location()).toBeLessThanOrEqual(normal.motion.location());
      expect(stalled.motion.location()).toBeGreaterThan(normal.motion.location() * 0.75);
      expect(stalled.motion.location()).toBeLessThan(0.75);
      expect(stalled.motion.isRunning()).toBeTruthy();
    });

    it('will carry drag velocity through the release with no hand-off', () => {
      const world = harness();
      world.motion.beginDrag();
      // Walk the finger across a step at a time.
      for (let i = 1; i <= 20; i++) {
        world.motion.setTarget(i / 40);
        world.advance(1);
      }

      const before = world.motion.location();
      // The drag loop is already running. This is one real step.
      world.motion.release(1);
      world.advance(1);
      const flung = world.motion.location() - before;

      // The same step from the same position but with no velocity. release() starts a new loop and needs the
      // extra clock frame.
      const control = harness();
      control.motion.snap(before);
      control.motion.release(1);
      control.advance(2);
      const fromRest = control.motion.location() - before;

      // Both ran one step from the same place, the difference is the speed the finger had.
      expect(flung).toBeGreaterThan(fromRest * 1.2);
    });

    it('will trail the finger while dragging rather than pinning to it', () => {
      const world = harness();
      world.motion.beginDrag();
      world.motion.setTarget(1);
      // A clock frame plus one real step.
      world.advance(2);

      expect(world.motion.location()).toBeGreaterThan(0);
      expect(world.motion.location()).toBeLessThan(1);
      expect(world.motion.target()).toBe(1);
    });

    it('will freeze an in flight animation exactly where it is', () => {
      const world = harness();
      world.motion.release(1);
      world.advance(4);
      const caught = world.motion.location();

      world.motion.hold();
      expect(world.motion.isRunning()).toBeFalsy();
      expect(world.motion.location()).toBe(caught);
      expect(world.motion.target()).toBe(caught);

      // No further frames are scheduled.
      world.advance(10);
      expect(world.motion.location()).toBe(caught);
    });

    it('will report rest exactly once per animation', () => {
      const world = harness();
      world.motion.release(1);
      world.advance(200);
      expect(world.restCount()).toBe(1);

      world.advance(50);
      expect(world.restCount()).toBe(1);
    });

    it('will jump and report rest when snapped', () => {
      const world = harness();
      world.motion.release(1);
      world.advance(4);

      world.motion.snap(0);
      expect(world.motion.location()).toBe(0);
      expect(world.motion.target()).toBe(0);
      expect(world.motion.isRunning()).toBeFalsy();
      // We resize the panel when it rests. A snap needs to report rest too.
      expect(world.restCount()).toBe(1);
      expect(world.rendered.at(-1)).toBe(0);
    });

    it('will snap instead of animating when reduced motion is preferred', () => {
      const matchMedia = rs.fn().mockImplementation((query: string) => ({
        matches: query.includes('prefers-reduced-motion'),
        media: query,
        addEventListener: rs.fn(),
        removeEventListener: rs.fn(),
      }));
      const original = window.matchMedia;
      Object.defineProperty(window, 'matchMedia', { configurable: true, writable: true, value: matchMedia });

      try {
        const world = harness();
        world.motion.release(1);
        expect(world.motion.location()).toBe(1);
        expect(world.motion.isRunning()).toBeFalsy();
        expect(world.restCount()).toBe(1);
      } finally {
        Object.defineProperty(window, 'matchMedia', { configurable: true, writable: true, value: original });
      }
    });

    it('will stop cleanly when destroyed mid animation', () => {
      const world = harness();
      world.motion.release(1);
      world.advance(3);
      world.motion.destroy();

      expect(world.motion.isRunning()).toBeFalsy();
      const painted = world.rendered.length;
      world.advance(10);
      expect(world.rendered.length).toBe(painted);
    });

    it('will contain a throw from render instead of dying silently', () => {
      let shouldThrow = false;
      const world = harness({
        render: () => {
          if (shouldThrow) {
            throw new Error('boom');
          }
        },
      });

      world.motion.release(1);
      world.advance(2);
      shouldThrow = true;
      world.advance(1);

      expect(world.errors).toHaveLength(1);
      expect(world.motion.isRunning()).toBeFalsy();
      expect(world.pending()).toBeFalsy();
    });

    it('will contain a throw from a manual repaint', () => {
      const world = harness({
        render: () => {
          throw new Error('boom');
        },
      });

      world.motion.repaint();

      expect(world.errors).toHaveLength(1);
      expect(world.motion.isRunning()).toBeFalsy();
    });
  });
});
