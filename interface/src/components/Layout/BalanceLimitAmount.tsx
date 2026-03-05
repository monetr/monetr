import { InfinityIcon } from 'lucide-react';

import Flex from '@monetr/interface/components/Flex';
import Typography from '@monetr/interface/components/Typography';
import { useCurrentBalance } from '@monetr/interface/hooks/useCurrentBalance';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import { BankAccountSubType } from '@monetr/interface/models/BankAccount';
import { AmountType } from '@monetr/interface/util/amounts';

export default function BalanceLimitAmount(): React.JSX.Element {
  const { data: locale } = useLocaleCurrency();
  const { data: bankAccount } = useSelectedBankAccount();
  const { data: balance } = useCurrentBalance();

  switch (bankAccount?.accountSubType) {
    case BankAccountSubType.CreditCard:
      return (
        <Flex gap='sm' justify='between'>
          <Flex flex='shrink'>
            <Typography color='emphasis' ellipsis size='lg' weight='semibold' wrapping='nowrap'>
              <InfinityIcon />
              Limit:
            </Typography>
          </Flex>
          <Typography color='emphasis' size='lg' weight='semibold'>
            {locale.formatAmount(balance?.limit, AmountType.Stored)}
          </Typography>
        </Flex>
      );
  }

  return null;
}
