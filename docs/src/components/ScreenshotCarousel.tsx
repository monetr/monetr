import * as React from 'react';

import { Carousel, CarouselContent, CarouselItem, CarouselNext, CarouselPrevious } from '@monetr/docs/components/Carousel';

import Autoplay from 'embla-carousel-autoplay';

export function CarouselPlugin() {
  const plugin = React.useRef(
    Autoplay({ delay: 2000, stopOnInteraction: true })
  );

  return (
    <Carousel
      plugins={ [plugin.current] }
      className='w-full max-w-xs'
      onMouseEnter={ plugin.current.stop }
      onMouseLeave={ plugin.current.reset }
    >
      <CarouselContent>
        {Array.from({ length: 5 }).map((_, index) => (
          <CarouselItem key={ index }>
            <div className='p-1'>

            </div>
          </CarouselItem>
        ))}
      </CarouselContent>
      <CarouselPrevious />
      <CarouselNext />
    </Carousel>
  );
}
