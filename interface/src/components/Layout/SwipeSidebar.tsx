import { useCallback, useEffect, useLayoutEffect, useRef, useState } from 'react';
import { captureException } from '@sentry/react';
import { useLocation } from 'wouter';

import { MobileSidebarContext } from '@monetr/interface/components/Layout/MobileSidebarContext';
import {
  clamp,
  createMotion,
  layerHold,
  type Motion,
  type MotionCallbacks,
  prefersReducedMotion,
  scrimOpacity,
} from '@monetr/interface/components/Layout/swipeMotion';
import useDrawerFocus from '@monetr/interface/components/Layout/useDrawerFocus';
import useSwipeDrag from '@monetr/interface/components/Layout/useSwipeDrag';
import useIsMobile from '@monetr/interface/hooks/useIsMobile';

import styles from './SwipeSidebar.module.scss';

type PanelShape = 'rail' | 'wide';

// If the panel has a secondary column then the drawer needs to be wide. We look for data-sidebar-secondary instead of
// checking the route. This way another secondary sidebar doesn't need any changes in here.
function resolvePanelShape(panel: HTMLElement | null): PanelShape {
  if (!panel) {
    return 'rail';
  }
  return panel.querySelector('[data-sidebar-secondary]') === null ? 'rail' : 'wide';
}

export interface SwipeSidebarProps {
  panel: React.ReactNode;
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
  const isOpenRef = useRef(false);
  // On desktop the panel is display: contents and the sidebars inside it are the real layout. So we must never write
  // an inline style unless we are on mobile.
  const mobileRef = useRef(false);
  const previousRef = useRef<{ offscreen: boolean | null; past: boolean | null }>({
    offscreen: null,
    past: null,
  });

  // paint is called every frame with the current position, 0 is closed and 1 is open.
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

    // Never transform the page itself. The header, action bar and floating button are position: fixed inside of it,
    // and a transform would make the page their containing block.
    panel.style.transform = `translate3d(${(visible - 1) * width}px,0,0)`;
    scrim.style.opacity = String(visible * scrimOpacity);
    edge.style.transform = `translate3d(${visible * width}px,0,0)`;
    catcher.style.transform = `translate3d(${visible * width}px,0,0)`;
    edge.style.opacity = String(visible);

    const previous = previousRef.current;
    const offscreen = visible <= 0.001;
    if (previous.offscreen !== offscreen) {
      previous.offscreen = offscreen;
      panel.style.visibility = offscreen ? 'hidden' : 'visible';
      edge.style.visibility = offscreen ? 'hidden' : 'visible';
      // This needs to be based on being offscreen and not on being past halfway. We commit to open at 0.42, and between
      // that and halfway the page would be inert with nothing catching taps.
      catcher.style.pointerEvents = offscreen ? 'none' : 'auto';
    }

    const past = visible > 0.5;
    if (previous.past !== past) {
      previous.past = past;
      // The toggle icon swaps in CSS off of this attribute. This way it follows the drag without a re-render.
      document.documentElement.dataset.sidebarPast = past ? 'true' : 'false';
    }

