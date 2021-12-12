import { Card, Divider, List, ListSubheader, Typography } from '@mui/material';
import TransactionItem from 'components/Transactions/TransactionItem';
import { useSnackbar } from 'notistack';
import React, { Fragment, useEffect } from 'react';
import { useSelector } from 'react-redux';
import useFetchInitialTransactionsIfNeeded from 'shared/transactions/actions/fetchInitialTransactionsIfNeeded';
import { getTransactions } from 'shared/transactions/selectors/getTransactions';

import './styles/TransactionsView.scss';
import useMountEffect from 'shared/util/useMountEffect';

function TransactionsView(): JSX.Element {
  const { enqueueSnackbar } = useSnackbar();
  const fetchInitialTransactionsIfNeeded = useFetchInitialTransactionsIfNeeded();
  const transactions = useSelector(getTransactions);

  useMountEffect(() => {
    fetchInitialTransactionsIfNeeded()
      .catch(() => enqueueSnackbar('Failed to retrieve transactions.', { variant: 'error' }))
  });

  function renderTransactions() {
    return transactions
      .groupBy(transaction => transaction.date.format('MMMM Do'))
      .map((transactions, group) => (
        <li key={ group }>
          <ul>
            <Fragment>
              <ListSubheader className="bg-white pl-0 pr-0 pt-1 bg-gray-50">
                <Typography className="ml-6 font-semibold opacity-75 text-base">{ group }</Typography>
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

  return (
    <div className="minus-nav">
      <div className="flex flex-col h-full p-10 max-h-full">
        <div className="grid grid-cols-3 gap-4 flex-grow">
          <div className="col-span-3">
            <Card elevation={ 4 } className="w-full transaction-list">
              <List disablePadding className="w-full">
                { renderTransactions() }
              </List>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
}

export default TransactionsView;
