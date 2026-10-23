import { Wallet } from 'lucide-react';

import { Avatar, AvatarFallback } from '@monetr/interface/components/Avatar';
import MerchantIcon from '@monetr/interface/components/MerchantIcon';
import type Spending from '@monetr/interface/models/Spending';
import { type FreeToUse, SpendingType } from '@monetr/interface/models/Spending';

import styles from './SelectSpendingIcon.module.scss';

export interface SelectSpendingIconProps {
  spending: Pick<Spending | FreeToUse, 'spendingType' | 'name'>;
}

export default function SelectSpendingIcon(props: SelectSpendingIconProps): React.JSX.Element {
  // Free-To-Use isn't a real spending item so there is no merchant to look up an icon for
  if (props.spending.spendingType === SpendingType.FreeToUse) {
    return (
      <Avatar size='small'>
        <AvatarFallback>
          <Wallet className={styles.walletIcon} />
        </AvatarFallback>
      </Avatar>
    );
  }

  return <MerchantIcon name={props.spending.name} size='small' />;
}
