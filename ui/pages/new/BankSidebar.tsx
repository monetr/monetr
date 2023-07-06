/* eslint-disable max-len */
import React from 'react';
import { Logout, PlusOne, Settings } from '@mui/icons-material';

import BankSidebarItem from './BankSidebarItem';

import { Logo } from 'assets';
import MDivider from 'components/MDivider';
import { useLinksSink } from 'hooks/links';


export default function BankSidebar(): JSX.Element {
  // Important things to note. The width is 16. The width of the icons is 12.
  // This leaves a padding of 2 on each side, which isn't even needed with items-center? Not sure which
  // would be better.
  // py-2 pushes the icons down the same distance they are from the side.
  // gap-2 makes sure they are evenly spaced.
  const { result: links, isLoading, isError } = useLinksSink();
  if (isLoading) {
    return null;
  }
  if (isError) {
    return <div>Error!</div>;
  }

  return (
    <div className='hidden lg:visible w-16 h-full lg:flex items-center lg:py-4 gap-4 lg:flex-col dark:border-r-dark-monetr-border flex-none border border-transparent'>
      <div className='h-10 w-10'>
        <img src={ Logo } className="w-full" />
      </div>
      <MDivider className='w-1/2' />
      <div className='h-full w-full flex items-center flex-col overflow-y-auto'>
        { Array.from(links.values()).map(link => (<BankSidebarItem key={ link.linkId } link={ link } />)) }
        <div className='w-full h-12 flex items-center justify-center relative group'>
          <div className='cursor-pointer absolute rounded-full w-10 h-10 dark:bg-dark-monetr-background-subtle dark:hover:bg-dark-monetr-background-emphasis drop-shadow-md flex justify-center items-center'>
            <PlusOne className='text-3xl' />
          </div>
        </div>
      </div>
      <Settings className='dark:hover:text-dark-monetr-content-emphasis dark:text-dark-monetr-content-subtle cursor-pointer' />
      <Logout className='dark:hover:text-dark-monetr-content-emphasis dark:text-dark-monetr-content-subtle cursor-pointer' />
    </div>
  );
}
