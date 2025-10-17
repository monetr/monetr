

import Feature from '@monetr/docs/components/Feature';

export default function Expenses(): JSX.Element {
  return (
    <Feature
      title={<h1 className='text-2xl lg:text-3xl text-start sm:text-center font-semibold'>Track Recurring Expenses</h1>}
      description={
        <h2 className='text-lg text-start sm:text-center text-dark-monetr-content'>
          monetr let's you budget for things that happen on all kinds of intervals, not confining you to a single
          monthly budget.
        </h2>
      }
      className='col-span-full md:col-span-2'
      link='/documentation/use/expense/'
      linkText='Learn More About Expenses'
    />
  );
}
