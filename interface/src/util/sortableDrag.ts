// This is the drag side of useSortableList. It is only ever loaded with a dynamic import the first time someone points
// at a sortable item, most page loads never drag anything so there is no reason to ship it up front.

// How far a mouse or pen has to move with the button held before it counts as a drag instead of a click.
const ACTIVATION_DISTANCE = 5;
// Touch has to long press before a drag starts, otherwise you could never scroll the list.
const TOUCH_ACTIVATION_DELAY = 250;
// Moving further than this before the long press finishes means you're scrolling, not dragging.
const TOUCH_TOLERANCE = 8;
const ITEM_SELECTOR = ':scope > [data-sortable-id]';
const EASING = 'transform 200ms cubic-bezier(0.2, 0, 0, 1)';

function transition(): string {
  return window.matchMedia?.('(prefers-reduced-motion: reduce)').matches ? 'none' : EASING;
}

export interface DragStart {
  id: string;
  pointerId: number;
  pointerType: string;
  element: HTMLElement;
  clientX: number;
  clientY: number;
}

export interface DragHost {
  getContainer(): HTMLElement | null;
  onStart(id: string): void;
  // Called with the new order of every item, or null when nothing moved or the drag was cancelled.
  onEnd(ids: Array<string> | null): void;
}

export interface DragController {
  start(event: DragStart): void;
  // Call after the host has rendered the order it got from onEnd, this animates the drop into place.
  settle(): void;
  dispose(): void;
}

interface Settle {
  id: string;
  // Where the dragged element was on screen when it was let go, so the drop can animate from there.
  visualTop: number;
  reordered: boolean;
}

interface Slot {
  id: string;
  element: HTMLElement;
  top: number;
  height: number;
}

interface Session extends DragStart {
  lastY: number;
  active: boolean;
  slots: Array<Slot>;
  fromIndex: number;
  overIndex: number;
  step: number;
  containerTop: number;
  timer: number | null;
  frame: number | null;
}

// A drag ends with a pointerup on the item, which the browser follows with a click. That click would navigate to
// whatever you just dropped, so eat it.
function suppressNextClick(): void {
  const handler = (event: MouseEvent) => {
    event.preventDefault();
    event.stopPropagation();
  };
  window.addEventListener('click', handler, { capture: true, once: true });
  // If no click comes (the pointer ended up somewhere else) make sure we don't eat some later unrelated click.
  window.setTimeout(() => window.removeEventListener('click', handler, { capture: true }), 0);
}

function preventTouchScroll(event: TouchEvent): void {
  if (event.cancelable) {
    event.preventDefault();
  }
}

