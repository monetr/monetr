import React, { Fragment } from 'react';
import { AccessTime, ChevronRight } from '@mui/icons-material';
import { Divider, ListItemAvatar, ListItemButton, ListItemText } from '@mui/material';
import classnames from 'classnames';

import { showEditTransactionMobileDialog } from './EditTransactionDialog.mobile';

import TransactionIcon from 'components/Transactions/components/TransactionIcon';
import { useSpending } from 'hooks/spending';
import Transaction from 'models/Transaction';

interface Props {
  transaction: Transaction;
}

export default function TransactionItemMobile(props: Props): JSX.Element {
  const spending = useSpending(props.transaction.spendingId);

  function SpentFromLine(): JSX.Element {
    if (props.transaction.getIsAddition()) {
      return (
        <span className="text-ellipsis overflow-hidden">
          Deposit
        </span>
      );
    }

    if (props.transaction.spendingId && !spending) {
      return (
        <span className="text-ellipsis overflow-hidden">
          Spent From <span className="opacity-75">...</span>
        </span>
      );
    }

    const name = spending ?
      <span className="text-black font-medium dark:text-white text-ellipsis overflow-hidden">
        {spending?.name}
      </span> : 'Free-To-Use';

    return (
      <span className="text-ellipsis overflow-hidden">
        Spent From {name}
      </span>
    );
  }

  const showEditDialog = () => showEditTransactionMobileDialog({
    transaction: props.transaction,
  });

  return (
    <Fragment>
      <ListItemButton className="pr-0" onClick={ showEditDialog }>
        <ListItemAvatar>
          <TransactionIcon transaction={ props.transaction } />
        </ListItemAvatar>
        <ListItemText
          className="flex-initial w-7/12"
          primaryTypographyProps={ {
            className: 'text-ellipsis overflow-hidden truncate',
          } }
          primary={ props.transaction.getName() }
          secondaryTypographyProps={ {
            className: 'text-ellipsis overflow-hidden truncate',
          } }
          secondary={ <SpentFromLine /> }
        />
        <div className="flex-1 flex justify-start">
          {props.transaction.isPending && <AccessTime />
          }
        </div>
        <span className={ classnames('h-full flex-none amount align-middle self-center justify-end place-self-center text-sm pr-1', {
          'text-green-600': props.transaction.getIsAddition(),
          'text-red-600': !props.transaction.getIsAddition(),
        }) }>
          <b>{props.transaction.getAmountString()}</b>
        </span>
        <ChevronRight className='opacity-75' />
      </ListItemButton>
      <Divider />
    </Fragment>
  );
}
