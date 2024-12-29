import Feature from '@monetr/docs/components/Feature';

export default function FreeToUse(): JSX.Element {
  return (
    <Feature
      title={ (
        <h1 className='text-2xl lg:text-3xl text-start sm:text-center font-semibold'>
          See What's Leftover
        </h1>
      ) }
      description={ (
        <h2 className='text-lg text-start sm:text-center text-dark-monetr-content'>
          monetr keeps track of how much you have put aside for your budgets so it can tell you exactly how much you
          have left over to use or spend.
        </h2>
      ) }
      className='col-span-full md:col-span-2'
      link='/documentation/use/free_to_use/'
      linkText='Learn About Free-To-Use'
    />
  );
}
