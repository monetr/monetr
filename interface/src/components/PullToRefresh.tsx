/*
  @license Overseerr
  This code is from https://github.com/sct/overseerr

  MIT License

  Copyright (c) 2020 sct

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
*/

import { useEffect, useRef, useState } from 'react';
import { RefreshCcw } from 'lucide-react';

import styles from './PullToRefresh.module.scss';

export default function PullToRefresh(): JSX.Element {
  const [pullStartPoint, setPullStartPoint] = useState(0);
  const [pullChange, setPullChange] = useState(0);
  const refreshDiv = useRef<HTMLDivElement>(null);

  // Various pull down thresholds that determine icon location
  const pullDownInitThreshold = pullChange > 20;
  const pullDownStopThreshold = 120;
  const pullDownReloadThreshold = pullChange > 340;
  const pullDownIconLocation = pullChange / 3;

  useEffect(() => {
    // Reload function that is called when reload threshold has been hit
    // Set the loading flag so the spin animation starts
    const forceReload = () => {
      refreshDiv.current?.setAttribute('data-loading', 'true');
      setTimeout(() => {
        window.location.reload();
      }, 1000);
    };

    const html = document.querySelector('html');

    // Determines if we are at the top of the page
    // Locks or unlocks page when pulling down to refresh
    const pullStart = (e: TouchEvent) => {
      // If there is a dialog open, then do nothing.
      if (document.querySelectorAll('[role="dialog"]').length > 0) {
        return;
      }
      // Prevent accidently pull to refresh on the wrong main view.
      if ((document.querySelector('ul')?.scrollTop ?? 0) > 0) {
        return;
      }

      setPullStartPoint(e.targetTouches[0].screenY);

      if (window.scrollY === 0 && window.scrollX === 0) {
        refreshDiv.current?.setAttribute('data-visible', 'true');
        document.body.style.touchAction = 'none';
        document.body.style.overscrollBehavior = 'none';
        if (html) {
          html.style.overscrollBehaviorY = 'none';
        }
      } else {
        refreshDiv.current?.setAttribute('data-visible', 'false');
      }
    };

    // Tracks how far we have pulled down the refresh icon
    const pullDown = async (e: TouchEvent) => {
      // If there is a dialog open, then do nothing.
      if (document.querySelectorAll('[role="dialog"]').length > 0) {
        return;
      }
      // Prevent accidently pull to refresh on the wrong main view.
      if ((document.querySelector('ul')?.scrollTop ?? 0) > 0) {
        return;
      }
      // On the details pages don't allow pull to refresh either
      if ((document.querySelector('form > div.overflow-y-auto')?.scrollTop ?? 0) > 0) {
        return;
      }

      const screenY = e.targetTouches[0].screenY;

      const pullLength = pullStartPoint < screenY ? Math.abs(screenY - pullStartPoint) : 0;

      setPullChange(pullLength);
    };

    // Will reload the page if we are past the threshold
    // Otherwise, we reset the pull
    const pullFinish = () => {
      if (document.querySelectorAll('[role="dialog"]').length > 0) {
        return;
      }
      if ((document.querySelector('ul')?.scrollTop ?? 0) > 0) {
        return;
      }
      if ((document.querySelector('form > div.overflow-y-auto')?.scrollTop ?? 0) > 0) {
        return;
      }

      setPullStartPoint(0);

      if (pullDownReloadThreshold) {
        forceReload();
      } else {
        setPullChange(0);
      }

      document.body.style.touchAction = 'auto';
      document.body.style.overscrollBehaviorY = 'auto';
      if (html) {
        html.style.overscrollBehaviorY = 'auto';
      }
    };

    window.addEventListener('touchstart', pullStart, { passive: false });
    window.addEventListener('touchmove', pullDown, { passive: false });
    window.addEventListener('touchend', pullFinish, { passive: false });

    return () => {
      window.removeEventListener('touchstart', pullStart);
      window.removeEventListener('touchmove', pullDown);
      window.removeEventListener('touchend', pullFinish);
    };
  }, [pullDownReloadThreshold, pullStartPoint]);

  return (
    <div
      className={styles.container}
      ref={refreshDiv}
      style={{
        top:
          pullDownIconLocation < pullDownStopThreshold && pullDownInitThreshold
            ? pullDownIconLocation
            : pullDownInitThreshold
              ? pullDownStopThreshold
              : '',
      }}
    >
      <div className={styles.circle} style={{ animationDirection: 'reverse' }}>
        <div className={styles.spinner}>
          <RefreshCcw className={styles.icon} data-reload={pullDownReloadThreshold} />
        </div>
      </div>
    </div>
  );
}
