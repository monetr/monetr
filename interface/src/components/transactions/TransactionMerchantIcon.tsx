import { Repeat } from 'lucide-react';

import MerchantIcon, { type MerchantIconProps } from '@monetr/interface/components/MerchantIcon';
import { Tooltip, TooltipContent, TooltipTrigger } from '@monetr/interface/components/Tooltip';

import styles from './TransactionMerchantIcon.module.scss';

export interface TransactionMerchantIconProps extends MerchantIconProps {
  pending?: boolean;
  recurring?: boolean;
}

export default function TransactionMerchantIcon(props: TransactionMerchantIconProps): React.JSX.Element {
  const { pending, recurring, ...merchantIconProps } = props;

  return (
    <div className={styles.root}>
      <MerchantIcon {...merchantIconProps} />
      {recurring && (
        <Tooltip delayDuration={100}>
          <TooltipTrigger asChild>
            <Repeat aria-label='Transaction is recurring' className={styles.recurringIndicator} />
          </TooltipTrigger>
          <TooltipContent side='right'>Transaction is recurring</TooltipContent>
        </Tooltip>
      )}
      {pending && (
        <Tooltip delayDuration={100}>
          <TooltipTrigger asChild>
            <span aria-label='Transaction is pending' className={styles.pendingIndicator} role='img'>
              <span className={styles.pendingPing} />
              <span className={styles.pendingDot} />
            </span>
          </TooltipTrigger>
          <TooltipContent side='right'>Transaction is pending</TooltipContent>
        </Tooltip>
      )}
    </div>
  );
}
