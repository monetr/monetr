import React from 'react';
import { useNavigate } from 'react-router-dom';

import { Combobox } from '@monetr/interface/components/Combobox';
import { Skeleton } from '@monetr/interface/components/Skeleton';
import { useBankAccounts, useSelectedBankAccount } from '@monetr/interface/hooks/bankAccounts';
import sortAccounts from '@monetr/interface/util/sortAccounts';


export default function SelectBankAccount(): JSX.Element {
  const navigate = useNavigate();
  const { data: allBankAccounts, isLoading: allIsLoading } = useBankAccounts();
  const { data: selectedBankAccount, isLoading: selectedIsLoading } = useSelectedBankAccount();

  const accounts = sortAccounts(allBankAccounts
    ?.filter(account => account.linkId === selectedBankAccount?.linkId))
    .map(account => ({
      label: account.name,
      value: account.bankAccountId,
      type: account.accountSubType,
      mask: account.mask,
    }));

  const current = accounts?.find(account => account.value === selectedBankAccount?.bankAccountId);

  if (allIsLoading || selectedIsLoading) {
    return (
      <Skeleton className='w-full' />
    );
  }

  return (
    <Combobox 
      options={ accounts } 
      value={ current?.value }
      placeholder='Select a bank account...'
      onSelect={ value => navigate(`/bank/${value}/transactions`) } 
    />
  );
}
