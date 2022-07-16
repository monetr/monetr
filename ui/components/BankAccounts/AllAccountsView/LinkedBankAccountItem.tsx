import React, { Fragment } from 'react';
import { Divider, ListItem, Typography } from '@mui/material';

import { useBalance } from 'hooks/balances';
import BankAccount from 'models/BankAccount';

interface LinkedBankAccountItemProps {
  bankAccount: BankAccount;
}

export default function LinkedBankAccountItem(props: LinkedBankAccountItemProps): JSX.Element {
  const balances = useBalance(props.bankAccount.bankAccountId);

  return (
    <Fragment>
      <ListItem button>
        <div className="flex w-full">
          <Typography className="w-1/3 overflow-hidden font-bold overflow-ellipsis flex-nowrap whitespace-nowrap">
            { props.bankAccount.name }
          </Typography>
          <div className="flex flex-auto">
            <Typography className="w-1/2 overflow-hidden m-w-1/2 overflow-ellipsis flex-nowrap whitespace-nowrap">
              <span
                className="font-semibold">Safe-To-Spend:</span> { balances ? balances.getSafeToSpendString() : '...' }
            </Typography>
            <div className="flex w-1/2">
              <Typography className="w-1/2 overflow-hidden text-sm overflow-ellipsis flex-nowrap whitespace-nowrap">
                <span className="font-semibold">Available:</span> { props.bankAccount.getAvailableBalanceString() }
              </Typography>
              <Typography className="w-1/2 overflow-hidden text-sm overflow-ellipsis flex-nowrap whitespace-nowrap">
                <span className="font-semibold">Current:</span> { props.bankAccount.getCurrentBalanceString() }
              </Typography>
            </div>
          </div>
        </div>
      </ListItem>
      <Divider />
    </Fragment>
  );
}
