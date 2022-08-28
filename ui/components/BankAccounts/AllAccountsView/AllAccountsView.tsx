import React, { Fragment, useState } from 'react';
import { AccountBalance, Add } from '@mui/icons-material';
import { Button, Fab, List, Typography } from '@mui/material';
import * as R from 'ramda';

import AddBankAccountDialog from 'components/BankAccounts/AllAccountsView/AddBankAccountDialog';
import LinkedAccountItem from 'components/BankAccounts/AllAccountsView/LinkedAccountItem';
import { useBankAccounts } from 'hooks/bankAccounts';
import { useLinks } from 'hooks/links';
import BankAccount from 'models/BankAccount';
import Link from 'models/Link';

export default function AllAccountsView(): JSX.Element {
  enum DialogOpen {
    CreateBankAccount,
  }

  const bankAccounts = useBankAccounts();
  const links = useLinks();
  const [dialog, setDialog] = useState<DialogOpen | null>();

  const openDialog = (dialog: DialogOpen) => () => setDialog(dialog);

  function closeDialog() {
    return setDialog(null);
  }

  function Dialogs(): JSX.Element {
    switch (dialog) {
      case DialogOpen.CreateBankAccount:
        return <AddBankAccountDialog open={ true } onClose={ closeDialog } />;
      default:
        return null;
    }
  }

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
              onClick={ openDialog(DialogOpen.CreateBankAccount) }
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
    if (bankAccounts.size === 0) {
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
        link: links.get(parseInt(linkId)),
      })),
      R.values,
      R.sortBy((item: TransformItem) => item.link.getName()),
      R.map((item: TransformItem) => (
        <LinkedAccountItem
          key={ item.link.linkId }
          link={ item.link }
          bankAccounts={ item.bankAccounts } />
      )),
    )(Array.from(bankAccounts.values()));

    return (
      <List disablePadding>
        { items }
      </List>
    );
  }

  return (
    <Fragment>
      <Dialogs />
      <div className="minus-nav bg-primary">
        <div className="w-full h-full view-inner">
          <Content />
        </div>
        <Fab
          color="primary"
          aria-label="add"
          className="absolute z-50 bottom-0 right-5"
          onClick={ openDialog(DialogOpen.CreateBankAccount) }
        >
          <Add />
        </Fab>
      </div>
    </Fragment>
  );
}
