import { AccountBalance, Add } from '@mui/icons-material';
import { Button, Fab, List, Typography } from '@mui/material';
import LinkedAccountItem from 'components/BankAccounts/AllAccountsView/LinkedAccountItem';
import BankAccount from 'models/BankAccount';
import React from 'react';
import { Fragment, useState } from 'react';
import { useSelector } from 'react-redux';
import { getBankAccounts } from 'shared/bankAccounts/selectors/getBankAccounts';
import { getLinks } from 'shared/links/selectors/getLinks';
import AddBankAccountDialog from 'components/BankAccounts/AllAccountsView/AddBankAccountDialog';

enum DialogOpen {
  CreateBankAccount,
}

export default function AllAccountsView(): JSX.Element {
  const bankAccounts = useSelector(getBankAccounts);
  const links = useSelector(getLinks);
  const [dialog, setDialog] = useState<DialogOpen | null>();

  const openDialog = (dialog: DialogOpen) => () => setDialog(dialog);

  function closeDialog() {
    return setDialog(null);
  }

  function Dialogs(): JSX.Element {
    switch (dialog) {
      case DialogOpen.CreateBankAccount:
        return <AddBankAccountDialog open={ true } onClose={ closeDialog }/>;
      default:
        return null;
    }
  }

  function Empty(): JSX.Element {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="grid grid-cols-1 grid-rows-3 grid-flow-col gap-2">
          <AccountBalance className="self-center w-full h-32 opacity-40"/>
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
    )
  }

  function Content(): JSX.Element {
    if (bankAccounts.isEmpty()) {
      return <Empty/>
    }

    return (
      <List disablePadding>
        { bankAccounts
          .groupBy((item: BankAccount) => item.linkId)
          .map((bankAccounts, linkId) => ({
            bankAccounts: bankAccounts.toMap(),
            link: links.get(linkId),
          }))
          .sortBy(item => item.link.getName())
          .map(item => (
            <LinkedAccountItem key={ item.link.linkId } link={ item.link } bankAccounts={ item.bankAccounts }/>
          ))
          .valueSeq()
          .toArray()
        }
      </List>
    )
  }

  return (
    <Fragment>
      <Dialogs/>
      <div className="minus-nav">
        <Content/>
        <Fab
          color="primary"
          aria-label="add"
          className="absolute z-50 bottom-0 right-5"
          onClick={ openDialog(DialogOpen.CreateBankAccount) }
        >
          <Add/>
        </Fab>
      </div>
    </Fragment>
  )
}
