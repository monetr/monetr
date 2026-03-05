import { formatDistance } from 'date-fns';
import { CreditCard } from 'lucide-react';
import { Link } from 'react-router-dom';

import { Tooltip, TooltipContent, TooltipTrigger } from '@monetr/interface/components/Tooltip';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import { useAuthentication } from '@monetr/interface/hooks/useAuthentication';

import styles from './BankSidebarSubscriptionItem.module.scss';

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
          <Link className={styles.link} data-testid='bank-sidebar-subscription' to={path}>
            <CreditCard className={`${styles.icon} ${styles.iconTrialing}`} />
            <span className={styles.statusIndicator}>
              <span className={styles.statusPing} />
              <span className={styles.statusDot} />
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
    <Link className={styles.link} data-testid='bank-sidebar-subscription' to={path}>
      <CreditCard className={styles.icon} />
    </Link>
  );
}
