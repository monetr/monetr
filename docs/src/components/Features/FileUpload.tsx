import React from 'react';

import Feature from '@monetr/docs/components/Feature';

export default function FileUpload(): JSX.Element {
  return (
    <Feature
      title={
        <h1 className='text-2xl lg:text-3xl text-start sm:text-center font-semibold'>Import Transactions Manually</h1>
      }
      description={
        <h2 className='text-lg text-start sm:text-center text-dark-monetr-content'>
          Upload OFX files directly from your bank account to make it easy to get data into monetr.
        </h2>
      }
      className='col-span-full md:col-span-2'
      link='/documentation/use/transactions/uploads/'
      linkText='Learn About Transaction Uploads'
    />
  );
}
