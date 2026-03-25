import Feature from '@monetr/docs/components/Feature';

export default function Plaid(): JSX.Element {
  return (
    <Feature
      className='col-span-full md:col-span-2'
      description={
        <h2 className='text-lg text-start sm:text-center text-dark-monetr-content'>
          Using Plaid, monetr can receive secure automated updates from your bank. You never need to manually import
          your transactions or balances.
        </h2>
      }
      link='/documentation/use/plaid/'
      linkText='Learn About Plaid'
      title={
        <h1 className='text-2xl lg:text-3xl text-start sm:text-center font-semibold'>Connect Your Bank With Plaid</h1>
      }
    />
  );
}
