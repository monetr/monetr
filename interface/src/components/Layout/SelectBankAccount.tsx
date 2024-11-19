import React from 'react';
import { useNavigate } from 'react-router-dom';
import { Check } from 'lucide-react';

import { Combobox, ComboboxItemProps, ComboboxOption } from '@monetr/interface/components/Combobox';
import MBadge from '@monetr/interface/components/MBadge';
import MSpan from '@monetr/interface/components/MSpan';
import { Skeleton } from '@monetr/interface/components/Skeleton';
import { useBankAccounts, useSelectedBankAccount } from '@monetr/interface/hooks/bankAccounts';
import { BankAccountStatus } from '@monetr/interface/models/BankAccount';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';
import sortAccounts from '@monetr/interface/util/sortAccounts';

interface SelectBankAccountItem extends ComboboxOption<string> {
  type: string;
  mask: string;
  status: BankAccountStatus;
}

export default function SelectBankAccount(): JSX.Element {
  const navigate = useNavigate();
  const { data: allBankAccounts, isLoading: allIsLoading } = useBankAccounts();
  const { data: selectedBankAccount, isLoading: selectedIsLoading } = useSelectedBankAccount();

  const accounts: Array<SelectBankAccountItem> = sortAccounts(allBankAccounts
    ?.filter(account => account.linkId === selectedBankAccount?.linkId))
    .map(account => ({
      label: account.name,
      value: account.bankAccountId,
      type: account.accountSubType,
      mask: account.mask,
      status: account.status,
    }));

  const current = accounts?.find(account => account.value === selectedBankAccount?.bankAccountId);

  if (allIsLoading || selectedIsLoading) {
    return (
      <Skeleton className='w-full h-10' />
    );
  }

  return (
    <Combobox 
      className='w-full h-[34px] test'
      variant='text'
      size='select'
      options={ accounts } 
      value={ current?.value }
      placeholder='Select a bank account...'
      searchPlaceholder='Search for an account...'
      showSearch={ false }
      onSelect={ value => navigate(`/bank/${value}/transactions`) } 
      components={ {
        Item: BankAccountSelectItem,
      } }
    />
  );
}

function BankAccountSelectItem(props: ComboboxItemProps<string, SelectBankAccountItem>): JSX.Element {
  return (
    <div className='flex items-center w-full gap-1'>
      <Check
        className={ mergeTailwind(
          'mr-1 h-5 w-5 flex-none',
          props.currentValue == props.option.value ? 'opacity-100' : 'opacity-0'
        ) }
      />
      <MSpan className='w-full' color='emphasis' ellipsis>
        { props.option.label }
      </MSpan>
      { props.option.status === 'inactive' && (
        <MBadge size='xs' className='bg-dark-monetr-border-subtle'>
          Inactive
        </MBadge>
      ) }
      { props.option.mask != '' && (
        <MBadge size='xs' className='font-mono'>
          { props.option.mask }
        </MBadge>
      ) }
    </div>
  );
}
