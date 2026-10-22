import { useCallback, useEffect, useRef } from 'react';

import {
  type Axis,
  commitThreshold,
  type DragMode,
  edgeZone,
  flingVelocity,
  type Motion,
  resolveAxis,
  rubber,
  type Sample,
  swipeBlocked,
  tick,
  velocityOf,
} from '@monetr/interface/components/Layout/swipeMotion';
import { setSidebarGestureActive } from '@monetr/interface/util/sidebarGesture';

export interface Drag {
  pointerId: number;
  // React nulls out currentTarget once the handler returns, and we don't capture the pointer until we know the axis.
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
  pageRef: React.RefObject<HTMLDivElement | null>;
  isOpenRef: React.RefObject<boolean>;
  lift(on: boolean): void;
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
  releaseDrag(): Drag | null;
}

export default function useSwipeDrag(options: SwipeDragOptions): SwipeDrag {
  const { isMobile, mobileRef, motion, widthRef, pageRef, isOpenRef, lift, rememberFocus, onCommit } = options;

  const dragRef = useRef<Drag | null>(null);
  const suppressClickRef = useRef(false);

  // releaseDrag is the one place a drag gets torn down, whether that is from the pointer being lifted, a route change,
  // an error or unmounting. It returns the drag it ended.
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

  // A press that never turned into a horizontal drag goes back to the state we were already in, not the nearest
  // end. If they tap in the middle of an animation then the animation should just finish. We can't use cancelDrag here,
  // at rest it stops the loop without rendering, and a panel that lift() made visible would be stuck on screen.
  const settleFromTap = useCallback(() => {
    motion.release(isOpenRef.current ? 1 : 0);
  }, [motion, isOpenRef]);

  const onPointerDown = useCallback(
    (event: React.PointerEvent<HTMLElement>) => {
      if (!mobileRef.current || !event.isPrimary || event.button > 0 || dragRef.current) {
        return;
      }
      // Reset this before the portal check below. A tap on portalled content would leave a stale one if we didn't.
      suppressClickRef.current = false;
      // Portalled content (like the account picker) still bubbles through the React tree even though it is not inside
      // the panel in the DOM. Those should never start a drag.
      if (!event.currentTarget.contains(event.target as Node)) {
        return;
      }

      const position = motion.location();
      let mode: DragMode = 'loose';
      if (position < 0.001 && event.clientX > edgeZone) {
        if (swipeBlocked(event.target, pageRef.current)) {
          return;
        }
        mode = 'page';
      }

      if (position < 0.001) {
        rememberFocus();
      }

      // Promote the layers now while the finger is still. If we waited for the first move then the raster cost would
      // land right when the user is watching it.
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
        samples: [
          {
            x: event.clientX,
            t: event.timeStamp,
          },
        ],
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
          releaseDrag();
          settleFromTap();
          return;
        }

        suppressClickRef.current = true;
        setSidebarGestureActive(true);
        // Capture here instead of on pointerdown. If we captured on pointerdown then normal taps on things in the panel
        // would get retargeted.
        drag.el.setPointerCapture?.(drag.pointerId);
        drag.captured = true;
        // Touch never selects text when dragging, and writing to body.style makes the entire document recalculate.
        if (event.pointerType !== 'touch') {
          drag.selectionLocked = true;
          document.body.style.userSelect = 'none';
        }
        motion.beginDrag();
      }

      // Coalesced events give us every sample behind this event, which makes the velocity more accurate. Safari only
      // added it in 18 and we still support iOS 15 to 17.
      const coalesced = event.nativeEvent.getCoalescedEvents?.();
      if (coalesced && coalesced.length > 0) {
        drag.samples.push(
          ...coalesced.map(item => ({
            x: item.clientX,
            t: item.timeStamp,
          })),
        );
      } else {
        drag.samples.push({
          x: event.clientX,
          t: event.timeStamp,
        });
      }
      while (drag.samples.length > 16) {
        drag.samples.shift();
      }

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

      // Use the target here and not the location. The target is where the finger actually is, the location trails it.
      const position = motion.target();
      const velocity = velocityOf(drag.samples);
      const threshold = commitThreshold(widthRef.current);
      const next = Math.abs(velocity) > flingVelocity ? (velocity > 0 ? 1 : 0) : position > threshold ? 1 : 0;

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

      // iOS cancels the pointer when the system takes over the gesture, like a back swipe. Go to the nearest end
      // instead of trusting a fling they might not have meant.
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
    // Keyboard and screen reader clicks have a detail of 0, let those through. A touch drag never makes a click since
    // we prevent the touchmove, which means the flag can still be set when one of these comes in.
    if (event.detail === 0) {
      return;
    }
    suppressClickRef.current = false;
    // wouter's Link is a real <a href> and only calls preventDefault in the Link's onClick. If we only stopped
    // propagation then the browser would follow the link and reload the page.
    event.preventDefault();
    event.stopPropagation();
  }, []);

  // React attaches touchmove as a passive listener and preventDefault does nothing in there. This is the listener that
  // stops the page from scrolling while dragging.
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
        if (touch && resolveAxis(drag.mode, touch.clientX - drag.x0, touch.clientY - drag.y0) === 'x') {
          event.preventDefault();
        }
      }
    };

    page.addEventListener('touchmove', onTouchMove, { passive: false });
    return () => page.removeEventListener('touchmove', onTouchMove);
  }, [isMobile, pageRef]);

  return {
    gesture: {
      onClickCapture,
      onPointerCancel,
      onPointerDown,
      onPointerMove,
      onPointerUp,
    },
    releaseDrag,
  };
}
