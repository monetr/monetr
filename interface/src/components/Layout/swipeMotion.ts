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

// Tuning for the gesture. Distances are in px, times in ms and velocity in px per second.
export const edgeZone = 28;
export const slopEdge = 6;
export const slopPage = 12;
export const swipeAngle = 1.4;
export const flingVelocity = 420;
export const commitFraction = 0.42;
export const commitMinimum = 56;
export const commitMaximumFraction = 0.6;
export const velocityWindow = 90;
export const layerHold = 450;
export const scrimOpacity = 0.62;
export const timeStep = 1000 / 60;

const dragFriction = 0.3;
const dragDuration = 0.75;
const settleFriction = 0.5;
const settleDuration = 6;
const maxCatchup = 5;
const restDistance = 0.0006;

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

// rubber will take how far something has been dragged past an edge and return how far it should actually move. It
// gets closer to dimension the further they drag but never reaches it.
export function rubber(over: number, dimension: number, c: number = 0.5): number {
  return (1 - 1 / ((over * c) / dimension + 1)) * dimension;
}

// commitThreshold will return the fraction of the panel width past which letting go opens the drawer. On a narrow
// panel 42% would only be about 15px, so there is a floor in px. But the floor is capped too, otherwise a really
// narrow panel could never be opened.
export function commitThreshold(width: number): number {
  if (width <= 0) {
    return commitFraction;
  }
  const px = Math.min(Math.max(commitFraction * width, commitMinimum), commitMaximumFraction * width);
  return px / width;
}

export function prefersReducedMotion(): boolean {
  if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') {
    return false;
  }
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

export function tick(): void {
  try {
    navigator.vibrate?.(6);
  } catch {
    // iOS doesn't support vibrate and some permissions policies throw, we don't care either way.
  }
}

export function swipeBlocked(target: EventTarget | null, root: Element | null): boolean {
  const selection = window.getSelection?.();
  if (selection && selection.rangeCount > 0 && !selection.isCollapsed) {
    return true;
  }

  for (let node = target; node instanceof Element; node = node.parentElement) {
    if (node.hasAttribute('data-no-swipe')) {
      return true;
    }
    // Check the attribute and not the draggable property. The property is true for every <a href>, and the rows in the
    // mobile lists are covered by links, so the property would block the swipe on most of the app.
    if (node.getAttribute('draggable') === 'true') {
      return true;
    }
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

// resolveAxis returns x when the drag should move the drawer, y when it belongs to the page, and null when it has not
// moved far enough to tell yet.
export function resolveAxis(mode: DragMode, dx: number, dy: number): Axis | null {
  const ax = Math.abs(dx);
  const ay = Math.abs(dy);
  const slop = mode === 'page' ? slopPage : slopEdge;
  if (ax < slop && ay < slop) {
    return null;
  }

  // Out in the page we only take drags to the right that are clearly horizontal. Drags to the left stay with the page
  // so we can use them for swipe to dismiss later.
  if (mode === 'page') {
    return dx > 0 && ax > ay * swipeAngle ? 'x' : 'y';
  }

  return ax > ay ? 'x' : 'y';
}

// Measure over a short window instead of just the last two samples. Otherwise one late event can decide the whole
// fling.
export function velocityOf(samples: Array<Sample>): number {
  const last = samples.at(-1);
  const previous = samples.at(-2);
  if (!last || !previous) {
    return 0;
  }

  let first = previous;
  for (let i = samples.length - 3; i >= 0; i--) {
    const sample = samples[i];
    if (!sample || last.t - sample.t > velocityWindow) {
      break;
    }
    first = sample;
  }

  const dt = last.t - first.t;
  return dt > 0 ? ((last.x - first.x) / dt) * 1000 : 0;
}

export interface Motion {
  location(): number;
  target(): number;
  isRunning(): boolean;
  hold(): void;
  beginDrag(): void;
  setTarget(next: number): void;
  release(next: number): void;
  cancelDrag(): void;
  snap(next: number): void;
  repaint(): void;
  destroy(): void;
}

export interface MotionCallbacks {
  render: (position: number) => void;
  onRest: () => void;
  // onError is called after the loop has been stopped if render or the physics throws. It needs to put the UI back
  // into a usable state.
  onError: (error: unknown) => void;
}

// createMotion is adapted from Embla Carousel's animation loop and scroll body, see the license at the top of the
// file. A drag never sets the position directly, it only moves target. Then every fixed timestep we do:
//
//   velocity += (target - location) / duration
//   velocity *= friction
//   location += velocity
//
// This way uneven touch sampling doesn't turn into judder. Dragging uses a loose friction and duration so location
// trails the finger, and letting go swaps in a stiff pair that lands without wobbling. The velocity carries over when
// we swap, which is why a fling doesn't need to be handled on its own.
export function createMotion(callbacks: MotionCallbacks): Motion {
  let location = 0;
  let previous = 0;
  let target = 0;
  let velocity = 0;
  let friction = settleFriction;
  let duration = settleDuration;
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
    return Math.abs(target - location) < restDistance && Math.abs(velocity) < restDistance;
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
      // If the tab stalled then don't try to catch up on all of it at once.
      if (elapsed > timeStep * maxCatchup) {
        elapsed = timeStep * maxCatchup;
      }

      accumulated += elapsed;
      while (accumulated >= timeStep) {
        step();
        accumulated -= timeStep;
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

      // Render somewhere between the last two steps. This keeps it smooth on displays that aren't 60Hz.
      callbacks.render(previous + (location - previous) * (accumulated / timeStep));
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
    friction = settleFriction;
    duration = settleDuration;
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
      friction = dragFriction;
      duration = dragDuration;
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
      friction = settleFriction;
      duration = settleDuration;
      target = next;
      start();
    },

    cancelDrag() {
      dragging = false;
      friction = settleFriction;
      duration = settleDuration;
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
