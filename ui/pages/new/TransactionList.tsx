import React, { Fragment } from 'react';
import useInfiniteScroll from 'react-infinite-scroll-hook';
import { ShoppingCartOutlined } from '@mui/icons-material';
import moment from 'moment';
import * as R from 'ramda';

import TransactionDateItem from './TransactionDateItem';
import TransactionItem from './TransactionItem';

import MSidebarToggle from 'components/MSidebarToggle';
import { useTransactions } from 'hooks/transactions';
import Transaction from 'models/Transaction';

export default function TransactionList(): JSX.Element {
  const { isLoading, isFetching, fetchNextPage, error, result: transactions, hasNextPage } = useTransactions();
  const loading = isLoading || isFetching;

  const [sentryRef] = useInfiniteScroll({
    loading,
    hasNextPage,
    onLoadMore: fetchNextPage,
    // When there is an error, we stop infinite loading.
    // It can be reactivated by setting "error" state as undefined.
    disabled: !!error,
    // `rootMargin` is passed to `IntersectionObserver`.
    // We can use it to trigger 'onLoadMore' when the sentry comes near to become
    // visible, instead of becoming fully visible on the screen.
    rootMargin: '0px 0px 0px 0px',
  });

  function TransactionItems() {
    interface TransactionGroup {
      transactions: Array<Transaction>;
      group: moment.Moment;
    }
    return R.pipe(
      R.groupBy((item: Transaction) => item.date.toString()),
      R.mapObjIndexed((transactions, date) => ({
        transactions: transactions,
        group: moment(date),
      })),
      R.values,
      R.map(({ transactions, group }: TransactionGroup): JSX.Element => (
        <li key={ group.unix() }>
          <ul className='flex gap-2 flex-col'>
            <TransactionDateItem date={ group } />
            {
              transactions.map(transaction =>
                (<TransactionItem key={ transaction.transactionId } transaction={ transaction } />))
            }
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

  return (
    <Fragment>
      <div className='w-full h-12 flex-none flex items-center px-4 gap-4'>
        <MSidebarToggle />
        <span className='text-2xl dark:text-dark-monetr-content-emphasis font-bold flex gap-2 items-center'>
          <ShoppingCartOutlined />
          Transactions
        </span>
        <span className='hidden md:block'>
          md
        </span>
        <span className='hidden lg:block'>
          lg
        </span>
        <span className='hidden xl:block'>
          xl
        </span>
      </div>
      <div className='flex flex-grow min-w-0 min-h-0'>
        <ul className='w-full overflow-y-auto'>
          <TransactionItems />
          {loading && (
            <li ref={ sentryRef }>
              <div className="w-full flex justify-center p-5 opacity-70">
                <h1>{message}</h1>
              </div>
            </li>
          )}
          {(!loading && hasNextPage) && (
            <li ref={ sentryRef }>
              <div className="w-full flex justify-center p-5 opacity-70">
                <h1>{message}</h1>
              </div>
            </li>
          )}
          {(!loading && !hasNextPage) && (
            <li>
              <div className="w-full flex justify-center p-5 opacity-70">
                <h1>{message}</h1>
              </div>
            </li>
          )}
        </ul>
      </div>
    </Fragment>
  );
}
