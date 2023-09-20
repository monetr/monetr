/* eslint-disable max-len */
import React from 'react';
import { Link } from 'react-router-dom';
import styled from '@emotion/styled';
import { CreditCard, ErrorOutline, Logout, PlusOne, Settings } from '@mui/icons-material';
import { Badge, Tooltip } from '@mui/material';

import { Logo } from 'assets';
import BankSidebarItem from 'components/Layout/BankSidebarItem';
import MDivider from 'components/MDivider';
import MSidebarToggle from 'components/MSidebarToggle';
import { ReactElement } from 'components/types';
import { differenceInDays } from 'date-fns';
import { useLinks } from 'hooks/links';
import { useAppConfiguration } from 'hooks/useAppConfiguration';
import { useAuthenticationSink } from 'hooks/useAuthentication';
import useTheme from 'hooks/useTheme';
import mergeTailwind from 'util/mergeTailwind';

export interface BankSidebarProps {
  className?: string;
}

export default function BankSidebar(props: BankSidebarProps): JSX.Element {
  // Important things to note. The width is 16. The width of the icons is 12.
  // This leaves a padding of 2 on each side, which isn't even needed with items-center? Not sure which
  // would be better.
  // py-2 pushes the icons down the same distance they are from the side.
  // gap-2 makes sure they are evenly spaced.
  const { data: links, isLoading, isError } = useLinks();
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

  // TODO Make it so that when we are in the "add link" page, we have the add link +1 button as active.
  return (
    <SidebarWrapper className={ props.className }>
      { Array.from(links.values()).map(link => (<BankSidebarItem key={ link.linkId } link={ link } />)) }
      <div className='w-full h-12 flex items-center justify-center relative group'>
        <Link
          to="/link/create"
          className='cursor-pointer absolute rounded-full w-10 h-10 dark:bg-dark-monetr-background-subtle dark:hover:bg-dark-monetr-background-emphasis drop-shadow-md flex justify-center items-center'
        >
          <PlusOne className='text-3xl' />
        </Link>
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
    <div className={ className } data-testid='bank-sidebar'>
      <MSidebarToggle className='flex lg:hidden' />
      <div className='h-10 w-10'>
        <img src={ Logo } className="w-full" />
      </div>
      <MDivider className='w-1/2' />
      <div className='h-full w-full flex items-center flex-col overflow-y-auto'>
        { props?.children }
      </div>
      <SubscriptionButton />
      <SettingsButton />
      <LogoutButton />
    </div>
  );
}

function SubscriptionButton(): JSX.Element {
  const config = useAppConfiguration();
  const theme = useTheme();
  const { result } = useAuthenticationSink();

  if (!config?.billingEnabled) {
    return null;
  }

  const StyledBadge = styled(Badge)(() => ({
    '& .MuiBadge-badge': {
      opacity: '90%',
      backgroundColor: theme.tailwind.colors['yellow']['600'],
      boxShadow: `0 0 0 2px ${theme.tailwind.colors['dark-monetr']['background']['DEFAULT']}`,
      '&::after': {
        color:  theme.tailwind.colors['yellow']['600'],
        position: 'absolute',
        top: 0,
        left: 0,
        width: '100%',
        height: '100%',
        borderRadius: '100%',
        animation: 'ripple-trial 3s infinite ease-in-out',
        border: '1px solid currentColor',
        content: '""',
      },
    },
    '@keyframes ripple-trial': {
      '0%': {
        transform: 'scale(.8)',
        opacity: 1,
      },
      '70%': {
        transform: 'scale(.9)',
        opacity: 1,
      },
      '100%': {
        transform: 'scale(2.4)',
        opacity: 0,
      },
    },
  }));

  if (result?.isTrialing) {
    return (
      <Link to='/subscription' data-testid='bank-sidebar-subscription'>
        <Tooltip
          title={ `Your trial ends in ${ differenceInDays(result.activeUntil, new Date())} day(s).` }
          arrow
          placement='right'
          classes={ {
            tooltip: 'text-base font-medium',
          } }
        >
          <StyledBadge
            overlap='circular'
            anchorOrigin={ { vertical: 'top', horizontal: 'right' } }
            badgeContent={ differenceInDays(result.activeUntil, new Date()) }
            classes={ { badge: 'left-1 top-1 text-[11px]' } }
          >
            <CreditCard className='dark:hover:text-dark-monetr-content-emphasis dark:text-dark-monetr-content-subtle cursor-pointer mt-1.5' />
          </StyledBadge>
        </Tooltip>
      </Link>
    );
  }

  return (
    <Link to='/subscription' data-testid='bank-sidebar-subscription'>
      <CreditCard className='dark:hover:text-dark-monetr-content-emphasis dark:text-dark-monetr-content-subtle cursor-pointer' />
    </Link>
  );
}

function SettingsButton(): JSX.Element {
  return (
    <Link to='/settings' data-testid='bank-sidebar-settings'>
      <Settings className='dark:hover:text-dark-monetr-content-emphasis dark:text-dark-monetr-content-subtle cursor-pointer' />
    </Link>
  );
}

function LogoutButton(): JSX.Element {
  // By doing reloadDocument, we are forcing the @tanstack/react-query cache to be emptied. This will naturally just make it
  // easier to prevent the current user's data from leaking into another session.
  return (
    <Link to='/logout' reloadDocument data-testid='bank-sidebar-logout'>
      <Logout className='dark:hover:text-dark-monetr-content-emphasis dark:text-dark-monetr-content-subtle cursor-pointer' />
    </Link>
  );
}