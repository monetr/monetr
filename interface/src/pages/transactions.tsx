import { Fragment, useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react';
import { useWindowVirtualizer, type VirtualItem } from '@tanstack/react-virtual';
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
import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import { useTransactions } from '@monetr/interface/hooks/useTransactions';
import { showNewTransactionModal } from '@monetr/interface/modals/NewTransactionModal';
import type Transaction from '@monetr/interface/models/Transaction';

import styles from './transactions.module.scss';

const showUploadTransactionsModal = async () =>
  await import('@monetr/interface/modals/UploadTransactions/UploadTransactionsModal').then(modal =>
    modal.showUploadTransactionsModal(),
  );

export default function Transactions(): React.JSX.Element {
  const { data: transactions, hasNextPage, isLoading, isError, isFetching, fetchNextPage } = useTransactions();

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
    rootMargin: '0px 0px 700px 0px',
  });

  const groups: Array<[string, Array<Transaction>]> = useMemo(
    () =>
      Object.entries(
        (transactions ?? []).reduce<{ [date: string]: Array<Transaction> }>((accumulator, item) => {
          // biome-ignore lint/suspicious/noAssignInExpressions: This is the cleanest way to do this group by...
          (accumulator[format(item.date, 'yyyy-MM-dd')] ??= []).push(item);
          return accumulator;
        }, {}),
      ),
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
        <Typography size='5xl'>Something isn&apos;t right...</Typography>
        <Typography size='2xl'>We weren&apos;t able to retrieve transactions at this time...</Typography>
      </div>
    );
  }

  let message = 'No more transactions...';
  if (loading) {
    message = 'Loading...';
  } else if (hasNextPage) {
    message = 'Load more?';
  }

  if (!isLoading && transactions?.length === 0) {
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
              You don&apos;t have any transactions yet...
            </Typography>
            <Typography align='center' color='subtle' size='lg'>
              <EmptyTransactionsMessage />
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
        <div className={styles.list}>
          <TransactionDateGroups groups={groups} />
          {loading && (
            <div className={styles.loadMore} ref={sentryRef}>
              <h1>{message}</h1>
            </div>
          )}
          {!loading && hasNextPage && (
            <div className={styles.loadMore} ref={sentryRef}>
              <h1>{message}</h1>
            </div>
          )}
          {!loading && !hasNextPage && (
            <div className={styles.loadMore}>
              <h1>{message}</h1>
            </div>
          )}
        </div>
      </div>
    </Fragment>
  );
}

interface TransactionDateGroupsProps {
  groups: Array<[string, Array<Transaction>]>;
}

// The measured height of each day for every bank account, kept around after the list goes away. Going back to the list
// relies on the browser restoring the scroll position, which only lands in the right spot if the list is the same
// height it was when they left. Without this every day would go back to its estimated height and they would end up
// somewhere else.
const measuredDays = new Map<string, Array<VirtualItem>>();

// Where the page was scrolled to the last time they were on this list, stored on the history entry itself so it only
// comes back when they go back to that exact entry. The browser does restore the scroll position on its own, but it
// does that after the list has already rendered. By then the list has rendered the days for wherever the page was
// scrolled before, so they would see an empty gap until the next scroll event caught the list up.
const scrollOffsetStateKey = 'transactionsScrollOffset';

function getSavedScrollOffset(): number | undefined {
  const offset = window.history.state?.[scrollOffsetStateKey];
  return typeof offset === 'number' ? offset : undefined;
}

