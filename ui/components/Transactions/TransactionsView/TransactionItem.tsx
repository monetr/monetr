import TransactionItemMobile from 'components/Transactions/TransactionsView/TransactionItem.mobile';
import useIsMobile from 'hooks/useIsMobile';
import React, { Fragment } from 'react';
import { AccessTime, Work } from '@mui/icons-material';
import { Avatar, Chip, Divider, ListItem, ListItemAvatar, ListItemText, useMediaQuery } from '@mui/material';
import classnames from 'classnames';

import TransactionIcon from 'components/Transactions/components/TransactionIcon';
import TransactionNameEditor from 'components/Transactions/TransactionsView/TransactionNameEditor';
import TransactionSpentFromSelection from 'components/Transactions/TransactionsView/TransactionSpentFromSelection';
import Transaction from 'models/Transaction';

import 'components/Transactions/TransactionsView/styles/TransactionItem.scss';

interface Props {
  transaction: Transaction;
}

function TransactionItem(props: Props): JSX.Element {
  const isMobile = useIsMobile();
  if (!isMobile) {
    return (
      <Fragment>
        <ListItem className="flex flex-row transactions-item pl-3 pr-1 md:pr-2.5">
          <div className="flex flex-col md:flex-row basis-9/12 md:basis-10/12">
            <TransactionIcon transaction={ props.transaction } />
            <TransactionNameEditor transaction={ props.transaction } />
            <TransactionSpentFromSelection transaction={ props.transaction } />
          </div>
          <div className="basis-3/12 md:basis-2/12 flex justify-end w-full items-center">
            { props.transaction.isPending && <Chip icon={ <AccessTime /> } label="Pending" className="mr-auto" /> }
            <span className={ classnames('h-full amount align-middle self-center place-self-center', {
              'text-green-600': props.transaction.getIsAddition(),
              'text-red-600': !props.transaction.getIsAddition(),
            }) }>
            <b>{ props.transaction.getAmountString() }</b>
          </span>
          </div>
        </ListItem>
        <Divider />
      </Fragment>
    );
  }

  return <TransactionItemMobile transaction={ props.transaction } />
}

export default React.memo(TransactionItem);
