import { Divider, List, ListSubheader, Typography } from '@mui/material';
import TransactionItem from 'components/Transactions/TransactionItem';
import { Moment } from 'moment';
import { useSnackbar } from 'notistack';
import React, { Fragment, useState } from 'react';
import useInfiniteScroll from 'react-infinite-scroll-hook';
import { useSelector } from 'react-redux';
import useFetchInitialTransactionsIfNeeded from 'shared/transactions/actions/fetchInitialTransactionsIfNeeded';
import useFetchTransactions from 'shared/transactions/hooks/useFetchTransactions';
import { getTransactions } from 'shared/transactions/selectors/getTransactions';
import useMountEffect from 'shared/util/useMountEffect';

import 'components/Transactions/TransactionsView/styles/TransactionsView.scss';

function TransactionsView(): JSX.Element {
  const { enqueueSnackbar } = useSnackbar();
  const fetchInitialTransactionsIfNeeded = useFetchInitialTransactionsIfNeeded();
  const fetchTransactions = useFetchTransactions();
  const transactions = useSelector(getTransactions);

  useMountEffect(() => {
    fetchInitialTransactionsIfNeeded()
      .catch(() => enqueueSnackbar('Failed to retrieve transactions.', { variant: 'error' }))
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  // TODO This is a temp approach, since we don't know how many transactions we have for a given bank account, we can
  //  just request transactions until we get a page that is not full. But this is not a good way to do it. If they have
  //  a total number of transactions divisible by 25 then we could continue to try to request more.
  const hasNextPage = transactions.count() % 25 === 0;

  function retrieveMoreTransactions() {
    setLoading(true);
    return fetchTransactions(transactions.count())
      .catch(error => {
        enqueueSnackbar('Failed to retrieve more transactions.', {
          variant: 'error',
          disableWindowBlurListener: true,
        });
        setError(error)
      })
      .finally(() => setLoading(false))
  }

  const [sentryRef] = useInfiniteScroll({
    loading,
    hasNextPage,
    onLoadMore: retrieveMoreTransactions,
    // When there is an error, we stop infinite loading.
    // It can be reactivated by setting "error" state as undefined.
    disabled: !!error,
    // `rootMargin` is passed to `IntersectionObserver`.
    // We can use it to trigger 'onLoadMore' when the sentry comes near to become
    // visible, instead of becoming fully visible on the screen.
    rootMargin: '0px 0px 400px 0px',
  });

  // formatDateHeader will just take the moment for a given transaction group and format it based on whether that day is
  // for the current year or not. If the date is the same year then it will not include the year in the suffix, if it is
  // different it will include the year.
  function formatDateHeader(moment: Moment): string {
    if (moment.year() !== new Date().getFullYear()) {
      return moment.format('MMMM Do, YYYY')
    }

    return moment.format('MMMM Do')
  }

  function renderTransactions() {
    return transactions
      // TODO Right now transactions don't have a "time", only a date. So is this a good group by key? Could we ever
      //  have two transactions for the same day that are not the same date object?
      .groupBy(transaction => transaction.date)
      .map((transactions, group) => (
        <li key={ group.unix() }>
          <ul>
            <Fragment>
              <ListSubheader className="bg-white pl-0 pr-0 pt-1 bg-gray-50">
                <Typography className="ml-6 font-semibold opacity-75 text-base">
                  { formatDateHeader(group) }
                </Typography>
                <Divider/>
              </ListSubheader>
            </Fragment>
            { transactions.map(transaction => (
              <TransactionItem
                key={ transaction.transactionId }
                transactionId={ transaction.transactionId }
              />)).valueSeq().toArray() }
          </ul>
        </li>
      ))
      .valueSeq()
      .toArray();
  }

  function TransactionListFooter(): JSX.Element {
    let message = 'No more transactions...'
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
      <div className="w-full transaction-list bg-white">
        <List disablePadding className="w-full">
          { renderTransactions() }
          { (loading || hasNextPage) && (
            <li ref={ sentryRef }>
              <TransactionListFooter/>
            </li>
          ) }
          { (!hasNextPage && !loading) && (
            <li>
              <TransactionListFooter/>
            </li>
          ) }
        </List>
      </div>
    </div>
  );
}

export default TransactionsView;