// Only the days that are on (or near) the screen are rendered, and each one is measured after it renders so nothing
// here needs to know how tall a transaction is.
function TransactionDateGroups({ groups }: TransactionDateGroupsProps): React.JSX.Element {
  // The virtualizer keeps the same instance between renders while the items it gives back change as you scroll. If the
  // React Compiler memoizes against that instance the list gets stuck on whatever it rendered first.
  'use no memo';

  const selectedBankAccountId = String(useSelectedBankAccountId());
  const listRef = useRef<HTMLDivElement>(null);
  const [scrollMargin, setScrollMargin] = useState(0);
  // Only read when the list first mounts, after that the page's actual scroll position is what matters.
  const [savedScrollOffset] = useState(getSavedScrollOffset);

  // The page scrolls, not the list, so the virtualizer needs to know how far down the page the list starts.
  useLayoutEffect(() => {
    if (listRef.current) {
      setScrollMargin(listRef.current.getBoundingClientRect().top + window.scrollY);
    }
  }, []);

  // Key by the date so days keep their measured height as more pages are loaded.
  const getItemKey = useCallback((index: number) => groups[index]?.[0] ?? index, [groups]);
  const virtualizer = useWindowVirtualizer({
    count: groups.length,
    getItemKey,
    // Just a rough guess for days that haven't rendered yet, the real height is measured once they do. If this gets out
    // of date nothing breaks, the scrollbar is just a bit off until those days render. Guessing high is better than
    // low.
    estimateSize: index => 40 + (groups[index]?.[1].length ?? 1) * 72,
    overscan: 2,
    scrollMargin,
    initialMeasurementsCache: measuredDays.get(selectedBankAccountId) ?? [],
    // The virtualizer scrolls the page to this offset when it mounts, before anything is painted, so the first frame
    // already has the right days rendered in the right spot.
    initialOffset: () => savedScrollOffset ?? window.scrollY,
  });
  const items = virtualizer.getVirtualItems();

  useEffect(() => {
    return () => {
      measuredDays.set(selectedBankAccountId, virtualizer.takeSnapshot());
    };
  }, [selectedBankAccountId, virtualizer]);

  // This cant be saved when the list unmounts, by then the history has already moved on to the next page. Instead it is
  // saved shortly after they stop scrolling. Safari throws if the history state is replaced too often, so this can't be
  // done on every scroll event either.
  useEffect(() => {
    let timeout: number | undefined;
    const onScroll = () => {
      window.clearTimeout(timeout);
      timeout = window.setTimeout(() => {
        window.history.replaceState({ ...window.history.state, [scrollOffsetStateKey]: window.scrollY }, '');
      }, 100);
    };
    window.addEventListener('scroll', onScroll, { passive: true });
    return () => {
      window.removeEventListener('scroll', onScroll);
      window.clearTimeout(timeout);
    };
  }, []);

  return (
    <div ref={listRef} style={{ height: virtualizer.getTotalSize() }}>
      <ul style={{ paddingTop: (items[0]?.start ?? scrollMargin) - scrollMargin }}>
        {items.map(item => {
          const group = groups[item.index];
          if (!group) {
            return null;
          }

          const [date, transactions] = group;
          return (
            <li data-index={item.index} key={item.key} ref={virtualizer.measureElement}>
              <ul className={styles.dateGroup}>
                <TransactionDateItem date={parse(date, 'yyyy-MM-dd', new Date())} />
                {transactions.map(transaction => (
                  <TransactionItem key={transaction.transactionId} transaction={transaction} />
                ))}
              </ul>
            </li>
          );
        })}
      </ul>
    </div>
  );
}

function EmptyTransactionsMessage(): React.JSX.Element {
  const { data: config } = useAppConfiguration();
  const { data: link } = useCurrentLink();

  if (link?.getIsManual()) {
    if (config?.uploadsEnabled) {
      return (
        <Fragment>
          This is a manual account, so transactions won&apos;t sync automatically. Add one with the + button, or upload
          a file exported from your bank to bring in a batch at once.
        </Fragment>
      );
    }

    return (
      <Fragment>
        This is a manual account, so transactions won&apos;t sync automatically. Add one with the + button to get
        started.
      </Fragment>
    );
  }

  if (link?.getIsPlaid()) {
    return (
      <Fragment>
        Transactions will show up here once Plaid sends them over, which can take a little while after you first connect
        your bank. If nothing shows up after that, this account might not provide transaction data through Plaid.
      </Fragment>
    );
  }

  if (link?.getIsLunchFlow()) {
    return (
      <Fragment>
        Transactions will show up here once Lunch Flow sends them over. If nothing shows up after the next sync, this
        account might not provide transaction data through Lunch Flow.
      </Fragment>
    );
  }

  return <Fragment>Transactions will show up here once they&apos;ve been added or synced for this account.</Fragment>;
}

function AddTransactionButton(): React.JSX.Element | null {
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

function UploadButtonMaybe(): React.JSX.Element | null {
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
