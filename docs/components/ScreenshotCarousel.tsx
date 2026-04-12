import * as React from 'react';

import {
  Carousel,
  CarouselContent,
  CarouselItem,
  CarouselNext,
  CarouselPrevious,
} from '@monetr/docs/components/Carousel';

import Autoplay from 'embla-carousel-autoplay';

export default function ScreenshotCarousel() {
  const plugin = React.useRef(Autoplay({ delay: 10000, stopOnInteraction: true }));

  return (
    <div className='m-view-width'>
      <Carousel
        className='w-full'
        onMouseEnter={plugin.current.stop}
        onMouseLeave={plugin.current.reset}
        plugins={[plugin.current]}
      >
        <CarouselContent>
          <CarouselItem>
            <div className='p-3'>
              <img
                alt='Screenshot of the monetr app showing the main view of transactions and balances of the budget.'
                className='hidden sm:block rounded-md z-10 shadow-lg w-full'
                src='/assets/screenshot.png'
              />
              <img
                alt='Screenshot of the monetr app showing the main view of transactions and balances of the budget.'
                className='block sm:hidden rounded-md z-10 shadow-lg w-full'
                src='/assets/screenshot_mobile.png'
              />
            </div>
          </CarouselItem>
          <CarouselItem>
            <div className='p-3'>
              <img
                alt='Screenshot of the monetr app showing expenses view on desktop.'
                className='hidden sm:block rounded-md z-10 shadow-lg w-full'
                src='/assets/screenshot_expenses.png'
              />
              <img
                alt='Screenshot of the monetr app showing expenses view on mobile.'
                className='block sm:hidden rounded-md z-10 shadow-lg w-full'
                src='/assets/screenshot_expenses_mobile.png'
              />
            </div>
          </CarouselItem>
          <CarouselItem>
            <div className='p-3'>
              <img
                alt='Screenshot of the monetr app showing funding and contribution details on desktop.'
                className='hidden sm:block rounded-md z-10 shadow-lg w-full'
                src='/assets/screenshot_funding.png'
              />
              <img
                alt='Screenshot of the monetr app showing funding and contribution details on mobile.'
                className='block sm:hidden rounded-md z-10 shadow-lg w-full'
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
