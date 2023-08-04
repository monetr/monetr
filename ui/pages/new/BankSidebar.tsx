/* eslint-disable max-len */
import React from 'react';
import { Link } from 'react-router-dom';
import { ErrorOutline, Logout, PlusOne, Settings } from '@mui/icons-material';

import BankSidebarItem from './BankSidebarItem';

import { Logo } from 'assets';
import MDivider from 'components/MDivider';
import { useLinksSink } from 'hooks/links';
import { ReactElement } from 'components/types';
import mergeTailwind from 'util/mergeTailwind';
import MSidebarToggle from 'components/MSidebarToggle';

export interface BankSidebarProps {
  className?: string;
}

export default function BankSidebar(props: BankSidebarProps): JSX.Element {
  // Important things to note. The width is 16. The width of the icons is 12.
  // This leaves a padding of 2 on each side, which isn't even needed with items-center? Not sure which
  // would be better.
  // py-2 pushes the icons down the same distance they are from the side.
  // gap-2 makes sure they are evenly spaced.
  const { result: links, isLoading, isError } = useLinksSink();
  if (isLoading) {
    return (
      <SidebarWrapper className={ props.className } />
    );
  }

  if (isError) {
    return (
      <SidebarWrapper className={ props.className }>
        <div className='w-full h-12 flex items-center justify-center relative group'>
          <div className='absolute rounded-full w-10 h-10 dark:bg-dark-monetr-background-subtle dark:hover:bg-dark-monetr-background-emphasis drop-shadow-md flex justify-center items-center'>
            <ErrorOutline className='text-3xl' />
          </div>
        </div>
      </SidebarWrapper>
    );
  }

  return (
    <SidebarWrapper className={ props.className }>
      { Array.from(links.values()).map(link => (<BankSidebarItem key={ link.linkId } link={ link } />)) }
      <div className='w-full h-12 flex items-center justify-center relative group'>
        <div className='cursor-pointer absolute rounded-full w-10 h-10 dark:bg-dark-monetr-background-subtle dark:hover:bg-dark-monetr-background-emphasis drop-shadow-md flex justify-center items-center'>
          <PlusOne className='text-3xl' />
        </div>
      </div>
    </SidebarWrapper>
  );
}

interface SidebarWrapperProps {
  className?: string;
  children?: ReactElement;
}

function SidebarWrapper(props: SidebarWrapperProps): JSX.Element {
  const className = mergeTailwind(
    'border',
    'border-transparent',
    'dark:border-r-dark-monetr-border',
    'flex',
    'flex-col',
    'flex-none',
    'gap-4',
    'h-full',
    'items-center',
    'lg:py-4',
    'pt-2',
    'pb-4',
    'w-16',
    props.className,
  );

  return (
    <div className={ className }>
      <MSidebarToggle className='flex lg:hidden' />
      <div className='h-10 w-10'>
        <img src={ Logo } className="w-full" />
      </div>
      <MDivider className='w-1/2' />
      <div className='h-full w-full flex items-center flex-col overflow-y-auto'>
        { props?.children }
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
  // By doing reloadDocument, we are forcing the @tanstack/react-query cache to be emptied. This will naturally just make it
  // easier to prevent the current user's data from leaking into another session.
  return (
    <Link to='/logout' reloadDocument>
      <Logout className='dark:hover:text-dark-monetr-content-emphasis dark:text-dark-monetr-content-subtle cursor-pointer' />
    </Link>
  );
}
