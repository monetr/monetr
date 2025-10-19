import * as React from 'react';

import TransactionScreenshot from '@monetr/docs/assets/screenshot.png';
import ExpensesScreenshot from '@monetr/docs/assets/screenshot_expenses.png';
import ExpensesScreenshotMobile from '@monetr/docs/assets/screenshot_expenses_mobile.png';
import FundingScreenshot from '@monetr/docs/assets/screenshot_funding.png';
import FundingScreenshotMobile from '@monetr/docs/assets/screenshot_funding_mobile.png';
import TransactionsScreenshotMobile from '@monetr/docs/assets/screenshot_mobile.png';
import {
  Carousel,
  CarouselContent,
  CarouselItem,
  CarouselNext,
  CarouselPrevious,
} from '@monetr/docs/components/Carousel';

import Autoplay from 'embla-carousel-autoplay';
import Image from 'next/image';

export default function ScreenshotCarousel() {
  const plugin = React.useRef(Autoplay({ delay: 10000, stopOnInteraction: true }));

  return (
    <div className='m-view-width'>
      <Carousel
        plugins={[plugin.current]}
        className='w-full'
        onMouseEnter={plugin.current.stop}
        onMouseLeave={plugin.current.reset}
      >
        <CarouselContent>
          <CarouselItem>
            <div className='p-3'>
              <Image
                priority={true}
                src={TransactionScreenshot}
                alt='Screenshot of the monetr app showing the main view of transactions and balances of the budget.'
                className='hidden sm:block rounded-md z-10 shadow-lg'
              />
              <Image
                priority={true}
                src={TransactionsScreenshotMobile}
                alt='Screenshot of the monetr app showing the main view of transactions and balances of the budget.'
                className='block sm:hidden rounded-md z-10 shadow-lg'
              />
            </div>
          </CarouselItem>
          <CarouselItem>
            <div className='p-3'>
              <Image
                src={ExpensesScreenshot}
                alt='Screenshot of the monetr app showing expenses view on desktop.'
                className='hidden sm:block rounded-md z-10 shadow-lg'
              />
              <Image
                src={ExpensesScreenshotMobile}
                alt='Screenshot of the monetr app showing expenses view on mobile.'
                className='block sm:hidden rounded-md z-10 shadow-lg'
              />
            </div>
          </CarouselItem>
          <CarouselItem>
            <div className='p-3'>
              <Image
                src={FundingScreenshot}
                alt='Screenshot of the monetr app showing funding and contribution details on destop.'
                className='hidden sm:block rounded-md z-10 shadow-lg'
              />
              <Image
                src={FundingScreenshotMobile}
                alt='Screenshot of the monetr app showing funding and contribution details on mobile.'
                className='block sm:hidden rounded-md z-10 shadow-lg'
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
