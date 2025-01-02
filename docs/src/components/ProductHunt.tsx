import React from 'react';
import Image from 'next/image';
import Link from 'next/link';

export default function ProductHunt(): JSX.Element {
  return (
    <Link
      href='https://www.producthunt.com/posts/monetr?embed=true&utm_source=badge-featured&utm_medium=badge&utm_souce=badge-monetr'
      target='_blank'
      data-umami-event='Product Hunt'
      className='w-full'
    >
      <Image
        src='https://api.producthunt.com/widgets/embed-image/v1/featured.svg?post_id=656178&theme=light'
        alt='monetr - Personal&#0032;financial&#0032;planning&#0032;focused&#0032;on&#0032;recurring&#0032;expenses&#0046; | Product Hunt'
        style={ {
          'width': '250px',
          'height': '54px',
        } }
        width='250'
        height='54'
      />
    </Link>
  );
}
