import { Coins } from 'lucide-react';

import Flex from '@monetr/interface/components/Flex';
import Typography from '@monetr/interface/components/Typography';
import { useCurrentBalance } from '@monetr/interface/hooks/useCurrentBalance';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { AmountType } from '@monetr/interface/util/amounts';

export default function BalanceCurrentAmount(): JSX.Element {
  const { data: locale } = useLocaleCurrency();
  const { data: balance } = useCurrentBalance();

  return (
    <Flex justify='between' gap='sm'>
      <Flex flex='shrink'>
        <Coins />
        <Typography color='emphasis' size='lg' weight='semibold' ellipsis>
          Current:
        </Typography>
      </Flex>
      <Typography color='emphasis' size='lg' weight='semibold'>
        {locale.formatAmount(balance?.current, AmountType.Stored)}
      </Typography>
    </Flex>
  );
}
