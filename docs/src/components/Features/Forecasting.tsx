import Feature from '@monetr/docs/components/Feature';

export default function Forecasting(): JSX.Element {
  return (
    <Feature
      title={
        <h1 className='text-2xl lg:text-3xl text-start sm:text-center font-semibold'>See Your Financial Future</h1>
      }
      description={
        <h2 className='text-lg text-start sm:text-center text-dark-monetr-content'>
          See a forecast of your finances based on your budget, so you can see how much you'll have and when you'll have
          it.
        </h2>
      }
      className='col-span-full md:col-span-2'
      link='/documentation/use/forecasting/'
      linkText='Learn About Forecasting'
    />
  );
}
