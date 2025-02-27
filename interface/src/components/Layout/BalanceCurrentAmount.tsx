import React from 'react';
import { TollOutlined } from '@mui/icons-material';

import MSpan from '@monetr/interface/components/MSpan';
import { useCurrentBalance } from '@monetr/interface/hooks/balances';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { AmountType } from '@monetr/interface/util/amounts';

export default function BalanceCurrentAmount(): JSX.Element {
  const { data: locale } = useLocaleCurrency();
  const { data: balance } = useCurrentBalance();

  return (
    <div className='flex w-full justify-between'>
      <MSpan size='lg' weight='semibold' className='dark:text-dark-monetr-content-emphasis'>
        <TollOutlined />
        Current:
      </MSpan>
      <MSpan size='lg' weight='semibold' className='dark:text-dark-monetr-content-emphasis'>
        { locale.formatAmount(balance?.current, AmountType.Stored) }
      </MSpan>
    </div>
  );
}
