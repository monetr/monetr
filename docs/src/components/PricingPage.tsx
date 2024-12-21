/* eslint-disable max-len */
import JoinWaitlist from '@monetr/docs/components/JoinWaitlist';
import PricingCards from '@monetr/docs/components/Pricing/PricingCards';


export default function PricingPage(): JSX.Element {
  return (
    <div className='w-full relative m-view-height py-8 flex flex-col gap-16'>
      <div className='absolute inset-0 overflow-hidden pointer-events-none -z-10' aria-hidden='true'>
        <div className='absolute flex items-center justify-center top-0 -translate-y-1/2 left-1/2 -translate-x-1/2 w-full sm:w-1/2 aspect-square'>
          <div className='absolute inset-0 translate-z-0 bg-purple-500 rounded-full blur-[120px] opacity-50 min-h-[10vh]' />
        </div>
      </div>

      <div className='flex flex-col m-view-width mx-auto items-center gap-4'>
        <div className='max-w-3xl flex flex-col gap-8 text-center items-center'>
          <div className='flex items-center justify-center ml-3 p-4'>
            <span className='absolute mx-auto flex border w-fit bg-gradient-to-r blur-xl opacity-50 from-purple-100 via-purple-200 to-purple-300 bg-clip-text text-5xl sm:text-6xl font-extrabold text-transparent text-center select-none'>
              Pricing
            </span>
            <h1 className='h-24 relative top-0 justify-center flex bg-gradient-to-r items-center from-purple-100 via-purple-200 to-purple-300 bg-clip-text text-5xl sm:text-6xl font-extrabold text-transparent text-center select-auto'>
              Pricing
            </h1>
          </div>
        </div>

        <h2 className='text-xl sm:text-2xl font-medium'>
          Get notified when monetr launches!
        </h2>
        <JoinWaitlist placeholder='Enter your email' />

      </div>

      <div className='flex w-full'>
        <PricingCards />
      </div>


      <div className='flex justify-center items-center flex-col gap-6 px-8'>
        <h2 className='font-bold text-4xl text-center'>
          Frequently Asked Questions
        </h2>
        <ul className='max-w-2xl flex flex-col gap-4'>
          <li>
            <h4 className='text-lg font-bold'>
              Can I cancel anytime?
            </h4>
            <p>
              Yes. You can cancel at any time. You will have access to your account until the end of the current billing
              period.
            </p>
          </li>
          <li>
            <h4 className='text-lg font-bold'>
              Is a payment method required to try monetr?
            </h4>
            <p>
              monetr does not require a payment method to try it out, however we do limit you to a single Plaid
              connection during the trial to try to prevent spam.
            </p>
          </li>
          <li>
            <h4 className='text-lg font-bold'>
              What if I sign up and find monetr doesn't meet my needs?
            </h4>
            <p>
              If you have not subscribed yet then you don't need to do anything! Your account will become inactive
              automatically at the end of your trial. If you have already subscribed then you can cancel your
              subscription at any time!
            </p>
          </li>
        </ul>
      </div>
    </div>
  );
}
