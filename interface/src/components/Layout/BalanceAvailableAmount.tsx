import React from 'react';
import { LocalAtmOutlined } from '@mui/icons-material';

import MSpan from '@monetr/interface/components/MSpan';
import { useCurrentBalance } from '@monetr/interface/hooks/useCurrentBalance';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import { AmountType } from '@monetr/interface/util/amounts';

export default function BalanceAvailableAmount(): JSX.Element {
  const { data: locale } = useLocaleCurrency();
  const { data: bankAccount } = useSelectedBankAccount();
  const { data: balance } = useCurrentBalance();

  switch (bankAccount?.accountSubType) {
    case 'checking':
    case 'savings':
      return (
        <div className='flex w-full justify-between'>
          <MSpan size='lg' weight='semibold' className='dark:text-dark-monetr-content-emphasis'>
            <LocalAtmOutlined />
            Available:
          </MSpan>
          <MSpan size='lg' weight='semibold' className='dark:text-dark-monetr-content-emphasis'>
            {locale.formatAmount(balance?.available, AmountType.Stored)}
          </MSpan>
        </div>
      );
  }

  return null;
}
