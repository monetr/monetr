import Feature from '@monetr/docs/components/Feature';

export default function MobileFriendly(): JSX.Element {
  return (
    <Feature
      title={<h1 className='text-2xl lg:text-3xl text-start sm:text-center font-semibold'>Mobile Friendly</h1>}
      description={
        <h2 className='text-lg text-start sm:text-center text-dark-monetr-content'>
          monetr is built to be mobile friendly out of the box, it can even be installed as a full featured web app on
          your desktop or mobile device.
        </h2>
      }
      className='col-span-full md:col-span-2'
    />
  );
}
