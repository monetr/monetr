/* eslint-disable max-len */
import React from 'react';
import { Link } from 'react-router-dom';
import { formatDistance } from 'date-fns';
import { CreditCard } from 'lucide-react';

import { Tooltip, TooltipContent, TooltipTrigger } from '@monetr/interface/components/Tooltip';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import { useAuthentication } from '@monetr/interface/hooks/useAuthentication';

export default function BankSidebarSubscriptionItem(): JSX.Element {
  const { data: config } = useAppConfiguration();
  const { data } = useAuthentication();
  const path = '/settings/billing';

  if (!config?.billingEnabled) {
    return null;
  }

  if (data?.isTrialing) {
    return (
      <Tooltip delayDuration={ 100 }>
        <TooltipTrigger>
          <Link to={ path } data-testid='bank-sidebar-subscription' className='relative group'>
            <CreditCard className='dark:group-hover:text-dark-monetr-content-emphasis dark:text-dark-monetr-content-subtle cursor-pointer mt-1.5' />
            <span className='absolute flex h-2 w-2 right-0 bottom-0'>
              <span className='animate-ping-slow absolute inline-flex h-full w-full rounded-full bg-yellow-400' />
              <span className='relative inline-flex rounded-full h-2 w-2 bg-yellow-500' />
            </span>
          </Link>
        </TooltipTrigger>
        <TooltipContent side='right'>
          Your trial ends in { formatDistance(data.trialingUntil, new Date())}.
        </TooltipContent>
      </Tooltip>
    );
  }

  return (
    <Link to={ path } data-testid='bank-sidebar-subscription'>
      <CreditCard className='dark:hover:text-dark-monetr-content-emphasis dark:text-dark-monetr-content-subtle cursor-pointer' />
    </Link>
  );
}
