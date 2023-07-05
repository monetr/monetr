/* eslint-disable max-len */
import React, { useState } from 'react';
import { Logout, Settings } from '@mui/icons-material';

import BankSidebarItem from './BankSidebarItem';

import { Logo } from 'assets';


export default function BankSidebar(): JSX.Element {
  // Important things to note. The width is 16. The width of the icons is 12.
  // This leaves a padding of 2 on each side, which isn't even needed with items-center? Not sure which
  // would be better.
  // py-2 pushes the icons down the same distance they are from the side.
  // gap-2 makes sure they are evenly spaced.
  // TODO: Need to show an active state on the icon somehow. This might need more padding.

  const [activeOne, setActiveOne] = useState('ins_15');

  return (
    <div className='hidden lg:visible w-16 h-full lg:flex items-center lg:py-4 gap-4 lg:flex-col dark:border-r-dark-monetr-border flex-none border border-transparent'>
      <div className='h-10 w-10'>
        <img src={ Logo } className="w-full" />
      </div>
      <hr className='w-1/2 border-0 border-b-[thin] dark:border-dark-monetr-border' />
      <div className='h-full w-full flex items-center flex-col overflow-y-auto'>
        <BankSidebarItem instituionId='ins_15' active={ activeOne === 'ins_15' } onClick={ () => setActiveOne('ins_15') } />
        <BankSidebarItem instituionId='ins_116794' active={ activeOne === 'ins_116794' } onClick={ () => setActiveOne('ins_116794') }  />
        <BankSidebarItem instituionId='ins_127990' active={ activeOne === 'ins_127990' } onClick={ () => setActiveOne('ins_127990') }  />
        <BankSidebarItem instituionId='ins_3' active={ activeOne === 'ins_3' }  onClick={ () => setActiveOne('ins_3') } />
      </div>
      <Settings className='dark:hover:text-dark-monetr-content-emphasis dark:text-dark-monetr-content-subtle cursor-pointer' />
      <Logout className='dark:hover:text-dark-monetr-content-emphasis dark:text-dark-monetr-content-subtle cursor-pointer' />
    </div>
  );
}
