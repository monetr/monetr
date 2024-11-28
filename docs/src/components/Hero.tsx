import Image from 'next/image';

import Particles from './Particles';
import MobileTransactionScreenshot from '@monetr/docs/assets/mobile_transactions_example.png';
import TransactionScreenshot from '@monetr/docs/assets/transactions_example.png';

export default function Hero(): JSX.Element {
  return (
    <div className='w-full relative'>
      <div className='absolute inset-0 overflow-hidden pointer-events-none -z-10' aria-hidden='true'>
        <div className='absolute flex items-center justify-center top-0 -translate-y-1/2 left-1/2 -translate-x-1/2 w-full sm:w-1/2 aspect-square'>
          <div className='absolute inset-0 translate-z-0 bg-purple-500 rounded-full blur-[120px] opacity-50 min-h-[10vh]' />
        </div>
      </div>
      <div className='max-md:hidden absolute bottom-0 -mb-20 left-2/3 -translate-x-1/2 blur-2xl opacity-70 pointer-events-none' aria-hidden='true'>
        <svg xmlns='http://www.w3.org/2000/svg' width='434' height='427'>
          <defs>
            <linearGradient id='bs5-a' x1='19.609%' x2='50%' y1='14.544%' y2='100%'>
              <stop offset='0%' stopColor='#A855F7' />
              <stop offset='100%' stopColor='#6366F1' stopOpacity='0' />
            </linearGradient>
          </defs>
          <path fill='url(#bs5-a)' fillRule='evenodd' d='m661 736 461 369-284 58z' transform='matrix(1 0 0 -1 -661 1163)' />
        </svg>
      </div>
      <Particles className='absolute inset-0 -z-10' />

      <div className='m-view-height m-view-width flex flex-col py-16 mx-auto items-center gap-8'>
        <div className='max-w-3xl flex flex-col gap-8 text-center items-center'>
          <div className='flex items-center justify-center ml-3 p-4'>
            <span className='absolute mx-auto flex border w-fit bg-gradient-to-r blur-xl opacity-50 from-purple-100 via-purple-200 to-purple-300 bg-clip-text text-4xl sm:text-6xl font-extrabold text-transparent text-center select-none'>
              Coming Soon
            </span>
            <h1 className='h-24 relative top-0 justify-center flex bg-gradient-to-r items-center from-purple-100 via-purple-200 to-purple-300 bg-clip-text text-4xl sm:text-6xl font-extrabold text-transparent text-center select-auto'>
              Coming Soon
            </h1>
          </div>

          <h1 className='text-4xl sm:text-5xl font-bold'>Always know what you can spend</h1>

          <h2 className='text-lg sm:text-xl font-medium'>
            Put a bit of money aside every time you get paid. Always be sure you'll have enough to cover your bills, and
            know what you have left-over to save or spend on whatever you'd like.
          </h2>
        </div>

        <Image
          priority={ true }
          src={ TransactionScreenshot }
          alt='Easily keep track of transactions'
          className='hidden sm:block rounded-md z-10 shadow-lg'
        />
        <Image
          priority={ true }
          src={ MobileTransactionScreenshot }
          alt='Easily keep track of transactions'
          className='block sm:hidden rounded-md z-10 shadow-lg'
        />
      </div>
    </div>
  );
}
