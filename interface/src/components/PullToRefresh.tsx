import type React from 'react';
import { lazy, Suspense, useEffect, useRef, useState } from 'react';

import styles from './PullToRefresh.module.scss';

// All of the actual pull to refresh code lives in PullToRefreshGesture, and we only load it once the page has
// finished loading. This way it isnt part of the initial bundle and doesnt slow down the first render.
const PullToRefreshGesture = lazy(() => import('@monetr/interface/components/PullToRefreshGesture'));

export interface PullToRefreshProps {
  children: React.ReactNode;
}

// PullToRefresh wraps the entire app. This part cant be lazy loaded cuz its the container that everything
// renders inside of, if we lazy loaded this then nothing would render until it finished loading. It is also the
// element that gets moved down during a pull.
export default function PullToRefresh(props: PullToRefreshProps): React.JSX.Element {
  const rootRef = useRef<HTMLDivElement>(null);
  // If the page already finished loading by the time we render then we dont need to wait for the load event.
  const [pageLoaded, setPageLoaded] = useState(() => document.readyState === 'complete');

  useEffect(() => {
    if (pageLoaded) {
      return;
    }

    const handleLoad = () => setPageLoaded(true);
    window.addEventListener('load', handleLoad, { once: true });
    return () => window.removeEventListener('load', handleLoad);
  }, [pageLoaded]);

  return (
    <div className={styles.root} ref={rootRef}>
      {props.children}
      {/* Fallback is null because there is nothing to show until the user actually pulls. */}
      {pageLoaded && (
        <Suspense fallback={null}>
          <PullToRefreshGesture rootRef={rootRef} />
        </Suspense>
      )}
    </div>
  );
}
