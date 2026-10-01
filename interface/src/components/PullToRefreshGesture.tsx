import type React from 'react';
import { Fragment, useCallback, useEffect, useRef, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { ArrowDown, Check, RotateCw, X } from 'lucide-react';

import styles from './PullToRefreshGesture.module.scss';

// This is the actual pull to refresh behavior, its lazy loaded by PullToRefresh once the page has finished loading so
// that none of this ends up in the initial bundle. iOS doesnt give installed PWAs a pull to refresh at all so we gotta
// do it ourselves.
//
// On mobile the window is the scroll container, so we only let a pull start when the window is scrolled all the way to
// the top. While the user is dragging we update the DOM directly instead of going through React state, otherwise wed be
// re-rendering the entire app on every single touchmove. React state only gets updated when the status changes.
//
// If theres ever a part of the page that should never start a pull just add data-ptr-ignore to it.

// Status is the current state of the pull, it goes idle -> pulling -> armed -> refreshing -> done (or error) -> idle.
// Armed means the user has pulled far enough that letting go will trigger a refresh. If they keep pulling past that it
// goes to reloadArmed instead, and letting go reloads the whole page (reloading). Theres no coming back from reloading
// since the page just goes away.
type Status = 'idle' | 'pulling' | 'armed' | 'reloadArmed' | 'refreshing' | 'reloading' | 'done' | 'error';

// Stage is just the part of the status we care about while the finger is still down.
type PullStage = 'pulling' | 'armed' | 'reloadArmed';

// How far (in pixels, after the rubber band is applied) the user needs to pull before letting go will refresh.
const PULL_THRESHOLD = 72;
// Pull even further than this and letting go does a full page reload instead of just refetching. We dont have a service
// worker so this is the only way an installed PWA will pick up a new version of monetr, or recover if loading a chunk
// failed. This works out to about 350px of actual finger movement cuz of the rubber band, which is roughly where the
// old pull to refresh used to reload.
const RELOAD_THRESHOLD = 110;
// The pull can never go further than this, the rubber band slows down as it gets closer to this distance.
const MAX_PULL_DISTANCE = 160;
// How stiff the rubber band feels, higher numbers make it feel lighter.
const PULL_RESISTANCE = 1;
// Keep the spinner up for at least this long, if the refresh is really fast the spinner would just flash otherwise.
const MIN_REFRESH_TIME = 500;
// How long we show the check mark or the error X before we slide everything back up.
const DONE_DISPLAY_TIME = 650;
const ERROR_DISPLAY_TIME = 1400;
// How long it takes to slide everything back up. This needs to match --ptr-dur in PullToRefresh.module.scss.
const RETRACT_TIME = 380;

// The progress ring is an SVG circle, we draw it in by changing the stroke dash offset. So we need the circumference to
// know how much of the circle to show.
const RING_RADIUS = 14;
const RING_CIRCUMFERENCE = 2 * Math.PI * RING_RADIUS;
// Roughly how tall the indicator is (the disc, the label and the padding around them), used to center it in the gap
// that the pull opens up.
const INDICATOR_HEIGHT = 74;

const statusLabels: Record<Exclude<Status, 'idle'>, string> = {
  pulling: 'Pull to refresh',
  armed: 'Release to refresh',
  reloadArmed: 'Release to reload',
  refreshing: 'Refreshing',
  reloading: 'Reloading',
  done: 'Updated',
  error: 'Refresh failed',
};

// The indicator itself is hidden from screen readers since its purely visual, instead we just announce these.
const screenReaderAnnouncements: Partial<Record<Status, string>> = {
  refreshing: 'Refreshing',
  reloading: 'Reloading',
  done: 'Content updated',
  error: 'Refresh failed',
};

function sleep(milliseconds: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, milliseconds));
}

// vibrate will give the user a little bit of haptic feedback. This only works on Android, iOS doesnt support the
// vibration API at all so this just does nothing there.
function vibrate(pattern: number | number[]) {
  try {
    navigator.vibrate?.(pattern);
  } catch {
    // Some browsers throw instead of just not having the function, dont really care either way.
  }
}

// rubberBand takes how far the user has actually dragged their finger and returns how far the content should move. At
// the start this is basically 1:1, but the further they pull the less the content moves, and itll never go past the
// limit. This is the same feel as the iOS overscroll bounce.
function rubberBand(distance: number, limit: number, stiffness: number): number {
  if (distance <= 0) {
    return 0;
  }

  return limit * (1 - 1 / ((distance * stiffness) / limit + 1));
}

