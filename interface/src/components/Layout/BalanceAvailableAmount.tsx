import { Banknote } from 'lucide-react';

import Flex from '@monetr/interface/components/Flex';
import Typography from '@monetr/interface/components/Typography';
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
        <Flex justify='between' gap='sm'>
          <Flex flex='shrink'>
            <Banknote />
            <Typography color='emphasis' size='lg' weight='semibold' ellipsis>
              Available:
            </Typography>
          </Flex>
          <Typography color='emphasis' size='lg' weight='semibold' align='left'>
            {locale.formatAmount(balance?.available, AmountType.Stored)}
          </Typography>
        </Flex>
      );
  }

  return null;
}
