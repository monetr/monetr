import { ArrowRight, CircleCheck } from 'lucide-react';
import Link from 'next/link';

export default function PricingCards(): JSX.Element {
  return (
    <div className='flex flex-wrap items-center justify-center mx-auto gap-4 md:gap-0 w-3/4 max-w-5xl'>
      <div className='w-full p-6 bg-black border border-zinc-700 rounded-lg md:w-1/2 bg-opacity-20 md:rounded-r-none md:p-8 backdrop-blur-sm'>
        <div className='mb-6'>
          <h3 className='text-2xl font-semibold jakarta text-gray-100 md:text-4xl'>You Host It</h3>
        </div>
        <div className='mb-4 space-x-2'>
          <span className='text-4xl font-bold text-gray-100'>Free Forever</span>
        </div>
        <ul className='mb-6 space-y-2 text-gray-300'>
          <li className='flex items-center gap-1.5'>
            <CircleCheck className='h-5 w-5' />
            <span className=''>Keep Your Data On Your Devices</span>
          </li>
          <li className='flex items-center gap-1.5'>
            <CircleCheck className='h-5 w-5' />
            <span className=''>Community Support</span>
          </li>
        </ul>
        <Link
          href='/documentation/install/'
          className='block px-8 py-3 text-md font-semibold text-center text-gray-100 transition duration-100 bg-white rounded-lg outline-none bg-opacity-10 hover:bg-opacity-20 md:text-base'
        >
          Get Started Now
        </Link>
      </div>

      <div className='w-full p-6 rounded-lg shadow-xl md:w-1/2 bg-gradient-to-br from-monetr-brand to-purple-400 md:p-8'>
        <div className='flex flex-col items-start justify-between gap-4 mb-6 lg:flex-row'>
          <div>
            <h3 className='text-2xl font-semibold text-white jakarta md:text-4xl'>We Host It</h3>
          </div>
        </div>
        <div className='mb-4 space-x-2 flex flex-wrap'>
          <span className='text-4xl font-bold text-white'>$4/month</span>
          <sup className='mt-3'>*USD, tax included in price</sup>
        </div>
        <ul className='mb-6 space-y-2 text-indigo-100'>
          <li className='flex items-center gap-1.5'>
            <CircleCheck className='h-5 w-5' />
            <span className=''>30 Day Free Trial</span>
          </li>
          <li className='flex items-center gap-1.5'>
            <CircleCheck className='h-5 w-5' />
            <span className=''>Access Anywhere Via Web</span>
          </li>
          <li className='flex items-center gap-1.5'>
            <CircleCheck className='h-5 w-5' />
            <span className=''>Automatic Updates Via Plaid</span>
          </li>
          <li className='flex items-center gap-1.5'>
            <CircleCheck className='h-5 w-5' />
            <span className=''>Import Via OFX</span>
          </li>
          <li className='flex items-center gap-1.5'>
            <CircleCheck className='h-5 w-5' />
            <span className=''>Email Support</span>
          </li>
        </ul>
        <Link
          href='https://my.monetr.app/register'
          className='inline-flex w-full items-center justify-center gap-2 px-8 py-3 text-md font-semibold text-white transition duration-100 bg-white rounded-lg outline-none bg-opacity-10 hover:bg-opacity-20 md:text-base'
        >
          Try Free for 30 Days
          <ArrowRight />
        </Link>
      </div>
    </div>
  );
}
