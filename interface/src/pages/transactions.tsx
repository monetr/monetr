import { Fragment, useMemo, useRef } from 'react';
import { format, parse } from 'date-fns';
import { HeartCrack, Plus, ShoppingCart, Upload } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import BalanceFreeToUseAmount from '@monetr/interface/components/Layout/BalanceFreeToUseAmount';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import Typography from '@monetr/interface/components/Typography';
import TransactionDateItem from '@monetr/interface/components/transactions/TransactionDateItem';
import TransactionItem from '@monetr/interface/components/transactions/TransactionItem';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import { useCurrentLink } from '@monetr/interface/hooks/useCurrentLink';
import { useInfiniteScroll } from '@monetr/interface/hooks/useInfiniteScroll';
import { useTransactions } from '@monetr/interface/hooks/useTransactions';
import { showNewTransactionModal } from '@monetr/interface/modals/NewTransactionModal';
import type Transaction from '@monetr/interface/models/Transaction';

import styles from './transactions.module.scss';

const showUploadTransactionsModal = async () =>
  await import('@monetr/interface/modals/UploadTransactions/UploadTransactionsModal').then(modal =>
    modal.showUploadTransactionsModal(),
  );

export default function Transactions(): JSX.Element {
  const { data: transactions, hasNextPage, isLoading, isError, isFetching, fetchNextPage } = useTransactions();
  const ref = useRef<HTMLUListElement>(null);

  const loading = isLoading || isFetching;

  const [sentryRef] = useInfiniteScroll({
    loading,
    hasNextPage,
    onLoadMore: fetchNextPage,
    // When there is an error, we stop infinite loading.
    // It can be reactivated by setting "error" state as undefined.
    disabled: isError,
    // `rootMargin` is passed to `IntersectionObserver`.
    // We can use it to trigger 'onLoadMore' when the sentry comes near to become
    // visible, instead of becoming fully visible on the screen.
    rootMargin: '0px 0px 400px 0px',
  });

  const groups: { [date: string]: Array<Transaction> } = useMemo(
    () =>
      (transactions ?? []).reduce((accumulator, item) => {
        // biome-ignore lint/suspicious/noAssignInExpressions: This is the cleanest way to do this group by...
        (accumulator[format(item.date, 'yyyy-MM-dd')] ??= []).push(item);
        return accumulator;
      }, {}),
    [transactions],
  );

  if (isLoading) {
    return (
      <div className={styles.centerState}>
        <Typography size='5xl'>One moment...</Typography>
      </div>
    );
  }

  if (isError) {
    return (
      <div className={styles.centerState}>
        <HeartCrack className={styles.errorIcon} />
        <Typography size='5xl'>Something isn't right...</Typography>
        <Typography size='2xl'>We weren't able to retrieve transactions at this time...</Typography>
      </div>
    );
  }

  let message = 'No more transactions...';
  if (loading) {
    message = 'Loading...';
  } else if (hasNextPage) {
    message = 'Load more?';
  }

  if (!isLoading && transactions.length === 0) {
    return (
      <Fragment>
        <MTopNavigation icon={ShoppingCart} title='Transactions'>
          <UploadButtonMaybe />
        </MTopNavigation>
        <AddTransactionButton />
        <div className={styles.empty}>
          <div className={styles.emptyInner}>
            <div className={styles.iconRow}>
              <ShoppingCart className={styles.emptyIcon} />
            </div>
            <Typography align='center' color='subtle' size='xl'>
              You don't have any transactions yet...
            </Typography>
            <Typography align='center' color='subtle' size='lg'>
              Transactions will show up here once we receive them from Plaid. Or the current account might not support
              transaction data from Plaid.
            </Typography>
          </div>
        </div>
      </Fragment>
    );
  }

  return (
    <Fragment>
      <MTopNavigation icon={ShoppingCart} title='Transactions'>
        <div className={styles.balanceRow}>
          <div className={styles.balanceSpacer} /> {/* These force the free to use to be more centered */}
          <BalanceFreeToUseAmount />
          <div className={styles.balanceSpacer} />
        </div>
        <UploadButtonMaybe />
      </MTopNavigation>
      <AddTransactionButton />
      <div className={styles.content}>
        <ul className={styles.list} ref={ref}>
          {Object.entries(groups).map(([date, transactionGroup]) => (
            <li key={date}>
              <ul className={styles.dateGroup}>
                <TransactionDateItem date={parse(date, 'yyyy-MM-dd', new Date())} />
                {transactionGroup.map(transaction => (
                  <TransactionItem key={transaction.transactionId} transaction={transaction} />
                ))}
              </ul>
            </li>
          ))}
          {loading && (
            <li ref={sentryRef}>
              <div className={styles.loadMore}>
                <h1>{message}</h1>
              </div>
            </li>
          )}
          {!loading && hasNextPage && (
            <li ref={sentryRef}>
              <div className={styles.loadMore}>
                <h1>{message}</h1>
              </div>
            </li>
          )}
          {!loading && !hasNextPage && (
            <li>
              <div className={styles.loadMore}>
                <h1>{message}</h1>
              </div>
            </li>
          )}
        </ul>
      </div>
    </Fragment>
  );
}

function AddTransactionButton(): JSX.Element {
  const { data: link } = useCurrentLink();

  if (!link?.getIsManual()) {
    return null;
  }

  return (
    <button className={styles.addButton} onClick={showNewTransactionModal} type='button'>
      <Plus className={styles.addButtonIcon} />
    </button>
  );
}

function UploadButtonMaybe(): JSX.Element {
  const { data: config } = useAppConfiguration();
  const { data: link } = useCurrentLink();
  if (!link?.getIsManual()) {
    return null;
  }

  if (!config?.uploadsEnabled) {
    return null;
  }

  return (
    <Button className={styles.uploadButton} onClick={showUploadTransactionsModal} variant='primary'>
      <Upload />
      Upload
    </Button>
  );
}
