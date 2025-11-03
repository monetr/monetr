import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import type Transaction from '@monetr/interface/models/Transaction';
import { AmountType } from '@monetr/interface/util/amounts';

import styles from './TransactionAmount.module.css';

export interface TransactionAmountProps {
  transaction: Transaction;
}

export default function TransactionAmount({ transaction }: TransactionAmountProps): React.JSX.Element {
  const { data: locale } = useLocaleCurrency();
  return (
    <span className={styles.transactionAmount} data-positive={transaction.getIsAddition()}>
      {locale.formatAmount(Math.abs(transaction.amount), AmountType.Stored, transaction.amount < 0)}
    </span>
  );
}
