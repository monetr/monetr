import React from 'react';

import MLogo from './components/MLogo';
import MSpan from './components/MSpan';

import './loading.css';

export default function Loading(): JSX.Element {


  return (
    <div className="w-full h-full items-center justify-center flex flex-col gap-8">
      <MLogo className='w-24 h-24' />
      <div className="blobs flex items-center justify-center">
        <div className="dot blob dark:bg-dark-monetr-brand" />
        <div className="dots blob dark:bg-dark-monetr-brand" />
        <div className="dots blob dark:bg-dark-monetr-brand" />
        <div className="dots blob dark:bg-dark-monetr-brand" />
      </div>
      <MSpan className='text-3xl'>
        Loading...
      </MSpan>
      <svg className='hidden' xmlns="http://www.w3.org/2000/svg" version="1.1">
        <defs>
          <filter id="magic">
            <feGaussianBlur in="SourceGraphic" stdDeviation="10" result="blur" />
            <feColorMatrix in="blur" mode="matrix" values="1 0 0 0 0  0 1 0 0 0  0 0 1 0 0  0 0 0 18 -7" result="goo" />
            <feBlend in="SourceGraphic" in2="goo" />
          </filter>
        </defs>
      </svg>
    </div>
  );
}
