import React from 'react';

import Feature from '@monetr/docs/components/Feature';

export default function SourceVisible(): JSX.Element {
  return (
    <Feature
      title={ (
        <h1 className='text-2xl lg:text-3xl text-start sm:text-center font-semibold'>
          Source Visible
        </h1>
      ) }
      description={ (
        <h2 className='text-lg text-start sm:text-center text-dark-monetr-content'>
          All of monetr's source code is publically available, you can see exactly how we handle your data and even
          contribute functionality!
        </h2>
      ) }
      className='col-span-full md:col-span-2'
      link='https://github.com/monetr/monetr'
      linkText='See The Source Code'
      linkExternal
    />
  );
}
