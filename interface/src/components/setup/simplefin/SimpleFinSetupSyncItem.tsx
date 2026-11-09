import { Check, LoaderCircle, X } from 'lucide-react';

import { Avatar, AvatarFallback } from '@monetr/interface/components/Avatar';
import { Item, ItemContent } from '@monetr/interface/components/Item';
import Typography from '@monetr/interface/components/Typography';
import type { SimpleFinLinkSyncProgress } from '@monetr/interface/hooks/useSimpleFinLinkSyncProgress';
import type BankAccount from '@monetr/interface/models/BankAccount';
import avatarLetter from '@monetr/interface/util/avatarLetter';
import capitalize from '@monetr/interface/util/capitalize';

import styles from './SimpleFinSetupSyncItem.module.scss';

export interface SimpleFinSetupSyncItemProps {
  bankAccount: BankAccount;
  status?: SimpleFinLinkSyncProgress['status'];
}

export default function SimpleFinSetupSyncItem(props: SimpleFinSetupSyncItemProps): React.JSX.Element {
  return (
    <Item>
      <Avatar>
        <AvatarFallback>{avatarLetter(props.bankAccount.name)}</AvatarFallback>
      </Avatar>
      <ItemContent align='default' flex='shrink' gap='none' justify='start' orientation='column' shrink='default'>
        <Typography ellipsis weight='medium'>
          {props.bankAccount.simpleFinBankAccount?.institutionName}
        </Typography>
        <Typography ellipsis>{props.bankAccount.name}</Typography>
      </ItemContent>
      <ItemContent align='center' flex='grow' justify='end' shrink='none' width='fit'>
        <Typography ellipsis>{capitalize(props.status ?? 'Waiting')}</Typography>
        {props.status !== 'complete' && props.status !== 'error' && (
          <LoaderCircle className={styles.loader} data-loading='true' data-testid='simplefin-sync-loading' />
        )}
        {props.status === 'complete' && <Check className={styles.iconComplete} data-testid='simplefin-sync-complete' />}
        {props.status === 'error' && <X className={styles.iconError} data-testid='simplefin-sync-error' />}
      </ItemContent>
    </Item>
  );
}
