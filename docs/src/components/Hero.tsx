import Image from 'next/image';

import MobileTransactionScreenshot from '@monetr/docs/assets/mobile_transactions_example.png';
import TransactionScreenshot from '@monetr/docs/assets/transactions_example.png';
import Expenses from '@monetr/docs/components/Features/Expenses';
import FileUpload from '@monetr/docs/components/Features/FileUpload';
import Forecasting from '@monetr/docs/components/Features/Forecasting';
import FreeToUse from '@monetr/docs/components/Features/FreeToUse';
import MobileFriendly from '@monetr/docs/components/Features/MobileFriendly';
import Plaid from '@monetr/docs/components/Features/Plaid';
import SelfHost from '@monetr/docs/components/Features/SelfHost';
import SourceVisible from '@monetr/docs/components/Features/SourceVisible';
import JoinWaitlist from '@monetr/docs/components/JoinWaitlist';

export default function Hero(): JSX.Element {
  return (
    <div className='w-full relative m-view-height py-8'>
      <div className='absolute inset-0 overflow-hidden pointer-events-none -z-10' aria-hidden='true'>
        <div className='absolute flex items-center justify-center top-0 -translate-y-1/2 left-1/2 -translate-x-1/2 w-full sm:w-1/2 aspect-square'>
          <div className='absolute inset-0 translate-z-0 bg-purple-500 rounded-full blur-[120px] opacity-50 min-h-[10vh]' />
        </div>
      </div>

      <div className='m-view-width flex flex-col mx-auto items-center gap-8'>
        <div className='max-w-3xl flex flex-col gap-8 text-center items-center'>
          <div className='flex items-center justify-center p-4'>
            <span className='absolute mx-auto flex border w-fit bg-gradient-to-r blur-xl opacity-50 from-purple-100 via-purple-200 to-purple-300 bg-clip-text text-4xl sm:text-6xl font-extrabold text-transparent text-center select-none'>
              Coming Soon
            </span>
            <h1 className='h-24 relative top-0 justify-center flex bg-gradient-to-r items-center from-purple-100 via-purple-200 to-purple-300 bg-clip-text text-4xl sm:text-6xl font-extrabold text-transparent text-center select-auto'>
              Coming Soon
            </h1>
          </div>

          <h1 className='text-4xl sm:text-5xl font-bold'>Take control of your finances, paycheck by paycheck</h1>

          <h2 className='text-xl sm:text-2xl font-medium'>
            Put aside what you need, spend what you want.
          </h2>
        </div>

        <div className='space-y-4'>
          <h2 className='text-xl sm:text-2xl font-medium'>
          Get notified when monetr launches!
          </h2>
          <JoinWaitlist placeholder='Enter your email' />
        </div>

        <Image
          priority={ true }
          src={ TransactionScreenshot }
          alt='Screenshot of the monetr app showing the main view of transactions and balances of the budget.'
          className='hidden sm:block rounded-md z-10 shadow-lg'
        />
        <Image
          priority={ true }
          src={ MobileTransactionScreenshot }
          alt='Screenshot of the monetr app showing the main view of transactions and balances of the budget.'
          className='block sm:hidden rounded-md z-10 shadow-lg'
        />

        <h1 className='text-4xl sm:text-5xl font-bold mt-16'>Features</h1>

        <div className='grid grid-cols-4 w-full gap-6'>
          <FreeToUse />
          <Expenses />
          <FileUpload />
          <Plaid />
          <Forecasting />
          <MobileFriendly />
          <SelfHost />
          <SourceVisible />
        </div>
      </div>
    </div>
  );
}
