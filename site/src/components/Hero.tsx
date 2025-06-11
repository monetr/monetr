import { ArrowRight } from 'lucide-react';

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
              Welcome to monetr,
            </span>
            <h1 className='h-24 relative top-0 justify-center flex bg-gradient-to-r items-center from-purple-100 via-purple-200 to-purple-300 bg-clip-text text-4xl sm:text-6xl font-extrabold text-transparent text-center select-auto'>
              Welcome to monetr
            </h1>
          </div>

          <h1 className='text-4xl sm:text-5xl font-bold'>
            Take control of your finances, paycheck by paycheck
          </h1>

          <h2 className='text-xl sm:text-2xl font-medium'>
            Put aside what you need, spend what you want.
          </h2>
        </div>

        <div className='flex flex-col sm:flex-row gap-4'>
          <a
            className='flex-none inline-flex items-center gap-2 rounded-md bg-purple-500 px-3.5 py-2.5 font-semibold text-white shadow-sm hover:bg-purple-400 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-purple-500 text-xl'
            href='https://my.monetr.app/register'
          >
            Try Free for 30 Days
            <ArrowRight />
          </a>
          <a
            className='rounded-md block py-2.5 px-3.5 font-semibold text-center text-white transition duration-100 bg-white outline-none bg-opacity-10 hover:bg-opacity-20 backdrop-blur-sm text-xl '
            href='/documentation/use/getting_started/'
          >
            Learn More
          </a>
        </div>
      </div>
    </div>
  );
}
