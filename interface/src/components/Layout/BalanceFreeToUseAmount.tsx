import React from 'react';
import { AccountBalanceWalletOutlined } from '@mui/icons-material';

import MSpan from '@monetr/interface/components/MSpan';
import { useCurrentBalance } from '@monetr/interface/hooks/useCurrentBalance';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import { AmountType } from '@monetr/interface/util/amounts';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export default function BalanceFreeToUseAmount(): JSX.Element {
  const { data: locale } = useLocaleCurrency();
  const { data: bankAccount } = useSelectedBankAccount();
  const { data: balance } = useCurrentBalance();

  switch (bankAccount?.accountSubType) {
    case 'checking':
    case 'savings':
      const valueClassName = mergeTailwind({
        'dark:text-dark-monetr-content-emphasis': balance?.free >= 0,
        'dark:text-dark-monetr-red': balance?.free < 0,
      });

      return (
        <div className='flex w-full justify-between flex-shrink min-w-fit'>
          <MSpan size='lg' weight='semibold' className='dark:text-dark-monetr-content-emphasis'>
            <AccountBalanceWalletOutlined />
            Free-To-Use:
          </MSpan>
          &nbsp;
          <MSpan size='lg' weight='semibold' className={valueClassName}>
            {locale.formatAmount(balance?.free, AmountType.Stored)}
          </MSpan>
        </div>
      );
  }

  return null;
}
