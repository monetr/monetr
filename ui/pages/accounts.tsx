import React from 'react';
import { useStore } from 'react-redux';

import AllAccountsView from 'components/BankAccounts/AllAccountsView/AllAccountsView';
import fetchMissingBankAccountBalances from 'shared/balances/actions/fetchMissingBankAccountBalances';
import useMountEffect from 'hooks/useMountEffect';

export default function AccountsPage(): JSX.Element {
  const { dispatch, getState } = useStore();
  useMountEffect(() => {
    fetchMissingBankAccountBalances()(dispatch, getState);
  });

  return <AllAccountsView />;
}
