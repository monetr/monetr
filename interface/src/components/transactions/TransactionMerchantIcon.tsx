import MerchantIcon, { type MerchantIconProps } from '@monetr/interface/components/MerchantIcon';
import { Tooltip, TooltipContent, TooltipTrigger } from '@monetr/interface/components/Tooltip';
import TransactionRecurringIndicator from '@monetr/interface/components/transactions/TransactionRecurringIndicator';
import type { ID } from '@monetr/interface/models/ID';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';

import styles from './TransactionMerchantIcon.module.scss';

export interface TransactionMerchantIconProps extends MerchantIconProps {
  pending?: boolean;
  transactionRecurringId?: ID<TransactionRecurring> | null;
}

export default function TransactionMerchantIcon(props: TransactionMerchantIconProps): React.JSX.Element {
  const { pending, transactionRecurringId, ...merchantIconProps } = props;

  return (
    <div className={styles.root}>
      <MerchantIcon {...merchantIconProps} />
      {transactionRecurringId && <TransactionRecurringIndicator transactionRecurringId={transactionRecurringId} />}
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
