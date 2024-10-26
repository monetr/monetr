import React from 'react';

import { Logo } from '@monetr/interface/assets';

export default function CenteredLogo(): JSX.Element {
  return (
    <div className='flex justify-center w-full mt-5 mb-5'>
      <img src={ Logo } className='w-1/3' />
    </div>
  );
}
