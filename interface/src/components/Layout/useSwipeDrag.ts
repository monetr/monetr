import { useCallback, useEffect, useRef } from 'react';

import {
  type Axis,
  commitThreshold,
  type DragMode,
  EDGE_ZONE,
  FLING,
  type Motion,
  resolveAxis,
  rubber,
  type Sample,
  swipeBlocked,
  tick,
  velocityOf,
} from '@monetr/interface/components/Layout/swipeMotion';
import { setSidebarGestureActive } from '@monetr/interface/util/sidebarGesture';

// One drag at a time, in a single object, so every way out of a drag (pointerup, pointercancel, a route change, or
// unmounting) goes through the same teardown.
export interface Drag {
  pointerId: number;
  // Kept because React nulls out currentTarget once the handler returns, and pointer capture isn't taken until the
  // axis resolves.
  el: Element;
  captured: boolean;
  mode: DragMode;
  x0: number;
  y0: number;
  p0: number;
  axis: Axis | null;
  selectionLocked: boolean;
  samples: Array<Sample>;
}

export interface SwipeDragOptions {
  isMobile: boolean;
  mobileRef: React.RefObject<boolean>;
  motion: Motion;
  widthRef: React.RefObject<number>;
  // The routed page. Used as the stopping point when checking whether something else owns a horizontal drag.
  pageRef: React.RefObject<HTMLDivElement | null>;
  // The committed open state, read from pointer handlers without changing their identity.
  isOpenRef: React.RefObject<boolean>;
  lift(on: boolean): void;
  // Called when a drag starts from closed, so focus can be restored after the drawer closes again.
  rememberFocus(): void;
  onCommit(open: boolean): void;
}

export interface SwipeDrag {
  gesture: {
    onClickCapture: (event: React.MouseEvent<HTMLElement>) => void;
    onPointerCancel: (event: React.PointerEvent<HTMLElement>) => void;
    onPointerDown: (event: React.PointerEvent<HTMLElement>) => void;
    onPointerMove: (event: React.PointerEvent<HTMLElement>) => void;
    onPointerUp: (event: React.PointerEvent<HTMLElement>) => void;
  };
  // The one teardown path, shared by pointerup, pointercancel, navigation, errors and unmount. Returns the drag it
  // ended so the caller can decide what happens next.
  releaseDrag(): Drag | null;
}

