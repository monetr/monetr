/*
  @license
  The fixed timestep integrator below (createMotion) is adapted from Embla Carousel
  https://github.com/davidjerleke/embla-carousel

  MIT License

  Copyright (c) David Jerleke.

  Permission is hereby granted, free of charge, to any person obtaining a copy
  of this software and associated documentation files (the "Software"), to deal
  in the Software without restriction, including without limitation the rights
  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
  copies of the Software, and to permit persons to whom the Software is
  furnished to do so, subject to the following conditions:

  The above copyright notice and this permission notice shall be included in all
  copies or substantial portions of the Software.

  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
  SOFTWARE.
  @license-end
*/

/*
  Motion and gesture helpers for the mobile sidebar. Nothing in here imports React or knows what a sidebar is, which
  keeps the integrator testable on its own.

  Dragging never sets the position directly. A pointermove only moves `target`, and a fixed timestep integrator moves
  the real position toward it, which absorbs uneven touch sampling instead of turning it into judder. Rendering
  interpolates between the last two physics steps so it stays smooth on displays that arent 60Hz.

  One integrator handles both dragging and settling; only friction and duration change between them:

    velocity += (target - location) / duration
    velocity *= friction
    location += velocity

  Dragging uses a slack pair so the position chases the finger, releasing uses a stiff pair so it lands without
  wobbling. Velocity carries across the switch, so a fling needs no hand off.
*/

// Distance from the left edge, in px, that uses the looser slop and angle rules.
export const EDGE_ZONE = 28;
// Movement in px before an axis is picked. Tighter at the edge than out in the page.
export const SLOP_EDGE = 6;
export const SLOP_PAGE = 12;
// Out in the page |dx| must beat |dy| by this ratio, roughly a 35 degree cone.
export const ANGLE = 1.4;
// px/s. Anything faster commits based on direction instead of position.
export const FLING = 420;
// Fraction of the panel width past which a release opens.
export const COMMIT = 0.42;
// Floor in px, so a narrow panel doesn't commit on ~15px of travel.
export const COMMIT_MIN_PX = 56;
// Cap as a fraction of width, so the floor can't make a narrow panel unopenable. See commitThreshold.
export const COMMIT_MAX_FRACTION = 0.6;
// Pointer history window used to measure a fling, in ms.
export const VELOCITY_WINDOW = 90;
// How long layers stay promoted after everything settles, in ms.
export const LAYER_HOLD = 450;
// Scrim opacity when the drawer is fully open.
export const SCRIM_MAX = 0.62;

// Dragging chases the finger with a slack pair; releasing swaps in a stiff pair just inside critical damping so it
// lands without wobbling.
const DRAG_FRICTION = 0.3;
const DRAG_DURATION = 0.75;
const SETTLE_FRICTION = 0.5;
const SETTLE_DURATION = 6;

// Fixed physics timestep. Exported so tests can drive frames at fractional multiples of it.
export const TIME_STEP = 1000 / 60;
// Cap on catch-up steps so a stalled tab can't spiral.
const MAX_CATCHUP = 5;
const REST = 0.0006;

export type DragMode = 'loose' | 'page';
export type Axis = 'x' | 'y';

export interface Sample {
  x: number;
  t: number;
}

export function clamp(value: number, low: number, high: number): number {
  if (value < low) {
    return low;
  }
  if (value > high) {
    return high;
  }
  return value;
}

// How far past an edge something should actually move after being dragged `over` px beyond it. Approaches `dimension`
// without reaching it, so the end feels like resistance instead of a wall.
export function rubber(over: number, dimension: number, c: number = 0.5): number {
  return (1 - 1 / ((over * c) / dimension + 1)) * dimension;
}

// The fraction of the panel width past which letting go opens instead of reverting. See COMMIT_MIN_PX.
export function commitThreshold(width: number): number {
  if (width <= 0) {
    return COMMIT;
  }
  const px = Math.min(Math.max(COMMIT * width, COMMIT_MIN_PX), COMMIT_MAX_FRACTION * width);
  return px / width;
}

// Checked per call so flipping the OS setting doesn't need a reload.
export function prefersReducedMotion(): boolean {
  if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') {
    return false;
  }
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

// Short haptic when a gesture commits. iOS doesn't support vibrate and some permissions policies throw.
export function tick(): void {
  try {
    navigator.vibrate?.(6);
  } catch {
    // Vibration is a nicety, never let it break a gesture.
  }
}

// Walks up from the touched element and reports whether something already owns the horizontal axis: a horizontal
// scroller, a form control, a live selection, or an explicit [data-no-swipe]. Any future component that wants the
// horizontal axis needs to be one of those.
export function swipeBlocked(target: EventTarget | null, root: Element | null): boolean {
  const selection = window.getSelection?.();
  if (selection && selection.rangeCount > 0 && !selection.isCollapsed) {
    return true;
  }

  for (let node = target; node instanceof Element; node = node.parentElement) {
    if (node.hasAttribute('data-no-swipe')) {
      return true;
    }
    // The attribute, not the property. The `draggable` property is true for every <a href>, and mobile list rows are
    // covered by absolutely positioned links, so the property would kill the gesture on most of the app.
    if (node.getAttribute('draggable') === 'true') {
      return true;
    }
    // isContentEditable lives on HTMLElement and this walk can pass through SVG.
    if (node instanceof HTMLElement && node.isContentEditable) {
      return true;
    }

    const tag = node.tagName;
    if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' || tag === 'VIDEO') {
      return true;
    }

    if (node.scrollWidth - node.clientWidth > 1) {
      const overflowX = getComputedStyle(node).overflowX;
      if (overflowX === 'auto' || overflowX === 'scroll') {
        return true;
      }
    }

    if (node === root) {
      break;
    }
  }

  return false;
}

