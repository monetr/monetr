import React, { Fragment, useCallback, useEffect, useRef } from 'react';
import useInfiniteScroll from 'react-infinite-scroll-hook';
import { useNavigationType } from 'react-router-dom';
import { HeartBroken, ShoppingCartOutlined } from '@mui/icons-material';
import { format, getUnixTime, parse } from 'date-fns';
import * as R from 'ramda';

import MSpan from '@monetr/interface/components/MSpan';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import { useTransactions } from '@monetr/interface/hooks/transactions';
import Transaction from '@monetr/interface/models/Transaction';
import TransactionDateItem from '@monetr/interface/pages/new/TransactionDateItem';
import TransactionItem from '@monetr/interface/pages/new/TransactionItem';

let evilScrollPosition: number = 0;

export default function Transactions(): JSX.Element {
  const {
    isLoading,
    isError,
    isFetching,
    fetchNextPage,
    result: transactions, hasNextPage,
  } = useTransactions();

  // Scroll restoration code.
  const ref = useRef<HTMLUListElement>(null);
  const navigationType = useNavigationType();
  const onScroll = useCallback(() => {
    evilScrollPosition = ref.current.scrollTop;
  }, [ref]);
  useEffect(() => {
    if (!ref.current) {
      return undefined;
    }

    if (navigationType === 'POP') {
      ref.current.scrollTop = evilScrollPosition;
    }
    const current = ref.current;
    ref.current.addEventListener('scroll', onScroll);
    return () => {
      current.removeEventListener('scroll', onScroll);
    };
  // Fix bug with current impl.
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ref.current, navigationType, onScroll]);

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
    rootMargin: '0px 0px 0px 0px',
  });

  // Uncomment this to make it so that transaction data is removed from memory upon navigating away.
  // useEffect(() => {
  //   return remove;
  // }, [remove]);

  if (isLoading) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <MSpan className='text-5xl'>
          One moment...
        </MSpan>
      </div>
    );
  }

  if (isError) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <HeartBroken className='dark:text-dark-monetr-content h-24 w-24' />
        <MSpan className='text-5xl'>
          Something isn't right...
        </MSpan>
        <MSpan className='text-2xl'>
          We weren't able to retrieve transactions at this time...
        </MSpan>
      </div>
    );
  }

  function TransactionItems() {
    interface TransactionGroup {
      transactions: Array<Transaction>;
      group: Date;
    }
    return R.pipe(
      R.groupBy((item: Transaction) => format(item.date, 'yyyy-MM-dd')),
      R.mapObjIndexed((transactions, date) => ({
        transactions: transactions,
        group: parse(date, 'yyyy-MM-dd', new Date()),
      })),
      R.values,
      R.map(({ transactions, group }: TransactionGroup): JSX.Element => (
        <li key={ getUnixTime(group) }>
          <ul className='flex gap-2 flex-col'>
            <TransactionDateItem date={ group } />
            { transactions .map(transaction => (
              <TransactionItem key={ transaction.transactionId } transaction={ transaction } />
            )) }
          </ul>
        </li>
      )),
    )(transactions);
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
        <MTopNavigation
          icon={ ShoppingCartOutlined }
          title='Transactions'
        />
        <div className='w-full h-full flex justify-center items-center'>
          <div className='flex flex-col gap-2 items-center max-w-md'>
            <div className='w-full flex justify-center space-x-4'>
              <ShoppingCartOutlined className='h-full text-5xl dark:text-dark-monetr-content-muted' />
            </div>
            <MSpan size='xl' color='subtle' className='text-center'>
              You don't have any transactions yet...
            </MSpan>
            <MSpan size='lg' color='subtle' className='text-center'>
              Transactions will show up here once we receive them from Plaid. Or the current account might not support
              transaction data from Plaid.
            </MSpan>
          </div>
        </div>
      </Fragment>
    );
  }

  return (
    <Fragment>
      <MTopNavigation
        icon={ ShoppingCartOutlined }
        title='Transactions'
      />
      <div className='flex flex-grow min-w-0 min-h-0'>
        <ul className='w-full overflow-y-auto' ref={ ref }>
          <TransactionItems />
          {loading && (
            <li ref={ sentryRef }>
              <div className='w-full flex justify-center p-5 opacity-70'>
                <h1>{message}</h1>
              </div>
            </li>
          )}
          {(!loading && hasNextPage) && (
            <li ref={ sentryRef }>
              <div className='w-full flex justify-center p-5 opacity-70'>
                <h1>{message}</h1>
              </div>
            </li>
          )}
          {(!loading && !hasNextPage) && (
            <li>
              <div className='w-full flex justify-center p-5 opacity-70'>
                <h1>{message}</h1>
              </div>
            </li>
          )}
        </ul>
      </div>
    </Fragment>
  );
}
