import React from 'react';

import Feature from '@monetr/docs/components/Feature';

export default function SelfHost(): JSX.Element {
  return (
    <Feature
      title={<h1 className='text-2xl lg:text-3xl text-start sm:text-center font-semibold'>Self-Host</h1>}
      description={
        <h2 className='text-lg text-start sm:text-center text-dark-monetr-content'>
          Host monetr yourself on your own hardware, for free. Keeping your data private.
        </h2>
      }
      className='col-span-full md:col-span-2'
      link='/documentation/install/'
      linkText='Installation Guide'
    />
  );
}
