import { Divider, ListItem, Typography } from '@mui/material';
import BankAccount from 'models/BankAccount';
import React, { Fragment } from 'react';
import { useSelector } from 'react-redux';
import { getBalances } from 'shared/balances/selectors/getBalances';

interface LinkedBankAccountItemProps {
  bankAccount: BankAccount;
}

export default function LinkedBankAccountItem(props: LinkedBankAccountItemProps): JSX.Element {
  const balances = useSelector(getBalances).get(props.bankAccount.bankAccountId);

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
      <Divider/>
    </Fragment>
  );
}
