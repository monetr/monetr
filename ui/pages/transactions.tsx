import React, { Fragment } from 'react';
import useInfiniteScroll from 'react-infinite-scroll-hook';
import { ShoppingCartOutlined } from '@mui/icons-material';
import * as R from 'ramda';

import TransactionDateItem from 'pages/new/TransactionDateItem';
import TransactionItem from 'pages/new/TransactionItem';

import MTopNavigation from 'components/MTopNavigation';
import { format, getUnixTime, parse } from 'date-fns';
import { useTransactions } from 'hooks/transactions';
import Transaction from 'models/Transaction';

export default function Transactions(): JSX.Element {
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

  // Uncomment this to make it so that transaction data is removed from memory upon navigating away.
  // useEffect(() => {
  //   return remove;
  // }, [remove]);

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
      <MTopNavigation
        icon={ ShoppingCartOutlined }
        title='Transactions'
      />
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
