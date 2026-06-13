import { useEffect, useRef } from 'react';

import styles from './FundingBar.module.scss';

// FundingBar paints a thin scroll-progress bar across the very top of the
// viewport. It reads as the "funded" portion of the page, filling as you scroll
// toward the bottom. The scroll handler is passive and rAF-throttled, and the
// fill transition is dropped under prefers-reduced-motion (see the module).
export default function FundingBar(): React.JSX.Element {
  const fillRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const fill = fillRef.current;
    if (!fill) {
      return;
    }

    let frame = 0;
    const update = () => {
      frame = 0;
      const doc = document.documentElement;
      const max = doc.scrollHeight - doc.clientHeight;
      const progress = max > 0 ? (doc.scrollTop / max) * 100 : 0;
      fill.style.width = `${progress}%`;
    };

    const onScroll = () => {
      if (frame) {
        return;
      }
      frame = requestAnimationFrame(update);
    };

    update();
    window.addEventListener('scroll', onScroll, { passive: true });
    window.addEventListener('resize', onScroll, { passive: true });
    return () => {
      window.removeEventListener('scroll', onScroll);
      window.removeEventListener('resize', onScroll);
      if (frame) {
        cancelAnimationFrame(frame);
      }
    };
  }, []);

  return (
    <div aria-hidden='true' className={styles.track}>
      <div className={styles.fill} ref={fillRef} />
    </div>
  );
}
