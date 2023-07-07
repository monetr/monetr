/* eslint-disable max-len */
import React from 'react';
import { Link } from 'react-router-dom';
import { ErrorOutline, Logout, PlusOne, Settings } from '@mui/icons-material';

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
    return (
      <div className='hidden lg:visible w-16 h-full lg:flex items-center lg:py-4 gap-4 lg:flex-col dark:border-r-dark-monetr-border flex-none border border-transparent'>
        <div className='h-10 w-10'>
          <img src={ Logo } className="w-full" />
        </div>
        <MDivider className='w-1/2' />
        <div className='h-full w-full flex items-center flex-col overflow-y-auto'>
        </div>
        <SettingsButton />
        <LogoutButton />
      </div>
    );
  }

  if (isError) {
    return (
      <div className='hidden lg:visible w-16 h-full lg:flex items-center lg:py-4 gap-4 lg:flex-col dark:border-r-dark-monetr-border flex-none border border-transparent'>
        <div className='h-10 w-10'>
          <img src={ Logo } className="w-full" />
        </div>
        <MDivider className='w-1/2' />
        <div className='h-full w-full flex items-center flex-col overflow-y-auto'>
          <div className='w-full h-12 flex items-center justify-center relative group'>
            <div className='absolute rounded-full w-10 h-10 dark:bg-dark-monetr-background-subtle dark:hover:bg-dark-monetr-background-emphasis drop-shadow-md flex justify-center items-center'>
              <ErrorOutline className='text-3xl' />
            </div>
          </div>
        </div>
        <SettingsButton />
        <LogoutButton />
      </div>
    );
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
      <SettingsButton />
      <LogoutButton />
    </div>
  );
}

function SettingsButton(): JSX.Element {
  return (
    <Link to='/settings'>
      <Settings className='dark:hover:text-dark-monetr-content-emphasis dark:text-dark-monetr-content-subtle cursor-pointer' />
    </Link>
  );
}

function LogoutButton(): JSX.Element {
  // By doing reloadDocument, we are forcing the react-query cache to be emptied. This will naturally just make it
  // easier to prevent the current user's data from leaking into another session.
  return (
    <Link to='/logout' reloadDocument>
      <Logout className='dark:hover:text-dark-monetr-content-emphasis dark:text-dark-monetr-content-subtle cursor-pointer' />
    </Link>
  );
}
