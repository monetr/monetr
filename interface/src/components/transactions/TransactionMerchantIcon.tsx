import MerchantIcon, { type MerchantIconProps } from '@monetr/interface/components/MerchantIcon';
import { Tooltip, TooltipContent, TooltipTrigger } from '@monetr/interface/components/Tooltip';

import styles from './TransactionMerchantIcon.module.scss';

export interface TransactionMerchantIconProps extends MerchantIconProps {
  pending?: boolean;
}

export default function TransactionMerchantIcon(props: TransactionMerchantIconProps): React.JSX.Element {
  const { pending, ...merchantIconProps } = props;

  if (pending) {
    return (
      <div className={styles.root}>
        <MerchantIcon {...merchantIconProps} />
        <Tooltip delayDuration={100}>
          <TooltipTrigger asChild>
            <span aria-label='Transaction is pending' className={styles.pendingIndicator} role='img'>
              <span className={styles.pendingPing} />
              <span className={styles.pendingDot} />
            </span>
          </TooltipTrigger>
          <TooltipContent>Transaction is pending</TooltipContent>
        </Tooltip>
      </div>
    );
  }

  return <MerchantIcon {...merchantIconProps} />;
}
