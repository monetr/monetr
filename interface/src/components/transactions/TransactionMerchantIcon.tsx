import MerchantIcon, { type MerchantIconProps } from '@monetr/interface/components/MerchantIcon';

import styles from './TransactionMerchantIcon.module.scss';

export interface TransactionMerchantIconProps extends MerchantIconProps {
  pending?: boolean;
}

export default function TransactionMerchantIcon(props: TransactionMerchantIconProps): JSX.Element {
  const { pending, ...merchantIconProps } = props;

  if (pending) {
    return (
      <div className={styles.root}>
        <MerchantIcon {...merchantIconProps} />
        <span className={styles.pendingIndicator}>
          <span className={styles.pendingPing} />
          <span className={styles.pendingDot} />
        </span>
      </div>
    );
  }

  return <MerchantIcon {...merchantIconProps} />;
}
