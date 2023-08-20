import React, { Fragment } from 'react';
import { AccountBalance, Add } from '@mui/icons-material';
import { Button, Fab, List, Typography } from '@mui/material';
import * as R from 'ramda';

import { showAddBankAccountDialog } from './AddBankAccountDialog';

import LinkedAccountItem from 'components/BankAccounts/AllAccountsView/LinkedAccountItem';
import { useBankAccounts } from 'hooks/bankAccounts';
import { useLinks } from 'hooks/links';
import BankAccount from 'models/BankAccount';
import Link from 'models/Link';

export default function AllAccountsView(): JSX.Element {
  const { data: bankAccounts } = useBankAccounts();
  const { data: links } = useLinks();

  function Empty(): JSX.Element {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="grid grid-cols-1 grid-rows-3 grid-flow-col gap-2">
          <AccountBalance className="self-center w-full h-32 opacity-40" />
          <div className="flex items-center">
            <Typography
              className="text-center opacity-50"
              variant="h3"
            >
              You don't have any bank accounts yet...
            </Typography>
          </div>
          <div className="w-full">
            <Button
              onClick={ showAddBankAccountDialog }
              color="primary"
              className="w-full"
            >
              <Typography
                variant="h6"
              >
                Create or Add a bank account
              </Typography>
            </Button>
          </div>
        </div>
      </div>
    );
  }

  function Content(): JSX.Element {
    if (bankAccounts.length === 0) {
      return <Empty />;
    }

    interface TransformItem {
      bankAccounts: Array<BankAccount>;
      link: Link;
    }

    const items = R.pipe(
      R.groupBy((item: BankAccount) => item.linkId.toString(10)),
      R.mapObjIndexed((bankAccounts, linkId) => ({
        bankAccounts: bankAccounts,
        link: links.find(link => link.linkId === +linkId),
      })),
      R.values,
      R.sortBy((item: TransformItem) => item.link.getName()),
      R.map((item: TransformItem) => (
        <LinkedAccountItem
          key={ item.link.linkId }
          link={ item.link }
          bankAccounts={ item.bankAccounts } />
      )),
    )(bankAccounts as Array<Readonly<BankAccount>>);

    return (
      <List disablePadding>
        { items }
      </List>
    );
  }

  return (
    <Fragment>
      <div className="minus-nav bg-primary">
        <div className="w-full h-full view-inner">
          <Content />
        </div>
        <Fab
          color="primary"
          aria-label="add"
          className="absolute z-50 bottom-5 right-5"
          onClick={ showAddBankAccountDialog }
        >
          <Add />
        </Fab>
      </div>
    </Fragment>
  );
}
