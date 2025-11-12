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
import { useNavigate } from 'react-router-dom';

export default function PullToRefresh(): JSX.Element {
  const navigate = useNavigate();

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
    // Add loading class to determine when to add spin animation
    const forceReload = () => {
      refreshDiv.current?.classList.add('loading');
      setTimeout(() => {
        navigate(0);
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
      if (document.querySelector('ul')?.scrollTop > 0) {
        return;
      }

      setPullStartPoint(e.targetTouches[0].screenY);

      if (window.scrollY === 0 && window.scrollX === 0) {
        refreshDiv.current?.classList.add('block');
        refreshDiv.current?.classList.remove('hidden');
        document.body.style.touchAction = 'none';
        document.body.style.overscrollBehavior = 'none';
        if (html) {
          html.style.overscrollBehaviorY = 'none';
        }
      } else {
        refreshDiv.current?.classList.remove('block');
        refreshDiv.current?.classList.add('hidden');
      }
    };

    // Tracks how far we have pulled down the refresh icon
    const pullDown = async (e: TouchEvent) => {
      // If there is a dialog open, then do nothing.
      if (document.querySelectorAll('[role="dialog"]').length > 0) {
        return;
      }
      // Prevent accidently pull to refresh on the wrong main view.
      if (document.querySelector('ul')?.scrollTop > 0) {
        return;
      }
      // On the details pages don't allow pull to refresh either
      if (document.querySelector('form > div.overflow-y-auto')?.scrollTop > 0) {
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
      if (document.querySelector('ul')?.scrollTop > 0) {
        return;
      }
      if (document.querySelector('form > div.overflow-y-auto')?.scrollTop > 0) {
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
  }, [pullDownReloadThreshold, pullStartPoint, navigate]);

  return (
    <div
      className='absolute left-0 right-0 z-50 m-auto w-fit transition-all ease-out'
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
      <div
        className='relative -top-16 h-9 w-9 rounded-full border-1 border-dark-monetr-border bg-dark-monetr-background shadow-md shadow-black ring-1 ring-dark-monetr-background flex items-center justify-center'
        style={{ animationDirection: 'reverse' }}
      >
        <div className={refreshDiv.current?.classList.contains('loading') ? 'animate-spin' : undefined}>
          <RefreshCcw
            className={`rounded-full ${
              pullDownReloadThreshold && 'rotate-180'
            } text-indigo-500 transition-all duration-300`}
          />
        </div>
      </div>
    </div>
  );
}
