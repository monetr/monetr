import { Wallet } from 'lucide-react';

import Flex from '@monetr/interface/components/Flex';
import Typography from '@monetr/interface/components/Typography';
import { useCurrentBalance } from '@monetr/interface/hooks/useCurrentBalance';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import { AmountType } from '@monetr/interface/util/amounts';

export default function BalanceFreeToUseAmount(): JSX.Element {
  const { data: locale } = useLocaleCurrency();
  const { data: bankAccount } = useSelectedBankAccount();
  const { data: balance } = useCurrentBalance();

  switch (bankAccount?.accountSubType) {
    case 'checking':
    case 'savings': {
      const color = balance?.free >= 0 ? 'emphasis' : 'negative';

      return (
        <Flex gap='sm' justify='between'>
          <Flex flex='shrink'>
            <Wallet />
            <Typography color='emphasis' ellipsis size='lg' weight='semibold'>
              Free-To-Use:
            </Typography>
          </Flex>
          <Typography color={color} size='lg' weight='semibold'>
            {locale.formatAmount(balance?.free, AmountType.Stored)}
          </Typography>
        </Flex>
      );
    }
  }

  return null;
}
