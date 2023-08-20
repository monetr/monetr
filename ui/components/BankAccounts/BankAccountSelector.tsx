import React, { Fragment, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { AccountBalanceWallet, ArrowDropDown, CheckCircle } from '@mui/icons-material';
import { Button, Divider, Menu, MenuItem, Typography } from '@mui/material';
import classnames from 'classnames';
import * as R from 'ramda';
import shallow from 'zustand/shallow';

import { useBankAccounts, useSelectedBankAccountId } from 'hooks/bankAccounts';
import { useLinks } from 'hooks/links';
import useStore from 'hooks/store';
import BankAccount from 'models/BankAccount';

const BankAccountSelectorMenu = (props: { closeMenu: () => void }): JSX.Element => {
  const navigate = useNavigate();
  const { selectedBankAccountId, setCurrentBankAccount } = useStore(state => ({
    selectedBankAccountId: state.selectedBankAccountId,
    setCurrentBankAccount: state.setCurrentBankAccount,
  }), shallow);
  const { data: bankAccounts } = useBankAccounts();
  const { data: links } = useLinks();

  function goToAllAccounts() {
    props.closeMenu();
    navigate('/accounts');
  }

  const changeBankAccount = (bankAccountId: number) => (): Promise<void> => {
    setCurrentBankAccount(bankAccountId);
    props.closeMenu();
    return Promise.resolve();
  };

  const bankAccountsViewButton = (
    <MenuItem key="viewBankAccounts" onClick={ goToAllAccounts }>
      <Typography>
        View Bank Accounts
      </Typography>
    </MenuItem>
  );

  if (bankAccounts.length === 0) {
    return bankAccountsViewButton;
  }

  const items = R.pipe(
    R.sortBy((bankAccount: BankAccount) => {
      const link = links.find(link => link.linkId === bankAccount.linkId);

      return `${ link.getName() } - ${ bankAccount.name }`;
    }),
    R.map((bankAccount: BankAccount) => {
      const link = links.find(link => link.linkId === bankAccount.linkId);
      return (
        <MenuItem
          key={ bankAccount.bankAccountId }
          onClick={ changeBankAccount(bankAccount.bankAccountId) }
        >
          <CheckCircle color="primary" className={ classnames('mr-1', {
            'opacity-0': bankAccount.bankAccountId !== selectedBankAccountId,
          }) } />
          { /* make it so its the link name - bank name */ }
          { link.getName() } - { bankAccount.name }
        </MenuItem>
      );
    }),
  )(bankAccounts as Array<Readonly<BankAccount>>);

  items.push(<Divider key="divider" className="w-96" />);
  items.push(bankAccountsViewButton);

  return ( // It won't let me just return the array as a valid JSX.Element, so wrapping it like this makes it valid.
    <Fragment>
      { items }
    </Fragment>
  );
};

const BankAccountSelector = (): JSX.Element => {
  const selectedBankAccountId = useSelectedBankAccountId();
  const {
    isLoading: bankAccountsLoading,
    data: bankAccounts,
  } = useBankAccounts();

  const { isLoading: linksLoading } = useLinks();

  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);

  function handleOpenMenu(event: React.MouseEvent<HTMLButtonElement>) {
    setAnchorEl(event.currentTarget);
  }

  function handleCloseMenu() {
    setAnchorEl(null);
  }

  if (bankAccountsLoading || linksLoading) {
    return null;
  }

  let title = 'Select A Bank Account';
  if (selectedBankAccountId) {
    title = bankAccounts?.find(item => item.bankAccountId === selectedBankAccountId)?.name;
  }

  return (
    <Fragment>
      <Button
        color="inherit"
        className="text-lg w-full"
        onClick={ handleOpenMenu }
        aria-label="menu"
      >
        <AccountBalanceWallet className="mr-2.5" />
        { title }
        <ArrowDropDown scale={ 1.25 } color="inherit" className="ml-auto" />
      </Button>
      <Menu
        className="w-96 pt-0 pb-0"
        id="bank-account-menu"
        anchorEl={ anchorEl }
        keepMounted
        open={ open }
        onClose={ handleCloseMenu }
      >
        <BankAccountSelectorMenu closeMenu={ handleCloseMenu } />
      </Menu>
    </Fragment>
  );
};

export default BankAccountSelector;
