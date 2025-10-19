import Logo from '@monetr/interface/assets/Logo';

export default function CenteredLogo(): JSX.Element {
  return (
    <div className='flex justify-center w-full mt-5 mb-5'>
      <Logo className='w-1/3' />
    </div>
  );
}
