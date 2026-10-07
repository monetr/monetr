import { Fragment } from 'react';
import { HeartCrack, Repeat } from 'lucide-react';
import { useParams } from 'wouter';

import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import RecurringChargeList from '@monetr/interface/components/recurring/RecurringChargeList';
import RecurringFundingCard from '@monetr/interface/components/recurring/RecurringFundingCard';
import RecurringSpendingCard from '@monetr/interface/components/recurring/RecurringSpendingCard';
import RecurringSummaryCard from '@monetr/interface/components/recurring/RecurringSummaryCard';
import Typography from '@monetr/interface/components/Typography';
import { useRecurringTransaction } from '@monetr/interface/hooks/useRecurringTransaction';
import { useTransactionCluster } from '@monetr/interface/hooks/useTransactionCluster';
import type { ID } from '@monetr/interface/models/ID';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';

import styles from './details.module.scss';

export default function RecurringDetails(): React.JSX.Element | null {
  const { transactionRecurringId } = useParams<{ transactionRecurringId: ID<TransactionRecurring> }>();
  const { data: recurring, isLoading, isError } = useRecurringTransaction(transactionRecurringId);
  // The recurring transaction doesn't have a name of its own, it uses the name of its similar transactions group.
  const { data: cluster } = useTransactionCluster(recurring?.transactionClusterId ?? null);

  if (!transactionRecurringId) {
    return (
      <div className={styles.centerState}>
        <HeartCrack className={styles.errorIcon} />
        <Typography size='5xl'>Something isn&apos;t right...</Typography>
        <Typography size='2xl'>There wasn&apos;t a recurring transaction specified...</Typography>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className={styles.centerState}>
        <Typography size='5xl'>One moment...</Typography>
      </div>
    );
  }

  if (isError || !recurring) {
    return (
      <div className={styles.centerState}>
        <HeartCrack className={styles.errorIcon} />
        <Typography size='5xl'>Something isn&apos;t right...</Typography>
        <Typography size='2xl'>Couldn&apos;t find the recurring transaction you specified...</Typography>
      </div>
    );
  }

  const name = cluster?.name ?? '';

  // Going back should land on the tab this shows up under, not always charges
  let tab = 'charges';
  if (recurring.ended) {
    tab = 'ended';
  } else if (recurring.direction === 'credit') {
    tab = 'deposits';
  }

  return (
    <Fragment>
      <MTopNavigation
        base={`/bank/${recurring.bankAccountId}/recurring?tab=${tab}`}
        breadcrumb={name}
        icon={Repeat}
        title='Recurring'
      />
      <div className={styles.body}>
        <div className={styles.content}>
          <RecurringSummaryCard recurring={recurring} />
          {recurring.direction === 'debit' && <RecurringSpendingCard name={name} recurring={recurring} />}
          {recurring.direction === 'credit' && <RecurringFundingCard name={name} recurring={recurring} />}
          <RecurringChargeList recurring={recurring} />
        </div>
      </div>
    </Fragment>
  );
}
