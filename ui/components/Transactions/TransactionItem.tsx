import { Chip, Divider, ListItem, Typography } from '@mui/material';
import classnames from 'classnames';
import TransactionNameEditor from 'components/Transactions/TransactionNameEditor';
import TransactionSpentFromSelection from 'components/Transactions/TransactionSpentFromSelection';
import React, { Fragment } from 'react';
import { useSelector } from 'react-redux';
import { getTransactionById } from 'shared/transactions/selectors/getTransactionById';

import './styles/TransactionItem.scss';

interface Props {
  transactionId: number;
}

export default function TransactionItem(props: Props): JSX.Element {
  const transaction = useSelector(getTransactionById(props.transactionId));

  return (
    <Fragment>
      <ListItem className="transactions-item h-12" role="transaction-row">
        <div className="flex flex-row w-full">
          <div
            className="flex-shrink w-2/5 pr-1 font-semibold transaction-item-name place-self-center"
          >
            <TransactionNameEditor transactionId={ transaction.transactionId }/>
          </div>

          <p className="flex-auto w-2/5 pr-1 transaction-item-spending">
            <TransactionSpentFromSelection transactionId={ transaction.transactionId }/>
          </p>
          <div className="flex items-center flex-none w-1/5">
            { transaction.isPending && <Chip label="Pending" className="self-center align-middle"/> }
            <div className="flex justify-end w-full">
              <Typography className={ classnames('amount align-middle self-center place-self-center', {
                'addition': transaction.getIsAddition(),
              }) }>
                <b>{ transaction.getAmountString() }</b>
              </Typography>
            </div>
          </div>
        </div>
      </ListItem>
      <Divider/>
    </Fragment>
  )
}
