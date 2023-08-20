import React from 'react';
import { useNavigate } from 'react-router-dom';
import Select, { Theme } from 'react-select';

import { useBankAccounts, useSelectedBankAccount } from 'hooks/bankAccounts';
import useTheme from 'hooks/useTheme';

import './MSelectAccount.scss';

export default function MSelectAccount(): JSX.Element {
  const theme = useTheme();
  const navigate = useNavigate();
  const { data: allBankAccounts, isLoading: allIsLoading } = useBankAccounts();
  const { data: selectedBankAccount, isLoading: selectedIsLoading } = useSelectedBankAccount();

  const accounts = allBankAccounts
    ?.filter(account => account.linkId === selectedBankAccount?.linkId)
    .sort((a, b) => {
      const items = [a, b];
      const values = [
        0, // a
        0, // b
      ];
      for (let i = 0; i < 2; i++) {
        const item = items[i];
        if (item.accountType === 'depository') {
          values[i] += 2;
        }
        switch (item.accountSubType) {
          case 'checking':
            values[i] += 2;
            break;
          case 'savings':
            values[i] += 1;
            break;
        }
      }

      return values[0];
    })
    .map(account => ({
      label: account.name,
      value: account.bankAccountId,
      type: account.accountSubType,
      mask: account.mask,
    }));

  const current = accounts?.find(account => account.value === selectedBankAccount?.bankAccountId);

  function onChange({ value }: { value: number }) {
    navigate(`/bank/${value}/transactions`);
  }

  return (
    <Select
      theme={ (baseTheme: Theme): Theme => ({
        ...baseTheme,
        borderRadius: 8,
        colors: {
          ...baseTheme.colors,
          neutral0: theme.tailwind.colors['dark-monetr']['background']['DEFAULT'],
          neutral5: theme.tailwind.colors['dark-monetr']['background']['subtle'],
          neutral10: theme.tailwind.colors['dark-monetr']['background']['emphasis'],
          neutral20: theme.tailwind.colors['dark-monetr']['border']['string'],
          neutral30: theme.tailwind.colors['dark-monetr']['content']['DEFAULT'],
          neutral60: theme.tailwind.colors['dark-monetr']['content']['emphasis'],
          neutral70: theme.tailwind.colors['dark-monetr']['content']['emphasis'],
          neutral80: theme.tailwind.colors['dark-monetr']['content']['emphasis'],
          neutral90: theme.tailwind.colors['dark-monetr']['content']['emphasis'],
          primary25: theme.tailwind.colors['dark-monetr']['background']['emphasis'],
          primary50: theme.tailwind.colors['dark-monetr']['brand']['faint'],
          primary: theme.tailwind.colors['dark-monetr']['brand']['DEFAULT'],
        },
      }) }
      onChange={ onChange }
      isClearable={ false }
      options={ accounts }
      value={ current }
      className="w-full font-medium"
      classNamePrefix='m-select-account'
      isLoading={ allIsLoading || selectedIsLoading }
      styles={ {
        option: (base: object) => ({
          ...base,
          color: theme.tailwind.colors['dark-monetr']['content']['emphasized'],
        }),
        menuPortal: (base: object) => ({
          ...base,
          zIndex: 9999,
        }),
      } }
    />
  );
}
