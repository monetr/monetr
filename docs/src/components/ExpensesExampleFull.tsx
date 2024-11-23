/* eslint-disable max-len */
'use client';

import React from 'react';

import MonetrWrapper from '@monetr/docs/components/MonetrWrapper';
import Monetr from '@monetr/interface/monetr';

export default function ExpensesExampleFull(): JSX.Element {
  const initialRoute = '/bank/bac_01gds6eqsq7h5mgevwtmw3cyxb/expenses';

  return (
    <div className='w-full h-full rounded-2xl shadow-2xl z-10 backdrop-blur-md bg-black/90 pointer-events-none select-none max-w-[1280px] max-h-[720px] aspect-video-vertical md:aspect-video'>
      <MonetrWrapper initialRoute={ initialRoute }>
        <Monetr />
      </MonetrWrapper>
    </div>
  );
}
