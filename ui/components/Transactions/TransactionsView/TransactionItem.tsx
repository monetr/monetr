import { AccessTime } from '@mui/icons-material';
import { Chip, Divider, ListItem } from '@mui/material';
import classnames from 'classnames';
import TransactionNameEditor from 'components/Transactions/TransactionsView/TransactionNameEditor';
import TransactionSpentFromSelection from 'components/Transactions/TransactionsView/TransactionSpentFromSelection';
import Transaction from 'models/Transaction';
import React, { Fragment } from 'react';

import 'components/Transactions/TransactionsView/styles/TransactionItem.scss';

interface Props {
  transaction: Transaction;
}

export default function TransactionItem(props: Props): JSX.Element {
  return (
    <Fragment>
      <ListItem className="flex flex-row transactions-item pl-1 pr-1 md:pr-2.5">
        <div className="flex flex-col md:flex-row basis-9/12 md:basis-10/12">
          <TransactionNameEditor transaction={ props.transaction }/>
          <TransactionSpentFromSelection transaction={ props.transaction }/>
        </div>
        <div className="basis-3/12 md:basis-2/12 flex justify-end w-full items-center">
          { props.transaction.isPending && <Chip icon={<AccessTime />} label="Pending" className="mr-auto"/> }
          <span className={ classnames('h-full amount align-middle self-center place-self-center', {
            'addition': props.transaction.getIsAddition(),
          }) }>
            <b>{ props.transaction.getAmountString() }</b>
          </span>
        </div>
      </ListItem>
      <Divider/>
    </Fragment>
  );
}