// Pointer input never writes a style. A pointermove only moves the integrator's target; the integrator moves the
// position on its own clock.
export default function useSwipeDrag(options: SwipeDragOptions): SwipeDrag {
  const { isMobile, mobileRef, motion, widthRef, pageRef, isOpenRef, lift, rememberFocus, onCommit } = options;

  const dragRef = useRef<Drag | null>(null);
  const suppressClickRef = useRef(false);

  const releaseDrag = useCallback((): Drag | null => {
    const drag = dragRef.current;
    dragRef.current = null;
    setSidebarGestureActive(false);
    if (!drag) {
      return null;
    }
    if (drag.selectionLocked) {
      document.body.style.userSelect = '';
    }
    if (drag.captured && drag.el.hasPointerCapture?.(drag.pointerId)) {
      drag.el.releasePointerCapture(drag.pointerId);
    }
    return drag;
  }, []);

  // Where every press that never turned into a horizontal drag ends up. Settles back to the committed state rather
  // than the nearest end: a press that never became a drag isn't a decision, and tapping mid animation should let the
  // animation finish. cancelDrag alone isn't enough, at rest it stops the loop without rendering a frame, which would
  // leave a panel that lift() made visible stranded on screen.
  const settleFromTap = useCallback(() => {
    motion.release(isOpenRef.current ? 1 : 0);
  }, [motion, isOpenRef]);

  const onPointerDown = useCallback(
    (event: React.PointerEvent<HTMLElement>) => {
      if (!mobileRef.current || !event.isPrimary || event.button > 0 || dragRef.current) {
        return;
      }
      // Portalled content (the account picker, the link menu) propagates through the React tree without being a DOM
      // descendant, and must not arm a drag on the sidebar underneath it. Clear the suppression first so a tap on
      // portalled content doesn't leave a stale one armed.
      suppressClickRef.current = false;
      if (!event.currentTarget.contains(event.target as Node)) {
        return;
      }

      const position = motion.location();
      let mode: DragMode = 'loose';
      if (position < 0.001 && event.clientX > EDGE_ZONE) {
        // Out in the page the gesture is still armed, it just has to prove itself with more slop, rightward only, and
        // inside the cone. It must not take the axis off something that already owns it.
        if (swipeBlocked(event.target, pageRef.current)) {
          return;
        }
        mode = 'page';
      }

      if (position < 0.001) {
        rememberFocus();
      }

      // Freeze any in flight animation and promote layers now, while the finger is still sitting still. Promoting on
      // first movement puts the raster cost exactly where the eye is looking.
      motion.hold();
      lift(true);
      dragRef.current = {
        pointerId: event.pointerId,
        el: event.currentTarget,
        captured: false,
        mode,
        x0: event.clientX,
        y0: event.clientY,
        p0: position,
        axis: null,
        selectionLocked: false,
        samples: [{ x: event.clientX, t: event.timeStamp }],
      };
    },
    [motion, lift, mobileRef, pageRef, rememberFocus],
  );

  const onPointerMove = useCallback(
    (event: React.PointerEvent<HTMLElement>) => {
      const drag = dragRef.current;
      if (!drag || event.pointerId !== drag.pointerId) {
        return;
      }

      const dx = event.clientX - drag.x0;
      const dy = event.clientY - drag.y0;

      if (!drag.axis) {
        const axis = resolveAxis(drag.mode, dx, dy);
        if (!axis) {
          return;
        }
        drag.axis = axis;

        if (axis === 'y') {
          // Not ours. Hand it back before anything has moved or any default has been prevented.
          releaseDrag();
          settleFromTap();
          return;
        }

        suppressClickRef.current = true;
        setSidebarGestureActive(true);
        // Capture here rather than on pointerdown, so ordinary taps and nested controls in the panel never get
        // retargeted out from under themselves.
        drag.el.setPointerCapture?.(drag.pointerId);
        drag.captured = true;
        if (event.pointerType !== 'touch') {
          // Touch never drag selects, and writing body.style forces a document wide style recalc mid gesture.
          drag.selectionLocked = true;
          document.body.style.userSelect = 'none';
        }
        motion.beginDrag();
      }

      // Coalesced events are the raw hardware samples behind this event, so a fresher position and a better velocity
      // estimate. The optional call matters: it only shipped in Safari 18 and browserslist still covers iOS 15-17.
      const coalesced = event.nativeEvent.getCoalescedEvents?.();
      if (coalesced && coalesced.length > 0) {
        for (const sample of coalesced) {
          drag.samples.push({ x: sample.clientX, t: sample.timeStamp });
        }
      } else {
        drag.samples.push({ x: event.clientX, t: event.timeStamp });
      }
      while (drag.samples.length > 16) {
        drag.samples.shift();
      }

      // Move the target and nothing else, no style writes.
      const width = widthRef.current;
      let px = drag.p0 * width + dx;
      if (px > width) {
        px = width + rubber(px - width, width);
      } else if (px < 0) {
        px = -rubber(-px, width);
      }
      motion.setTarget(px / width);

      if (event.pointerType !== 'touch' && event.cancelable) {
        event.preventDefault();
      }
    },
    [motion, widthRef, releaseDrag, settleFromTap],
  );

  const onPointerUp = useCallback(
    (event: React.PointerEvent<HTMLElement>) => {
      if (!dragRef.current || event.pointerId !== dragRef.current.pointerId) {
        return;
      }
      const drag = releaseDrag();
      if (!drag) {
        return;
      }
      if (drag.axis !== 'x') {
        settleFromTap();
        return;
      }

      // Decide from the target, which is where the finger actually is; the physics position trails it on purpose.
      const position = motion.target();
      const velocity = velocityOf(drag.samples);
      const threshold = commitThreshold(widthRef.current);
      const next = Math.abs(velocity) > FLING ? (velocity > 0 ? 1 : 0) : position > threshold ? 1 : 0;

      motion.release(next);
      if (next !== (drag.p0 > 0.5 ? 1 : 0)) {
        tick();
      }
      onCommit(next === 1);
    },
    [motion, widthRef, releaseDrag, settleFromTap, onCommit],
  );

  const onPointerCancel = useCallback(
    (event: React.PointerEvent<HTMLElement>) => {
      if (!dragRef.current || event.pointerId !== dragRef.current.pointerId) {
        return;
      }
      const drag = releaseDrag();
      if (!drag) {
        return;
      }
      if (drag.axis !== 'x') {
        settleFromTap();
        return;
      }

      // iOS fires this when the system takes the gesture, like a back swipe or a notification pull. Settle to the
      // nearest end instead of honoring a fling the user might not have meant.
      const next = motion.target() > 0.5 ? 1 : 0;
      motion.release(next);
      onCommit(next === 1);
    },
    [motion, releaseDrag, settleFromTap, onCommit],
  );

  const onClickCapture = useCallback((event: React.MouseEvent<HTMLElement>) => {
    if (!suppressClickRef.current) {
      return;
    }
    // Keyboard and assistive tech activations arrive with detail 0 and never came from the pointer being suppressed.
    // A touch drag preventDefaults the touchmove so no click is synthesised and the flag survives the gesture.
    if (event.detail === 0) {
      return;
    }
    suppressClickRef.current = false;
    // preventDefault is the half that matters. wouter's Link renders a real <a href> and calls preventDefault inside
    // its own onClick; only stopping propagation would let the browser do a native navigation, a full reload.
    event.preventDefault();
    event.stopPropagation();
  }, []);

  // The listener that actually stops the page scrolling. React attaches touchmove passively at the root, so
  // preventDefault inside an onTouchMove prop silently does nothing.
  useEffect(() => {
    if (!isMobile) {
      return;
    }
    const page = pageRef.current;
    if (!page) {
      return;
    }

    const onTouchMove = (event: TouchEvent) => {
      const drag = dragRef.current;
      if (!drag || !event.cancelable) {
        return;
      }
      if (drag.axis === 'x') {
        event.preventDefault();
        return;
      }
      if (drag.axis === null && event.touches.length === 1) {
        const touch = event.touches[0];
        // Same rule pointermove uses, so the two can't disagree about whose gesture this is.
        if (touch && resolveAxis(drag.mode, touch.clientX - drag.x0, touch.clientY - drag.y0) === 'x') {
          event.preventDefault();
        }
      }
    };

    page.addEventListener('touchmove', onTouchMove, { passive: false });
    return () => page.removeEventListener('touchmove', onTouchMove);
  }, [isMobile, pageRef]);

  return {
    gesture: { onClickCapture, onPointerCancel, onPointerDown, onPointerMove, onPointerUp },
    releaseDrag,
  };
}
