import * as React from 'react';

import {
  Carousel,
  CarouselContent,
  CarouselItem,
  CarouselNext,
  CarouselPrevious,
} from '@monetr/docs/components/Carousel';

import styles from './ScreenshotCarousel.module.scss';

import Autoplay from 'embla-carousel-autoplay';

export default function ScreenshotCarousel() {
  const plugin = React.useRef(Autoplay({ delay: 10000, stopOnInteraction: true }));

  return (
    <div className='m-view-width'>
      <Carousel
        className={styles.carousel}
        onMouseEnter={plugin.current.stop}
        onMouseLeave={plugin.current.reset}
        plugins={[plugin.current]}
      >
        <CarouselContent>
          <CarouselItem>
            <div className={styles.slide}>
              <img
                alt='Screenshot of the monetr app showing the main view of transactions and balances of the budget.'
                className={styles.imageDesktop}
                src='/assets/screenshot.png'
              />
              <img
                alt='Screenshot of the monetr app showing the main view of transactions and balances of the budget.'
                className={styles.imageMobile}
                src='/assets/screenshot_mobile.png'
              />
            </div>
          </CarouselItem>
          <CarouselItem>
            <div className={styles.slide}>
              <img
                alt='Screenshot of the monetr app showing expenses view on desktop.'
                className={styles.imageDesktop}
                src='/assets/screenshot_expenses.png'
              />
              <img
                alt='Screenshot of the monetr app showing expenses view on mobile.'
                className={styles.imageMobile}
                src='/assets/screenshot_expenses_mobile.png'
              />
            </div>
          </CarouselItem>
          <CarouselItem>
            <div className={styles.slide}>
              <img
                alt='Screenshot of the monetr app showing funding and contribution details on desktop.'
                className={styles.imageDesktop}
                src='/assets/screenshot_funding.png'
              />
              <img
                alt='Screenshot of the monetr app showing funding and contribution details on mobile.'
                className={styles.imageMobile}
                src='/assets/screenshot_funding_mobile.png'
              />
            </div>
          </CarouselItem>
        </CarouselContent>
        <CarouselPrevious />
        <CarouselNext />
      </Carousel>
    </div>
  );
}