    if (reducedRef.current) {
      return;
    }
    staggerRef.current.forEach((item, index) => {
      const progress = clamp((position - index * 0.028) * 1.5, 0, 1.05);
      item.style.opacity = String(clamp(progress, 0, 1));
      item.style.transform = `translate3d(${(1 - progress) * -16}px,0,0)`;
    });
  }, []);

  // lift promotes everything that moves while we animate. We hold onto the layers for a moment after things stop. This
  // way a quick open, close and open again doesn't pay for promoting them each time.
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
      // If the rows aren't promoted too then the entire panel gets rastered again every frame.
      staggerRef.current.forEach(item => {
        item.style.willChange = on ? 'transform, opacity' : 'auto';
      });
    };

    if (!on) {
      holdRef.current = window.setTimeout(apply, layerHold);
      return;
    }

    // Find the rows again each time, they change with the route and as queries load.
    staggerRef.current = Array.from(panelRef.current?.querySelectorAll<HTMLElement>('[data-stagger]') ?? []);
    reducedRef.current = prefersReducedMotion();
    // Show the panel before anything moves. Browsers don't raster hidden content, and we don't want the first frame
    // to have to raster the entire panel.
    if (panelRef.current) {
      panelRef.current.style.visibility = 'visible';
    }
    if (edgeRef.current) {
      edgeRef.current.style.visibility = 'visible';
    }
    // Let paint know we made it visible. If we don't then paint never hides it again when a press doesn't turn into a
    // drag.
    previousRef.current.offscreen = false;
    apply();
  }, []);

  // reset puts every inline style we wrote back to what the stylesheet says. We need this when the window gets bigger
  // than the breakpoint, a leftover visibility: hidden would hide the desktop sidebars. And to recover if something in
  // the animation loop throws.
  const reset = useCallback(() => {
    window.clearTimeout(holdRef.current);
    previousRef.current = {
      offscreen: null,
      past: null,
    };
    document.documentElement.removeAttribute('data-sidebar-past');
    document.body.style.overflow = '';

    [panelRef.current, scrimRef.current, edgeRef.current, catcherRef.current].forEach(item => {
      if (!item) {
        return;
      }
      item.style.transform = '';
      item.style.visibility = '';
      item.style.opacity = '';
      item.style.willChange = '';
      item.style.pointerEvents = '';
    });
    staggerRef.current.forEach(item => {
      item.style.transform = '';
      item.style.opacity = '';
      item.style.willChange = '';
    });
    staggerRef.current = [];
  }, []);

  // The motion only gets created once. Its callbacks go through this ref, this way it always calls the latest paint.
  const sinkRef = useRef<MotionCallbacks>({
    render: () => {},
    onRest: () => {},
    onError: () => {},
  });
  const motionRef = useRef<Motion | null>(null);
  motionRef.current ??= createMotion({
    render: position => sinkRef.current.render(position),
    onRest: () => sinkRef.current.onRest(),
    onError: error => sinkRef.current.onError(error),
  });
  const motion = motionRef.current;

  // The width itself comes from CSS, this just picks which one and then measures it. If we are in the middle of an
  // animation then wait until it is done. The panel would visibly jump to the new width if we didn't.
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
    const next = resolvePanelShape(panel);
    shapeRef.current = next;
    panel.toggleAttribute('data-secondary', next === 'wide');
    widthRef.current = Math.max(1, panel.offsetWidth);
    // Paint again since the resting transform used the old width.
    motion.repaint();
  }, [motion]);

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

  const { rememberFocus } = useDrawerFocus({
    isMobile,
    isOpen,
    panelRef,
    close,
  });

  function open() {
    rememberFocus();
    animateTo(1);
  }

  // Check where the drawer actually is instead of isOpen. This way pressing the toggle in the middle of an animation
  // does what the user would expect from what they can see.
  function toggle() {
    if (motion.location() < 0.5) {
      open();
      return;
    }
    close();
  }

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

  // If something throws in the animation loop then the drawer would be stuck with the page inert and scrolling locked.
  // Clearing the inline styles puts it back to closed. We need to release the drag too. If we don't then the next
  // pointermove would run into the same error again.
  const handleError = useCallback(
    (error: unknown) => {
      captureException(error);
      pendingShapeRef.current = false;
      try {
        releaseDrag();
        reset();
      } catch {
        // Do nothing
      }
      setIsOpen(false);
    },
    [reset, releaseDrag],
  );

  useLayoutEffect(() => {
    sinkRef.current = {
      render: paint,
      onRest: handleRest,
      onError: handleError,
    };
  }, [paint, handleRest, handleError]);

  useLayoutEffect(() => {
    isOpenRef.current = isOpen;
  }, [isOpen]);

  // If the window gets bigger than the breakpoint then give back everything we touched. A drag left going here would
  // leave the gesture flag set, and pull to refresh wouldn't work again until a reload.
  useLayoutEffect(() => {
    mobileRef.current = isMobile;
    if (!isMobile) {
      releaseDrag();
      motion.snap(0);
      reset();
      setIsOpen(false);
      return;
    }

    applyShape();
  }, [isMobile, motion, reset, applyShape, releaseDrag]);

  // Close on navigation. This is a layout effect because React removes the old secondary column in the same commit
  // that changes the pathname. With a normal effect we could paint a full width panel with nothing in it.
  useLayoutEffect(() => {
    if (!isMobile || !pathname) {
      return;
    }
    // Release the drag first. A <Redirect> or an action button can navigate while a finger is still down.
    const drag = releaseDrag();

    const shapeChanged = resolvePanelShape(panelRef.current) !== shapeRef.current;

    // Check the target and whether the loop is running, not just the location. The location trails the target, and
    // right after a drag starts it can still be 0 while the drawer is on its way open.
    if (motion.isRunning() || motion.location() > 0.001 || motion.target() > 0.001) {
      lift(true);
      if (shapeChanged) {
        // If we animated then they would see the panel change width on its way out.
        motion.snap(0);
      } else {
        motion.release(0);
      }
    } else if (drag) {
      // The press promoted the layers but never moved anything. Nothing is going to rest and give them back.
      lift(false);
    }
    setIsOpen(false);
    applyShape();
  }, [pathname, isMobile, motion, lift, applyShape, releaseDrag]);

  // Rotating the phone changes the width of the panel, and nothing else would measure it again.
  useEffect(() => {
    if (!isMobile) {
      return;
    }
    const onResize = () => applyShape();
    window.addEventListener('resize', onResize);
    return () => window.removeEventListener('resize', onResize);
  }, [isMobile, applyShape]);

  useEffect(
    () => () => {
      motion.destroy();
      // If we unmount in the middle of a drag then the gesture flag would stay set and break pull to refresh.
      releaseDrag();
      window.clearTimeout(holdRef.current);
      document.body.style.overflow = '';
      document.body.style.userSelect = '';
      document.documentElement.removeAttribute('data-sidebar-past');
    },
    [motion, releaseDrag],
  );

  // Tapping the link for the page you are already on doesn't change the pathname. The navigation effect won't run,
  // we have to close the drawer here.
  function closeOnLinkClick(event: React.MouseEvent<HTMLElement>) {
    if (!mobileRef.current) {
      return;
    }
    const target = event.target as Element | null;
    if (target?.closest('a[href]')) {
      close();
    }
  }

  const context = {
    isOpen,
    open,
    close,
    toggle,
  };

  // Only make the panel a dialog on mobile. It is always mounted, and on desktop display: contents keeps it in the
  // accessibility tree where a dialog role would hide the desktop sidebars.
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
      <aside className={styles.panel} onClick={closeOnLinkClick} ref={panelRef} {...dialogProps} {...gesture}>
        {props.panel}
      </aside>
      <div aria-hidden='true' className={styles.scrim} ref={scrimRef} />
      <div aria-hidden='true' className={styles.edge} ref={edgeRef} />
      <div className={styles.page} inert={isMobile && isOpen} ref={pageRef} {...gesture}>
        {props.children}
      </div>
      {/*
        biome-ignore lint/a11y/noStaticElementInteractions: This is pointer only and aria-hidden. Keyboard and screen
        readers close the drawer with Escape or the toggle.
      */}
      <div aria-hidden='true' className={styles.catcher} onClick={close} ref={catcherRef} {...gesture} />
    </MobileSidebarContext.Provider>
  );
}
