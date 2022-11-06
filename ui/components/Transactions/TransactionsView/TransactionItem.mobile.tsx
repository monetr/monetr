import { AccessTime } from '@mui/icons-material';
import classnames from 'classnames';
import { useSpending } from 'hooks/spending';
import React, { Fragment } from 'react';
import { Chip, Divider, ListItem, ListItemAvatar, ListItemButton, ListItemText, Skeleton } from '@mui/material';

import TransactionIcon from 'components/Transactions/components/TransactionIcon';
import Transaction from 'models/Transaction';

import 'components/Transactions/TransactionsView/styles/TransactionItem.scss';

interface Props {
  transaction: Transaction;
}

export default function TransactionItemMobile(props: Props): JSX.Element {
  const spending = useSpending(props.transaction.spendingId)

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
      )
    }

    const name = spending ?
      <span className="text-black font-medium dark:text-white text-ellipsis overflow-hidden">
        { spending?.name }
      </span> : 'Safe-To-Spend';

    return (
      <span className="text-ellipsis overflow-hidden">
        Spent From { name }
      </span>
    );
  }

  return (
    <Fragment>
      <ListItemButton className="pr-0">
        <ListItemAvatar>
          <TransactionIcon transaction={ props.transaction }/>
        </ListItemAvatar>
        <ListItemText
          className="flex-initial w-7/12"
          primary={ props.transaction.getName() }
          secondaryTypographyProps={{
            className: "text-ellipsis overflow-hidden truncate"
          }}
          secondary={ <SpentFromLine /> }
        />
        <div className="flex-1 flex justify-start">
          { props.transaction.isPending && <AccessTime />
        }
        </div>
        <span className={ classnames('h-full flex-none amount align-middle self-center justify-end place-self-center text-sm', {
          'text-green-600': props.transaction.getIsAddition(),
          'text-red-600': !props.transaction.getIsAddition(),
        }) }>
          <b>{ props.transaction.getAmountString() }</b>
        </span>
      </ListItemButton>
      <Divider />
    </Fragment>
  );
}
