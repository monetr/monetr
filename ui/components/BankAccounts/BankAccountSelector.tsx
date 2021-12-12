import { Button, Divider, Menu, MenuItem, Typography } from '@mui/material';
import React, { Fragment, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import fetchBalances from 'shared/balances/actions/fetchBalances';
import setSelectedBankAccountId from 'shared/bankAccounts/actions/setSelectedBankAccountId';
import { getBankAccounts } from 'shared/bankAccounts/selectors/getBankAccounts';
import { getBankAccountsLoading } from 'shared/bankAccounts/selectors/getBankAccountsLoading';
import { getSelectedBankAccountId } from 'shared/bankAccounts/selectors/getSelectedBankAccountId';
import { fetchFundingSchedulesIfNeeded } from 'shared/fundingSchedules/actions/fetchFundingSchedulesIfNeeded';
import { getLinks } from 'shared/links/selectors/getLinks';
import { getLinksLoading } from 'shared/links/selectors/getLinksLoading';
import fetchSpending from 'shared/spending/actions/fetchSpending';
import useFetchInitialTransactionsIfNeeded from 'shared/transactions/actions/fetchInitialTransactionsIfNeeded';
import { ArrowDropDown, CheckCircle } from '@mui/icons-material';
import classnames from 'classnames';
import { useNavigate } from 'react-router-dom';

const BankAccountSelectorMenu = (props: { closeMenu: () => void }): JSX.Element => {
  const dispatch = useDispatch();

  const navigate = useNavigate();
  const selectedBankAccountId = useSelector(getSelectedBankAccountId);
  const bankAccounts = useSelector(getBankAccounts);
  const links = useSelector(getLinks);

  const fetchInitialTransactionsIfNeeded = useFetchInitialTransactionsIfNeeded();

  function goToAllAccounts() {
    navigate('/accounts');
  }

  const changeBankAccount = (bankAccountId: number) => (): Promise<[void, void, void, void]> => {
    dispatch(setSelectedBankAccountId(bankAccountId));
    props.closeMenu();
    return Promise.all([
      void fetchInitialTransactionsIfNeeded(),
      void dispatch(fetchFundingSchedulesIfNeeded()),
      void dispatch(fetchSpending()),
      void dispatch(fetchBalances()),
    ]);
  };

  const bankAccountsViewButton = (
    <MenuItem key="viewBankAccounts" onClick={ goToAllAccounts }>
      <Typography>
        View Bank Accounts
      </Typography>
    </MenuItem>
  );

  if (bankAccounts.isEmpty()) {
    return bankAccountsViewButton
  }

  let items = bankAccounts
    .sortBy(bankAccount => {
      const link = links.get(bankAccount.linkId);

      return `${ link.getName() } - ${ bankAccount.name }`;
    })
    .map(bankAccount => {
      const link = links.get(bankAccount.linkId);
      return (
        <MenuItem
          key={ bankAccount.bankAccountId }
          onClick={ changeBankAccount(bankAccount.bankAccountId) }
        >
          <CheckCircle color="primary" className={ classnames('mr-1', {
            'opacity-0': bankAccount.bankAccountId !== selectedBankAccountId,
          }) }/>
          { /* make it so its the link name - bank name */ }
          { link.getName() } - { bankAccount.name }
        </MenuItem>
      )
    })
    .valueSeq()
    .toArray();

  items.push(<Divider key="divider" className="w-96"/>);
  items.push(bankAccountsViewButton);

  return ( // It won't let me just return the array as a valid JSX.Element, so wrapping it like this makes it valid.
    <Fragment>
      { items }
    </Fragment>
  );
}

const BankAccountSelector = (): JSX.Element => {
  const selectedBankAccountId = useSelector(getSelectedBankAccountId);
  const bankAccounts = useSelector(getBankAccounts);
  const bankAccountsLoading = useSelector(getBankAccountsLoading);
  const linksLoading = useSelector(getLinksLoading);

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
    title = bankAccounts.get(selectedBankAccountId, null)?.name;
  }

  return (
    <Fragment>
      <Button
        className="text-white"
        onClick={ handleOpenMenu }
        aria-label="menu"
      >
        <Typography
          color="inherit"
          className="mr-1"
          variant="h6"
        >
          { title }
        </Typography>
        <ArrowDropDown scale={ 1.25 } color="inherit"/>
      </Button>
      <Menu
        className="w-96 pt-0 pb-0"
        id="bank-account-menu"
        anchorEl={ anchorEl }
        keepMounted
        open={ open }
        onClose={ handleCloseMenu }
      >
        <BankAccountSelectorMenu closeMenu={ handleCloseMenu }/>
      </Menu>
    </Fragment>
  );
};

export default BankAccountSelector;
