import React, { Fragment } from 'react';
import useInfiniteScroll from 'react-infinite-scroll-hook';
import { Divider, List, ListSubheader, Typography } from '@mui/material';
import moment, { Moment } from 'moment';
import * as R from 'ramda';

import TransactionItem from 'components/Transactions/TransactionsView/TransactionItem';
import { useTransactionsSink } from 'hooks/transactions';
import Transaction from 'models/Transaction';

function TransactionsView(): JSX.Element {
  const { isLoading, isFetching, fetchNextPage, error, result: transactions, hasNextPage } = useTransactionsSink();
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

  // formatDateHeader will just take the moment for a given transaction group and format it based on whether that day is
  // for the current year or not. If the date is the same year then it will not include the year in the suffix, if it is
  // different it will include the year.
  function formatDateHeader(moment: Moment): string {
    if (moment.year() !== new Date().getFullYear()) {
      return moment.format('MMMM Do, YYYY');
    }

    return moment.format('MMMM Do');
  }

  function renderTransactions() {
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
          <ul>
            <Fragment>
              <ListSubheader className="pl-0 pr-0 pt-1 bg-zinc-100 dark:bg-neutral-900 leading-none dark:opacity-100">
                <span className="ml-3 md:ml-3 font-semibold opacity-75 text-base text-sm text-gray-700 dark:text-gray-100">
                  { formatDateHeader(group) }
                </span>
                <Divider />
              </ListSubheader>
            </Fragment>
            { transactions.map(transaction => (
              <TransactionItem
                key={ transaction.transactionId }
                transaction={ transaction }
              />)) }
          </ul>
        </li>
      )),
    )(transactions);
  }

  function TransactionListFooter(): JSX.Element {
    let message = 'No more transactions...';
    if (loading) {
      message = 'Loading...';
    } else if (hasNextPage) {
      message = 'Load more?';
    }

    return (
      <div className="w-full flex justify-center p-5 opacity-70">
        <h1>{ message }</h1>
      </div>
    );
  }

  return (
    <div className="minus-nav">
      <div className="w-full view-area bg-white dark:bg-neutral-800">
        <List disablePadding className="w-full">
          { renderTransactions() }
          { (loading || hasNextPage) && (
            <li ref={ sentryRef }>
              <TransactionListFooter />
            </li>
          ) }
          { (!hasNextPage && !loading) && (
            <li>
              <TransactionListFooter />
            </li>
          ) }
        </List>
      </div>
    </div>
  );
}

export default TransactionsView;
