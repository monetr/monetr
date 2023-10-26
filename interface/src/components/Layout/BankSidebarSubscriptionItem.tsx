/* eslint-disable max-len */
import React from 'react';
import { Link } from 'react-router-dom';
import styled from '@emotion/styled';
import { CreditCard } from '@mui/icons-material';
import { Badge, Tooltip } from '@mui/material';

import { differenceInDays } from 'date-fns';
import { useAppConfiguration } from 'hooks/useAppConfiguration';
import { useAuthenticationSink } from 'hooks/useAuthentication';
import useTheme from 'hooks/useTheme';

export default function BankSidebarSubscriptionItem(): JSX.Element {
  const config = useAppConfiguration();
  const theme = useTheme();
  const { result } = useAuthenticationSink();
  const path = '/settings/billing';

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
      <Link to={ path } data-testid='bank-sidebar-subscription'>
        <Tooltip
          title={ `Your trial ends in ${ differenceInDays(result.trialingUntil, new Date())} day(s).` }
          arrow
          placement='right'
          classes={ {
            tooltip: 'text-base font-medium',
          } }
        >
          <StyledBadge
            overlap='circular'
            anchorOrigin={ { vertical: 'top', horizontal: 'right' } }
            badgeContent={ differenceInDays(result.trialingUntil, new Date()) }
            classes={ { badge: 'left-1 top-1 text-[11px]' } }
          >
            <CreditCard className='dark:hover:text-dark-monetr-content-emphasis dark:text-dark-monetr-content-subtle cursor-pointer mt-1.5' />
          </StyledBadge>
        </Tooltip>
      </Link>
    );
  }

  return (
    <Link to={ path } data-testid='bank-sidebar-subscription'>
      <CreditCard className='dark:hover:text-dark-monetr-content-emphasis dark:text-dark-monetr-content-subtle cursor-pointer' />
    </Link>
  );
}
