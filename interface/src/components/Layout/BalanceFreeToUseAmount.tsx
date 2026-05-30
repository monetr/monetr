import { Wallet } from 'lucide-react';

import Typography from '@monetr/interface/components/Typography';
import { useCurrentBalance } from '@monetr/interface/hooks/useCurrentBalance';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import { BankAccountSubType } from '@monetr/interface/models/BankAccount';
import { AmountType } from '@monetr/interface/util/amounts';

import styles from './BalanceFreeToUseAmount.module.scss';

export default function BalanceFreeToUseAmount(): JSX.Element {
  const { data: locale } = useLocaleCurrency();
  const { data: bankAccount } = useSelectedBankAccount();
  const { data: balance } = useCurrentBalance();

  switch (bankAccount?.accountSubType) {
    case BankAccountSubType.Checking:
    case BankAccountSubType.Savings: {
      const color = balance?.free >= 0 ? 'emphasis' : 'negative';

      return (
        <div className={styles.root}>
          <div className={styles.freeToUseText}>
            <Wallet />
            <Typography color='emphasis' ellipsis size='lg' weight='semibold' wrapping='nowrap'>
              Free-To-Use:
            </Typography>
          </div>
          <Typography color={color} size='lg' weight='semibold' wrapping='nowrap'>
            {locale.formatAmount(balance?.free, AmountType.Stored)}
          </Typography>
        </div>
      );
    }
  }

  return null;
}