export function createDragController(host: DragHost): DragController {
  let session: Session | null = null;
  let pendingSettle: Settle | null = null;

  function update(): void {
    const s = session;
    const container = host.getContainer();
    if (!s?.active || !container) {
      return;
    }
    s.frame = null;

    // Account for the list scrolling underneath the pointer while we drag.
    const scrolled = s.containerTop - container.getBoundingClientRect().top;
    const offset = s.lastY - s.clientY + scrolled;
    s.element.style.transform = `translate3d(0, ${offset}px, 0)`;

    const dragged = s.slots[s.fromIndex];
    const first = s.slots[0];
    if (!dragged || !first) {
      return;
    }
    const center = dragged.top + dragged.height / 2 + offset;
    const overIndex = Math.max(0, Math.min(s.slots.length - 1, Math.floor((center - first.top) / s.step)));
    if (overIndex === s.overIndex) {
      return;
    }
    s.overIndex = overIndex;

    // Slide everything between where the item started and where it is now over by one slot to make room.
    s.slots.forEach((slot, index) => {
      if (index === s.fromIndex) {
        return;
      }
      let shift = 0;
      if (s.fromIndex < overIndex && index > s.fromIndex && index <= overIndex) {
        shift = -s.step;
      } else if (s.fromIndex > overIndex && index >= overIndex && index < s.fromIndex) {
        shift = s.step;
      }
      slot.element.style.transform = shift ? `translate3d(0, ${shift}px, 0)` : '';
    });
  }

  function activate(): void {
    const s = session;
    const container = host.getContainer();
    if (!s || !container || s.active) {
      return;
    }
    if (s.timer !== null) {
      window.clearTimeout(s.timer);
      s.timer = null;
    }

    // Measure everything once up front, the rest of the drag just does math against these.
    const containerTop = container.getBoundingClientRect().top;
    const slots = Array.from(container.querySelectorAll<HTMLElement>(ITEM_SELECTOR)).map(element => {
      const rect = element.getBoundingClientRect();
      return { id: element.dataset.sortableId ?? '', element, top: rect.top - containerTop, height: rect.height };
    });
    const fromIndex = slots.findIndex(slot => slot.id === s.id);
    const dragged = slots[fromIndex];
    if (!dragged) {
      finish(false);
      return;
    }

    s.active = true;
    s.slots = slots;
    s.fromIndex = fromIndex;
    s.overIndex = fromIndex;
    const second = slots[1];
    s.step = second && slots[0] ? second.top - slots[0].top : dragged.height;
    s.containerTop = containerTop;

    const easing = transition();
    for (const slot of slots) {
      slot.element.style.transition = slot.id === s.id ? 'none' : easing;
    }
    try {
      s.element.setPointerCapture(s.pointerId);
    } catch {
      // The pointer can already be gone by the time a long press fires, the pointerup handler will clean up.
    }
    window.addEventListener('touchmove', preventTouchScroll, { passive: false });
    host.onStart(s.id);
  }

  function finish(shouldCommit: boolean): void {
    const s = session;
    if (!s) {
      return;
    }
    if (s.active && s.frame !== null) {
      cancelAnimationFrame(s.frame);
      update();
    }
    session = null;
    window.removeEventListener('pointermove', onPointerMove);
    window.removeEventListener('pointerup', onPointerUp);
    window.removeEventListener('pointercancel', onPointerCancel);
    window.removeEventListener('touchmove', preventTouchScroll);
    if (s.timer !== null) {
      window.clearTimeout(s.timer);
    }
    if (!s.active) {
      return;
    }

    suppressNextClick();
    let ids: Array<string> | null = null;
    if (shouldCommit && s.overIndex !== s.fromIndex) {
      ids = s.slots.map(slot => slot.id);
      const [moved] = ids.splice(s.fromIndex, 1);
      if (moved !== undefined) {
        ids.splice(s.overIndex, 0, moved);
      }
    }
    pendingSettle = { id: s.id, visualTop: s.element.getBoundingClientRect().top, reordered: ids !== null };
    host.onEnd(ids);
  }

  function onPointerMove(event: PointerEvent): void {
    const s = session;
    if (!s || event.pointerId !== s.pointerId) {
      return;
    }
    // A press that started while this module was still loading might have already been let go.
    if (!s.active && s.pointerType !== 'touch' && event.buttons === 0) {
      finish(false);
      return;
    }
    s.lastY = event.clientY;
    if (!s.active) {
      const distance = Math.hypot(event.clientX - s.clientX, event.clientY - s.clientY);
      if (s.pointerType === 'touch') {
        if (distance > TOUCH_TOLERANCE) {
          finish(false);
        }
        return;
      }
      if (distance < ACTIVATION_DISTANCE) {
        return;
      }
      activate();
    }
    if (s.frame === null) {
      s.frame = requestAnimationFrame(update);
    }
  }

  function onPointerUp(event: PointerEvent): void {
    if (session && event.pointerId === session.pointerId) {
      finish(true);
    }
  }

  function onPointerCancel(event: PointerEvent): void {
    if (session && event.pointerId === session.pointerId) {
      finish(false);
    }
  }

  return {
    start(event: DragStart): void {
      if (session) {
        return;
      }
      session = {
        ...event,
        lastY: event.clientY,
        active: false,
        slots: [],
        fromIndex: -1,
        overIndex: -1,
        step: 0,
        containerTop: 0,
        timer: null,
        frame: null,
      };
      window.addEventListener('pointermove', onPointerMove);
      window.addEventListener('pointerup', onPointerUp);
      window.addEventListener('pointercancel', onPointerCancel);
      if (event.pointerType === 'touch') {
        session.timer = window.setTimeout(activate, TOUCH_ACTIVATION_DELAY);
      }
    },
    settle(): void {
      const pending = pendingSettle;
      const container = host.getContainer();
      if (!pending || !container) {
        return;
      }
      pendingSettle = null;

      // Clear out the transforms the drag left behind. The dragged item is now in its new spot in the DOM, so move it
      // back to where it was let go and animate it into place instead of having it jump.
      const easing = transition();
      let dropped: HTMLElement | null = null;
      for (const element of container.querySelectorAll<HTMLElement>(ITEM_SELECTOR)) {
        if (element.dataset.sortableId === pending.id) {
          dropped = element;
          continue;
        }
        // When the order changed everything is already in its final DOM position, so the shifts have to go away
        // instantly. When it didn't, let them ease back.
        element.style.transition = pending.reordered ? 'none' : easing;
        element.style.transform = '';
      }
      if (!dropped) {
        return;
      }

      dropped.style.transition = 'none';
      dropped.style.transform = '';
      const delta = pending.visualTop - dropped.getBoundingClientRect().top;
      if (Math.abs(delta) < 1) {
        return;
      }
      dropped.style.transform = `translate3d(0, ${delta}px, 0)`;
      // Force a layout so the browser picks up the starting position before we turn the transition back on.
      void dropped.offsetHeight;
      dropped.style.transition = easing;
      dropped.style.transform = '';
    },
    dispose(): void {
      finish(false);
    },
  };
}
