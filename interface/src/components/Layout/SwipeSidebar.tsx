import { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react';
import { captureException } from '@sentry/react';
import { useLocation } from 'wouter';

import { MobileSidebarContext } from '@monetr/interface/components/Layout/MobileSidebarContext';
import {
  clamp,
  createMotion,
  LAYER_HOLD,
  type Motion,
  prefersReducedMotion,
  SCRIM_MAX,
} from '@monetr/interface/components/Layout/swipeMotion';
import useDrawerFocus from '@monetr/interface/components/Layout/useDrawerFocus';
import useSwipeDrag from '@monetr/interface/components/Layout/useSwipeDrag';
import useIsMobile from '@monetr/interface/hooks/useIsMobile';

import styles from './SwipeSidebar.module.scss';

// The panel is the bank rail plus zero or one secondary column. Which component fills that column is none of the
// drawer's business; it looks for the [data-sidebar-secondary] marker so a future secondary sidebar gets the right
// width and close behavior without touching anything in here.
type PanelShape = 'rail' | 'wide';

function resolvePanelShape(panel: HTMLElement | null): PanelShape {
  if (!panel) {
    return 'rail';
  }
  return panel.querySelector('[data-sidebar-secondary]') === null ? 'rail' : 'wide';
}

export interface SwipeSidebarProps {
  // The navigation panel. On mobile this is what goes inside the drawer.
  panel: React.ReactNode;
  // The routed page.
  children: React.ReactNode;
}

export default function SwipeSidebar(props: SwipeSidebarProps): React.JSX.Element {
  const isMobile = useIsMobile();
  const [pathname] = useLocation();
  const [isOpen, setIsOpen] = useState(false);

  const panelRef = useRef<HTMLElement>(null);
  const scrimRef = useRef<HTMLDivElement>(null);
  const edgeRef = useRef<HTMLDivElement>(null);
  const pageRef = useRef<HTMLDivElement>(null);
  const catcherRef = useRef<HTMLDivElement>(null);

  const widthRef = useRef(1);
  const shapeRef = useRef<PanelShape>('rail');
  const pendingShapeRef = useRef(false);
  const staggerRef = useRef<Array<HTMLElement>>([]);
  const holdRef = useRef(0);
  const reducedRef = useRef(false);
  // The committed state, readable from handlers without changing their identity on open and close.
  const isOpenRef = useRef(false);
  // Written from an effect, not during render, and checked by everything that writes to the DOM. Above the breakpoint
  // the drawer must not touch a single inline style: the panel goes display: contents and the sidebars it wraps are
  // the actual desktop layout.
  const mobileRef = useRef(false);
  // The only two paint outputs that invalidate anything when they change, cached so paint can skip them most frames.
  const previousRef = useRef<{ offscreen: boolean | null; past: boolean | null }>({ offscreen: null, past: null });

  // The one place anything is written per frame.
  const paint = useCallback((position: number) => {
    if (!mobileRef.current) {
      return;
    }
    const panel = panelRef.current;
    const scrim = scrimRef.current;
    const edge = edgeRef.current;
    const catcher = catcherRef.current;
    if (!panel || !scrim || !edge || !catcher) {
      return;
    }

    const width = widthRef.current;
    const visible = clamp(position, 0, 1);

    // Translate only, and never the page itself. The fixed header, action bar and floating button live inside the
    // page, and a transformed ancestor would become their containing block.
    panel.style.transform = `translate3d(${(visible - 1) * width}px,0,0)`;
    scrim.style.opacity = String(visible * SCRIM_MAX);
    edge.style.transform = `translate3d(${visible * width}px,0,0)`;
    catcher.style.transform = `translate3d(${visible * width}px,0,0)`;
    edge.style.opacity = String(visible);

    const previous = previousRef.current;
    const offscreen = visible <= 0.001;
    if (previous.offscreen !== offscreen) {
      previous.offscreen = offscreen;
      // paint owns visibility. lift() shows the panel early, before the first moving frame, but records it here too
      // so the next resting frame puts it back.
      panel.style.visibility = offscreen ? 'hidden' : 'visible';
      edge.style.visibility = offscreen ? 'hidden' : 'visible';
      // Off the resting position, not a midpoint. The commit point is 0.42, so 0.5 here would leave a window where
      // the page is inert but the catcher isn't listening yet.
      catcher.style.pointerEvents = offscreen ? 'none' : 'auto';
    }

    const past = visible > 0.5;
    if (previous.past !== past) {
      previous.past = past;
      // The toggle icon follows the live position through CSS; through React it would be a re-render a frame.
      document.documentElement.dataset.sidebarPast = past ? 'true' : 'false';
    }

    if (reducedRef.current) {
      return;
    }
    const rows = staggerRef.current;
    for (let i = 0; i < rows.length; i++) {
      const row = rows[i];
      if (!row) {
        continue;
      }
      const progress = clamp((position - i * 0.028) * 1.5, 0, 1.05);
      row.style.opacity = String(clamp(progress, 0, 1));
      row.style.transform = `translate3d(${(1 - progress) * -16}px,0,0)`;
    }
  }, []);

  // Layer promotion, held for a beat after rest so a quick open, close, open doesn't pay for it three times.
  const lift = useCallback((on: boolean) => {
    if (!mobileRef.current) {
      return;
    }
    window.clearTimeout(holdRef.current);

    const apply = () => {
      const transform = on ? 'transform' : 'auto';
      if (panelRef.current) {
        panelRef.current.style.willChange = transform;
      }
      if (edgeRef.current) {
        edgeRef.current.style.willChange = transform;
      }
      if (scrimRef.current) {
        scrimRef.current.style.willChange = on ? 'opacity' : 'auto';
      }
      // Without this, animating the rows re-rasters the whole panel layer every frame.
      for (const row of staggerRef.current) {
        row.style.willChange = on ? 'transform, opacity' : 'auto';
      }
    };

    if (!on) {
      holdRef.current = window.setTimeout(apply, LAYER_HOLD);
      return;
    }

    // Re-query instead of capturing once; the panel contents change with the route and with queries resolving. Runs
    // once per interaction, not per frame.
    staggerRef.current = Array.from(panelRef.current?.querySelectorAll<HTMLElement>('[data-stagger]') ?? []);
    reducedRef.current = prefersReducedMotion();
    // Promote and show before anything moves. Browsers don't raster invisible content, so otherwise the first visible
    // frame pays for the whole panel.
    if (panelRef.current) {
      panelRef.current.style.visibility = 'visible';
    }
    if (edgeRef.current) {
      edgeRef.current.style.visibility = 'visible';
    }
    // Tell paint, or it never writes hidden back and a press that never became a drag leaves the panel showing.
    previousRef.current.offscreen = false;
    apply();
  }, []);

  // Puts every inline style the drawer writes back to whatever the stylesheet says. Needed when crossing above the
  // breakpoint, where a leftover `visibility: hidden` would blank out the desktop sidebars, and to recover when the
  // frame loop throws.
  const reset = useCallback(() => {
    window.clearTimeout(holdRef.current);
    previousRef.current = { offscreen: null, past: null };
    document.documentElement.removeAttribute('data-sidebar-past');
    document.body.style.overflow = '';

    for (const element of [panelRef.current, scrimRef.current, edgeRef.current, catcherRef.current]) {
      if (!element) {
        continue;
      }
      element.style.transform = '';
      element.style.visibility = '';
      element.style.opacity = '';
      element.style.willChange = '';
      element.style.pointerEvents = '';
    }
    for (const row of staggerRef.current) {
      row.style.transform = '';
      row.style.opacity = '';
      row.style.willChange = '';
    }
    staggerRef.current = [];
  }, []);

  // The callbacks go through a ref so the integrator never captures a stale paint, and so the lazy init below stays
  // clear of anything the React compiler would complain about.
  const sinkRef = useRef<{ render: (position: number) => void; onRest: () => void; onError: (error: unknown) => void }>(
    {
      render: () => {},
      onRest: () => {},
      onError: () => {},
    },
  );
  const motionRef = useRef<Motion | null>(null);
  if (motionRef.current === null) {
    motionRef.current = createMotion({
      render: position => sinkRef.current.render(position),
      onRest: () => sinkRef.current.onRest(),
      onError: error => sinkRef.current.onError(error),
    });
  }
  const motion = motionRef.current;

  // CSS owns the actual width; this picks which one and measures the result. Deferred while anything is in flight
  // because a panel that changes size mid animation visibly jumps.
  const applyShape = useCallback(() => {
    const panel = panelRef.current;
    if (!panel || !mobileRef.current) {
      return;
    }
    if (motion.isRunning()) {
      pendingShapeRef.current = true;
      return;
    }
    pendingShapeRef.current = false;
    // Derived rather than passed in, so callers never have to know what could have changed it. The route is the
    // obvious one, but BudgetingSidebar also returns null when its query errors.
    const next = resolvePanelShape(panel);
    shapeRef.current = next;
    panel.toggleAttribute('data-secondary', next === 'wide');
    widthRef.current = Math.max(1, panel.offsetWidth);
    // Not optional: the resting transform is written in terms of the old width.
    motion.repaint();
  }, [motion]);

  // Visibility is deliberately not touched here, paint owns it and renders the resting position before rest is
  // reported. Note the closed panel being untabbable depends on `visibility`; swapping it for opacity would make the
  // whole panel keyboard reachable again.
  const handleRest = useCallback(() => {
    lift(false);

    if (pendingShapeRef.current) {
      pendingShapeRef.current = false;
      applyShape();
    }
  }, [lift, applyShape]);

  const animateTo = useCallback(
    (next: 0 | 1) => {
      lift(true);
      motion.release(next);
      setIsOpen(next === 1);
    },
    [lift, motion],
  );

  const close = useCallback(() => {
    animateTo(0);
  }, [animateTo]);

  const { rememberFocus } = useDrawerFocus({ isMobile, isOpen, panelRef, close });

  const open = useCallback(() => {
    rememberFocus();
    animateTo(1);
  }, [rememberFocus, animateTo]);

  const toggle = useCallback(() => {
    // Against the live position, not the committed state, so pressing the toggle mid animation does whatever the user
    // can actually see on screen.
    if (motion.location() < 0.5) {
      open();
      return;
    }
    close();
  }, [motion, open, close]);

  const { gesture, releaseDrag } = useSwipeDrag({
    isMobile,
    isOpenRef,
    lift,
    mobileRef,
    motion,
    onCommit: setIsOpen,
    pageRef,
    rememberFocus,
    widthRef,
  });

  // A throw inside the frame loop would leave the drawer frozen with the page inert and scrolling locked. Clearing
  // the inline styles falls back to the stylesheet, where the resting state is closed. The drag has to be released
  // too, or pointermoves keep driving the integrator back into whatever just threw.
  const handleError = useCallback(
    (error: unknown) => {
      captureException(error);
      pendingShapeRef.current = false;
      try {
        releaseDrag();
        reset();
      } catch {
        // Don't throw out of the error handler as well.
      }
      setIsOpen(false);
    },
    [reset, releaseDrag],
  );

  useLayoutEffect(() => {
    sinkRef.current = { render: paint, onRest: handleRest, onError: handleError };
  }, [paint, handleRest, handleError]);

  useLayoutEffect(() => {
    isOpenRef.current = isOpen;
  }, [isOpen]);

  // Everything the drawer owns has to be handed back when crossing above the breakpoint. A drag left live here would
  // keep the gesture flag set, which kills pull to refresh for the rest of the session.
  useLayoutEffect(() => {
    mobileRef.current = isMobile;
    if (!isMobile) {
      releaseDrag();
      // paint and lift already no-op above the breakpoint, so this only resets the physics.
      motion.snap(0);
      reset();
      setIsOpen(false);
      return;
    }

    applyShape();
  }, [isMobile, motion, reset, applyShape, releaseDrag]);

  // Close on navigation. A layout effect, not a passive one: React unmounts the old secondary column in the same
  // commit that changes the pathname, and a passive effect leaves room for a paint of a full width panel with nothing
  // in most of it.
  useLayoutEffect(() => {
    if (!isMobile || !pathname) {
      return;
    }
    // First. A <Redirect> or an action button can navigate with a finger still down, and a live drag would either get
    // the settle spring swapped in mid gesture or carry on from a stale origin after a snap.
    const drag = releaseDrag();

    const shapeChanged = resolvePanelShape(panelRef.current) !== shapeRef.current;

    // Check the target and the loop, not just the position. location trails target, so in the frames between a drag
    // arming and the first physics step it still reads 0 while the finger is already gone. Gating on location alone
    // let a navigation leave the loop running toward an unreachable target.
    if (motion.isRunning() || motion.location() > 0.001 || motion.target() > 0.001) {
      lift(true);
      if (shapeChanged) {
        // Animating would show the panel changing width and contents on its way out.
        motion.snap(0);
      } else {
        motion.release(0);
      }
    } else if (drag) {
      // A press that promoted layers but never moved anything. Nothing is going to rest, so hand them back by hand.
      lift(false);
    }
    setIsOpen(false);
    applyShape();
  }, [pathname, isMobile, motion, lift, applyShape, releaseDrag]);

  // A rotation changes what 86vw works out to, and neither the route nor the breakpoint effect re-runs for that.
  useEffect(() => {
    if (!isMobile) {
      return;
    }
    const onResize = () => applyShape();
    window.addEventListener('resize', onResize);
    return () => window.removeEventListener('resize', onResize);
  }, [isMobile, applyShape]);

  // Teardown.
  useEffect(
    () => () => {
      motion.destroy();
      // Unmounting mid drag with the gesture flag still set would kill pull to refresh for the rest of the session.
      releaseDrag();
      window.clearTimeout(holdRef.current);
      document.body.style.overflow = '';
      document.body.style.userSelect = '';
      document.documentElement.removeAttribute('data-sidebar-past');
    },
    [motion, releaseDrag],
  );

  // Tapping the nav item for the route you are already on doesn't change the pathname, so the navigation effect never
  // runs and the drawer would just sit there.
  const onPanelClick = useCallback(
    (event: React.MouseEvent<HTMLElement>) => {
      if (!mobileRef.current) {
        return;
      }
      const target = event.target as Element | null;
      if (target?.closest('a[href]')) {
        close();
      }
    },
    [close],
  );

  const context = useMemo(() => ({ isOpen, open, close, toggle }), [isOpen, open, close, toggle]);

  // Conditional rather than permanent, because the element is always mounted. display: contents keeps it in the
  // accessibility tree, so a permanent dialog role would hide the real desktop sidebars.
  const dialogProps = isMobile
    ? ({
        'aria-hidden': !isOpen,
        'aria-label': 'Navigation',
        'aria-modal': isOpen || undefined,
        role: 'dialog',
        tabIndex: -1,
      } as const)
    : null;

  return (
    <MobileSidebarContext.Provider value={context}>
      <aside className={styles.panel} onClick={onPanelClick} ref={panelRef} {...dialogProps} {...gesture}>
        {props.panel}
      </aside>
      <div aria-hidden='true' className={styles.scrim} ref={scrimRef} />
      <div aria-hidden='true' className={styles.edge} ref={edgeRef} />
      {/* Never transformed, see paint(). */}
      <div className={styles.page} inert={isMobile && isOpen} ref={pageRef} {...gesture}>
        {props.children}
      </div>
      {/*
        biome-ignore lint/a11y/noStaticElementInteractions: Pointer only and aria-hidden. A role here would put a full
        viewport button in front of screen reader users; keyboard and assistive tech close with Escape or the toggle.
      */}
      <div aria-hidden='true' className={styles.catcher} onClick={close} ref={catcherRef} {...gesture} />
    </MobileSidebarContext.Provider>
  );
}
