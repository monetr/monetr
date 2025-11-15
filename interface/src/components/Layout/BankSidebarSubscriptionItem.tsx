import { formatDistance } from 'date-fns';
import { CreditCard } from 'lucide-react';
import { Link } from 'react-router-dom';

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
      <Tooltip delayDuration={100}>
        <TooltipTrigger>
          <Link className='relative group' data-testid='bank-sidebar-subscription' to={path}>
            <CreditCard className='dark:group-hover:text-dark-monetr-content-emphasis dark:text-dark-monetr-content-subtle cursor-pointer mt-1.5' />
            <span className='absolute flex h-2 w-2 right-0 bottom-0'>
              <span className='animate-ping-slow absolute inline-flex h-full w-full rounded-full bg-yellow-400' />
              <span className='relative inline-flex rounded-full h-2 w-2 bg-yellow-500' />
            </span>
          </Link>
        </TooltipTrigger>
        <TooltipContent side='right'>
          Your trial ends in {formatDistance(data.trialingUntil, new Date())}.
        </TooltipContent>
      </Tooltip>
    );
  }

  return (
    <Link data-testid='bank-sidebar-subscription' to={path}>
      <CreditCard className='dark:hover:text-dark-monetr-content-emphasis dark:text-dark-monetr-content-subtle cursor-pointer' />
    </Link>
  );
}