// isBlockedFromPulling will walk up from the element the user touched to the root. If anything along the way has opted
// out with data-ptr-ignore, or is its own scroll container that isnt scrolled to the top, then we shouldnt start a
// pull. Otherwise trying to scroll back up inside something like a dropdown would refresh the page instead.
function isBlockedFromPulling(target: EventTarget | null, root: HTMLElement | null): boolean {
  let element = target as HTMLElement | null;
  while (element && element !== root) {
    if (element.hasAttribute?.('data-ptr-ignore')) {
      return true;
    }

    if (element.scrollTop > 0) {
      const overflow = getComputedStyle(element).overflowY;
      if (overflow === 'auto' || overflow === 'scroll') {
        return true;
      }
    }

    element = element.parentElement;
  }

  return false;
}

export interface PullToRefreshGestureProps {
  // rootRef is the element we listen for touches on and that gets moved down during a pull, it is owned by
  // PullToRefresh.
  rootRef: React.RefObject<HTMLDivElement | null>;
}

export default function PullToRefreshGesture(props: PullToRefreshGestureProps): React.JSX.Element {
  const { rootRef } = props;
  const queryClient = useQueryClient();

  const indicatorRef = useRef<HTMLDivElement>(null);
  const backdropRef = useRef<HTMLDivElement>(null);
  const indicatorContentRef = useRef<HTMLDivElement>(null);
  const ringArcRef = useRef<SVGCircleElement>(null);
  const ringRotationRef = useRef<SVGGElement>(null);

  const [status, setStatus] = useState<Status>('idle');
  // We keep a copy of the status in a ref too, the touch handlers are only registered once so they would only ever see
  // the very first value of the state.
  const statusRef = useRef<Status>('idle');
  const isMountedRef = useRef(true);
  const isRefreshingRef = useRef(false);
  // Everything about the touch that is currently happening. This lives in a ref because it changes on every touchmove
  // and we dont want any of that to cause a re-render.
  const touchRef = useRef({
    // waiting means the finger is down but we dont know yet if they are pulling down or scrolling.
    phase: 'none' as 'none' | 'waiting' | 'pulling',
    touchId: -1,
    startX: 0,
    startY: 0,
    // How far the content is pulled down right now, after the rubber band.
    distance: 0,
    stage: 'pulling' as PullStage,
    // The requestAnimationFrame handle for the next draw, 0 when nothing is scheduled.
    frame: 0,
    // The element the finger first touched, we listen on it directly too. See the comment at the bottom of the effect.
    target: null as EventTarget | null,
  });

  useEffect(() => {
    isMountedRef.current = true;
    return () => {
      isMountedRef.current = false;
    };
  }, []);

  const updateStatus = useCallback((next: Status) => {
    if (statusRef.current === next) {
      return;
    }

    statusRef.current = next;
    setStatus(next);
  }, []);

  // drawPull moves the page content and the indicator to match how far the user has pulled. When animate is true the
  // browser will ease into the new position (like when we let go), otherwise it snaps there immediately which is what
  // we want while the finger is still dragging.
  const drawPull = useCallback(
    (distance: number, animate: boolean) => {
      const content = rootRef.current;
      const indicator = indicatorRef.current;
      const backdrop = backdropRef.current;
      const indicatorContent = indicatorContentRef.current;
      const ringArc = ringArcRef.current;
      const ringRotation = ringRotationRef.current;
      if (!content || !indicator || !backdrop || !indicatorContent || !ringArc || !ringRotation) {
        return;
      }

      // 0 when nothing has been pulled, 1 once they have pulled far enough to refresh.
      const progress = Math.min(distance / PULL_THRESHOLD, 1);

      // We move the content with top instead of a transform on purpose. A transform would make anything inside with
      // position: fixed (like the top navigation) move along with the content instead of staying where it is. We also
      // always set it to a pixel value, even 0px. If we cleared it back to auto the browser cant animate that and the
      // page would just snap back up.
      content.style.transition = animate ? 'top var(--ptr-dur) var(--ptr-ease)' : 'none';
      content.style.top = `${distance}px`;

      // Keep the indicator centered in the gap that the pull opens up, and grow it a bit as they pull.
      const indicatorOffset = Math.round(distance / 2 - INDICATOR_HEIGHT / 2);
      const indicatorScale = 0.7 + 0.3 * progress;
      indicator.style.transition = animate ? 'transform var(--ptr-dur) var(--ptr-ease)' : 'none';
      indicator.style.transform = `translate3d(0,${indicatorOffset}px,0) scale(${indicatorScale})`;

      // Fade it in as they pull. The backdrop and the content get faded separately instead of just fading the whole
      // indicator, cuz browsers turn off backdrop-filter for anything inside of an element thats not fully opaque. So
      // if we faded the whole indicator the blur wouldnt show up until it was done fading in.
      const opacity = distance > 0 ? String(Math.min(1, progress * 1.4)) : '0';
      for (const element of [backdrop, indicatorContent]) {
        element.style.transition = animate ? 'opacity var(--ptr-dur) var(--ptr-ease)' : 'none';
        element.style.opacity = opacity;
      }

      // Draw in the ring as they pull, and spin it a little at the same time so it looks like its winding up.
      ringArc.style.strokeDashoffset = String(RING_CIRCUMFERENCE * (1 - progress));
      ringRotation.setAttribute('transform', `rotate(${Math.round(progress * 200)} 18 18)`);
    },
    [rootRef],
  );

  // reload does a full page reload. We still show the spinner so its obvious something is happening while the browser
  // goes and gets the page again.
  const reload = useCallback(() => {
    isRefreshingRef.current = true;
    updateStatus('reloading');
    drawPull(PULL_THRESHOLD, true);
    if (ringArcRef.current) {
      ringArcRef.current.style.strokeDashoffset = String(RING_CIRCUMFERENCE * 0.72);
    }
    window.location.reload();
  }, [drawPull, updateStatus]);

  // refresh refetches everything that is currently on the screen. It walks the indicator through the refreshing, done
  // (or error) and back to idle states on its own, so all the caller needs to do is kick it off.
  const refresh = useCallback(async () => {
    if (isRefreshingRef.current) {
      return;
    }
    isRefreshingRef.current = true;

    updateStatus('refreshing');
    drawPull(PULL_THRESHOLD, true);
    // Leave a gap in the ring while it is spinning so you can actually tell its spinning.
    if (ringArcRef.current) {
      ringArcRef.current.style.strokeDashoffset = String(RING_CIRCUMFERENCE * 0.72);
    }

    const startedAt = performance.now();
    let succeeded = true;
    try {
      await queryClient.refetchQueries({ type: 'active' }, { throwOnError: true });
    } catch (error) {
      succeeded = false;
      console.error('failed to refresh', error);
    }

    const timeLeft = MIN_REFRESH_TIME - (performance.now() - startedAt);
    if (timeLeft > 0) {
      await sleep(timeLeft);
    }
    // Every time we wait we gotta make sure were still mounted before touching anything again.
    if (!isMountedRef.current) {
      return;
    }

    if (ringArcRef.current) {
      ringArcRef.current.style.strokeDashoffset = '0';
    }
    updateStatus(succeeded ? 'done' : 'error');
    vibrate(succeeded ? 8 : [12, 60, 12]);
    await sleep(succeeded ? DONE_DISPLAY_TIME : ERROR_DISPLAY_TIME);
    if (!isMountedRef.current) {
      return;
    }

    drawPull(0, true);
    await sleep(RETRACT_TIME);
    if (!isMountedRef.current) {
      return;
    }

    isRefreshingRef.current = false;
    updateStatus('idle');
  }, [drawPull, updateStatus, queryClient]);

  useEffect(() => {
    const root = rootRef.current;
    if (!root) {
      return;
    }
    const touch = touchRef.current;

    // Touch events can fire way more often than the screen actually refreshes, so instead of drawing on every touchmove
    // we only draw once per animation frame.
    function scheduleDraw() {
      if (touch.frame) {
        return;
      }

      touch.frame = requestAnimationFrame(() => {
        touch.frame = 0;
        drawPull(touch.distance, false);
      });
    }

    function cancelScheduledDraw() {
      if (touch.frame) {
        cancelAnimationFrame(touch.frame);
      }
      touch.frame = 0;
    }

    function resetTouch() {
      cancelScheduledDraw();
      stopListeningToTarget();
      touch.phase = 'none';
      touch.touchId = -1;
    }

    // Find the finger that started the pull, if they put a second finger down we just ignore it whatever.
    function findOurTouch(touches: TouchList): Touch | undefined {
      return Array.from(touches).find(item => item.identifier === touch.touchId);
    }

    function handleTouchStart(event: TouchEvent) {
      const firstTouch = event.touches[0];
      if (event.touches.length !== 1 || !firstTouch) {
        return;
      }
      if (isRefreshingRef.current) {
        return;
      }
      // If theres still a pull hanging around from a previous touch then something went wrong and we never got the
      // touchend for it. Clean it up here instead of just bailing, otherwise pull to refresh would be stuck until the
      // page is reloaded.
      if (touch.phase !== 'none') {
        const wasPulling = touch.phase === 'pulling';
        resetTouch();
        if (wasPulling) {
          drawPull(0, true);
          updateStatus('idle');
        }
      }
      // If there is a dialog open then do nothing, pulling on a dialog should never refresh the page behind it.
      if (document.querySelector('[role="dialog"]')) {
        return;
      }
      // Same thing for the mobile sidebar, when its open pulling anywhere should never refresh the page.
      if (document.querySelector('#root.sidebar-open')) {
        return;
      }
      if (window.scrollY > 0 || isBlockedFromPulling(event.target, root)) {
        return;
      }

      // We dont know yet if this is a pull or just a normal scroll, so we wait for the first touchmove to decide.
      touch.phase = 'waiting';
      touch.touchId = firstTouch.identifier;
      touch.startX = firstTouch.clientX;
      touch.startY = firstTouch.clientY;
      touch.distance = 0;
      touch.stage = 'pulling';
      listenToTarget(event.target);
    }

    function handleTouchMove(event: TouchEvent) {
      if (touch.phase === 'none') {
        return;
      }
      const ourTouch = findOurTouch(event.touches);
      if (!ourTouch) {
        return;
      }

      const movedX = ourTouch.clientX - touch.startX;
      const movedY = ourTouch.clientY - touch.startY;

      if (touch.phase === 'waiting') {
        if (movedX === 0 && movedY === 0) {
          return;
        }

        // We need to decide on the very first move if this is a pull or not. Once iOS starts scrolling natively it wont
        // let us cancel the touchmove anymore, so if we waited for them to move a few pixels first the native bounce
        // would win. Moving mostly sideways or upwards means they arent pulling, so let the browser have it. scrollY
        // can also go negative on iOS while the page is bouncing, which is why this is <= and not ===.
        const isPullingDown = movedY > 0 && movedY >= Math.abs(movedX) && window.scrollY <= 0;
        // If the touchmove cant be canceled then the browser is already scrolling on its own, like when they touch the
        // screen while a fast fling is still going. We cant stop the native bounce at that point anyway so just let the
        // browser have it instead of fighting it. This is also why we dont turn off overscroll for the whole page, the
        // native bounce still works for flings and at the bottom of the page, and calling preventDefault on the first
        // move is enough to stop the bounce (and the browsers own pull to refresh) when they are actually pulling.
        if (!isPullingDown || !event.cancelable) {
          resetTouch();
          return;
        }

        touch.phase = 'pulling';
        updateStatus('pulling');
      }

      // Stop the browser from scrolling the page while were pulling.
      if (event.cancelable) {
        event.preventDefault();
      }

      // If they drag back up past where they started then they changed their mind, put everything back and let them
      // scroll normally.
      if (movedY < 0) {
        touch.distance = 0;
        drawPull(0, false);
        resetTouch();
        updateStatus('idle');
        return;
      }

      touch.distance = rubberBand(movedY, MAX_PULL_DISTANCE, PULL_RESISTANCE);

      // Only update the status when we cross one of the thresholds, not on every move.
      let stage: PullStage = 'pulling';
      if (touch.distance >= RELOAD_THRESHOLD) {
        stage = 'reloadArmed';
      } else if (touch.distance >= PULL_THRESHOLD) {
        stage = 'armed';
      }
      if (stage !== touch.stage) {
        touch.stage = stage;
        updateStatus(stage);
        if (stage !== 'pulling') {
          vibrate(10);
        }
      }

      scheduleDraw();
    }

    function handleTouchEnd(event: TouchEvent) {
      if (touch.phase === 'none') {
        return;
      }
      // A different finger was lifted, we only care about the one that started the pull.
      if (!findOurTouch(event.changedTouches)) {
        return;
      }
      if (touch.phase !== 'pulling') {
        resetTouch();
        return;
      }

      const pulledDistance = touch.distance;
      resetTouch();
      if (pulledDistance >= RELOAD_THRESHOLD) {
        reload();
      } else if (pulledDistance >= PULL_THRESHOLD) {
        refresh();
      } else {
        // Didnt pull far enough, just slide everything back up.
        drawPull(0, true);
        updateStatus('idle');
      }
    }

    // The browser can cancel a touch on its own (like when a system gesture takes over), when that happens we just
    // put everything back without refreshing.
    function handleTouchCancel(event: TouchEvent) {
      if (touch.phase === 'none' || !findOurTouch(event.changedTouches)) {
        return;
      }
      const wasPulling = touch.phase === 'pulling';
      resetTouch();
      if (wasPulling) {
        drawPull(0, true);
        updateStatus('idle');
      }
    }

    // Touch events are always sent to the element the finger first touched, even if React removes that element from
    // the page mid pull (like a skeleton getting swapped out for real data). Once its removed the events dont bubble
    // up to the window anymore, so we would never find out the finger was lifted and the page would be stuck pulled
    // down. To get around that we also listen on the touched element itself while a pull is happening. These only do
    // anything once the element has been removed, while its still on the page the window listeners handle everything
    // and we dont want to handle each event twice.
    function handleDetachedTouchMove(event: TouchEvent) {
      if (!(event.currentTarget as Node).isConnected) {
        handleTouchMove(event);
      }
    }

    function handleDetachedTouchEnd(event: TouchEvent) {
      if (!(event.currentTarget as Node).isConnected) {
        handleTouchEnd(event);
      }
    }

    function handleDetachedTouchCancel(event: TouchEvent) {
      if (!(event.currentTarget as Node).isConnected) {
        handleTouchCancel(event);
      }
    }

    function listenToTarget(target: EventTarget | null) {
      stopListeningToTarget();
      if (!target) {
        return;
      }
      touch.target = target;
      target.addEventListener('touchmove', handleDetachedTouchMove as EventListener, { passive: false });
      target.addEventListener('touchend', handleDetachedTouchEnd as EventListener);
      target.addEventListener('touchcancel', handleDetachedTouchCancel as EventListener);
    }

    function stopListeningToTarget() {
      if (!touch.target) {
        return;
      }
      touch.target.removeEventListener('touchmove', handleDetachedTouchMove as EventListener);
      touch.target.removeEventListener('touchend', handleDetachedTouchEnd as EventListener);
      touch.target.removeEventListener('touchcancel', handleDetachedTouchCancel as EventListener);
      touch.target = null;
    }

    // Only touchstart goes on the root, thats how we know the pull started inside the app. The rest go on the window so
    // that the browser knows up front that we might call preventDefault on touchmove.
    // touchmove has to be passive: false, otherwise preventDefault does nothing and the page scrolls while we pull.
    root.addEventListener('touchstart', handleTouchStart, { passive: true });
    window.addEventListener('touchmove', handleTouchMove, { passive: false });
    window.addEventListener('touchend', handleTouchEnd);
    window.addEventListener('touchcancel', handleTouchCancel);
    return () => {
      root.removeEventListener('touchstart', handleTouchStart);
      window.removeEventListener('touchmove', handleTouchMove);
      window.removeEventListener('touchend', handleTouchEnd);
      window.removeEventListener('touchcancel', handleTouchCancel);
      stopListeningToTarget();
      cancelScheduledDraw();
    };
  }, [drawPull, refresh, reload, updateStatus, rootRef]);

  // Keep showing the last label while we go back to idle, otherwise the text would disappear while the indicator is
  // still sliding back up.
  const lastLabelRef = useRef(statusLabels.pulling);
  if (status !== 'idle') {
    lastLabelRef.current = statusLabels[status];
  }

  return (
    <Fragment>
      <div aria-hidden='true' className={styles.indicator} data-status={status} ref={indicatorRef}>
        <div className={styles.backdrop} ref={backdropRef} />
        <div className={styles.indicatorContent} ref={indicatorContentRef}>
          <div className={styles.disc}>
            <svg aria-hidden='true' className={styles.ring} height='36' viewBox='0 0 36 36' width='36'>
              <circle className={styles.track} cx='18' cy='18' r={RING_RADIUS} />
              <g ref={ringRotationRef}>
                {/* Rotated -90 degrees so the ring starts drawing from the top instead of from the right. */}
                <circle
                  className={styles.arc}
                  cx='18'
                  cy='18'
                  r={RING_RADIUS}
                  ref={ringArcRef}
                  strokeDasharray={RING_CIRCUMFERENCE}
                  strokeDashoffset={RING_CIRCUMFERENCE}
                  transform='rotate(-90 18 18)'
                />
              </g>
            </svg>
            <ArrowDown className={`${styles.icon} ${styles.arrow}`} />
            <RotateCw className={`${styles.icon} ${styles.reload}`} />
            <Check className={`${styles.icon} ${styles.check}`} />
            <X className={`${styles.icon} ${styles.cross}`} />
          </div>
          <span className={styles.label}>{lastLabelRef.current}</span>
        </div>
      </div>

      <span aria-live='polite' className={styles.screenReaderOnly} role='status'>
        {screenReaderAnnouncements[status] ?? ''}
      </span>
    </Fragment>
  );
}