// 'x' opens/closes, 'y' goes back to the page, null means keep waiting.
export function resolveAxis(mode: DragMode, dx: number, dy: number): Axis | null {
  const ax = Math.abs(dx);
  const ay = Math.abs(dy);
  const slop = mode === 'page' ? SLOP_PAGE : SLOP_EDGE;
  if (ax < slop && ay < slop) {
    return null;
  }

  if (mode === 'page') {
    // Rightward only and inside the cone. Leftward drags stay with the page so swipe to dismiss can have them later.
    return dx > 0 && ax > ay * ANGLE ? 'x' : 'y';
  }

  return ax > ay ? 'x' : 'y';
}

// Measured over a short window rather than the last two samples, which would let a single trailing event decide the
// whole fling.
export function velocityOf(samples: Array<Sample>): number {
  const last = samples.at(-1);
  const previous = samples.at(-2);
  if (!last || !previous) {
    return 0;
  }

  let first = previous;
  for (let i = samples.length - 3; i >= 0; i--) {
    const sample = samples[i];
    if (!sample || last.t - sample.t > VELOCITY_WINDOW) {
      break;
    }
    first = sample;
  }

  const dt = last.t - first.t;
  return dt > 0 ? ((last.x - first.x) / dt) * 1000 : 0;
}

export interface Motion {
  // Where the physics currently is. Trails the finger on purpose.
  location(): number;
  // Where the finger actually is. Commit decisions should read this one.
  target(): number;
  // True while a drag is live or an animation is running.
  isRunning(): boolean;
  // Freeze an in flight animation exactly where it is.
  hold(): void;
  beginDrag(): void;
  setTarget(next: number): void;
  release(next: number): void;
  cancelDrag(): void;
  // Jump straight to a position and report rest, no animation.
  snap(next: number): void;
  repaint(): void;
  destroy(): void;
}

export interface MotionCallbacks {
  render: (position: number) => void;
  onRest: () => void;
  // Called if render or the integrator throws, after the loop has been stopped. The handler is responsible for
  // getting the UI back to something usable.
  onError: (error: unknown) => void;
}

// Adapted from Embla Carousel's animation loop and scroll body, see the license at the top of the file. The fixed
// timestep clock and the friction/duration scroll body are theirs. The API around it is ours: snap, destroy,
// isRunning, the capped catch up, and containing a throw rather than letting it kill the loop.
export function createMotion(callbacks: MotionCallbacks): Motion {
  let location = 0;
  let previous = 0;
  let target = 0;
  let velocity = 0;
  let friction = SETTLE_FRICTION;
  let duration = SETTLE_DURATION;
  let dragging = false;
  let raf = 0;
  let last = 0;
  let accumulated = 0;

  function step(): void {
    previous = location;
    velocity += (target - location) / duration;
    velocity *= friction;
    location += velocity;
  }

  function settled(): boolean {
    if (dragging) {
      return false;
    }
    return Math.abs(target - location) < REST && Math.abs(velocity) < REST;
  }

  function stop(): void {
    if (raf) {
      cancelAnimationFrame(raf);
    }
    raf = 0;
    last = 0;
    accumulated = 0;
  }

  function fail(error: unknown): void {
    stop();
    dragging = false;
    velocity = 0;
    callbacks.onError(error);
  }

  function frame(now: number): void {
    try {
      if (!last) {
        last = now;
      }
      let elapsed = now - last;
      last = now;
      if (elapsed > TIME_STEP * MAX_CATCHUP) {
        elapsed = TIME_STEP * MAX_CATCHUP;
      }

      accumulated += elapsed;
      while (accumulated >= TIME_STEP) {
        step();
        accumulated -= TIME_STEP;
      }

      if (settled()) {
        location = target;
        previous = target;
        velocity = 0;
        callbacks.render(location);
        stop();
        callbacks.onRest();
        return;
      }

      callbacks.render(previous + (location - previous) * (accumulated / TIME_STEP));
    } catch (error) {
      fail(error);
      return;
    }

    raf = requestAnimationFrame(frame);
  }

  function start(): void {
    if (raf) {
      return;
    }
    last = 0;
    raf = requestAnimationFrame(frame);
  }

  function settle(next: number): void {
    stop();
    dragging = false;
    friction = SETTLE_FRICTION;
    duration = SETTLE_DURATION;
    location = next;
    previous = next;
    target = next;
    velocity = 0;
    try {
      callbacks.render(location);
    } catch (error) {
      fail(error);
      return;
    }
    callbacks.onRest();
  }

  return {
    location: () => location,
    target: () => target,
    isRunning: () => raf !== 0 || dragging,

    hold() {
      stop();
      dragging = false;
      target = location;
      previous = location;
      velocity = 0;
    },

    beginDrag() {
      dragging = true;
      friction = DRAG_FRICTION;
      duration = DRAG_DURATION;
      velocity = 0;
      start();
    },

    setTarget(next: number) {
      target = next;
    },

    release(next: number) {
      if (prefersReducedMotion()) {
        settle(next);
        return;
      }
      dragging = false;
      friction = SETTLE_FRICTION;
      duration = SETTLE_DURATION;
      target = next;
      // No velocity argument. The integrator is already moving at whatever speed the finger was, so a fling just
      // carries on by itself.
      start();
    },

    cancelDrag() {
      dragging = false;
      friction = SETTLE_FRICTION;
      duration = SETTLE_DURATION;
      if (settled()) {
        stop();
      }
    },

    snap(next: number) {
      settle(next);
    },

    repaint() {
      try {
        callbacks.render(location);
      } catch (error) {
        fail(error);
      }
    },

    destroy() {
      stop();
      dragging = false;
    },
  };
}
