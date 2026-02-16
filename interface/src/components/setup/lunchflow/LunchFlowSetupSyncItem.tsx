import { Check, LoaderCircle, X } from 'lucide-react';

import { Avatar, AvatarFallback } from '@monetr/interface/components/Avatar';
import { Item, ItemContent } from '@monetr/interface/components/Item';
import Typography from '@monetr/interface/components/Typography';
import useLunchFlowLinkSyncProgress from '@monetr/interface/hooks/useLunchFlowLinkSyncProgress';
import type BankAccount from '@monetr/interface/models/BankAccount';
import capitalize from '@monetr/interface/util/capitalize';

export interface LunchFlowSetupSyncItemProps {
  bankAccount: BankAccount;
}

export default function LunchFlowSetupSyncItem({ bankAccount }: LunchFlowSetupSyncItemProps): React.JSX.Element {
  const { data } = useLunchFlowLinkSyncProgress(bankAccount.linkId, bankAccount.bankAccountId);

  return (
    <Item>
      <Avatar>
        <AvatarFallback>{bankAccount.name.toUpperCase().charAt(0) || '?'}</AvatarFallback>
      </Avatar>
      <ItemContent align='default' flex='shrink' gap='none' justify='start' orientation='column' shrink='default'>
        <Typography ellipsis weight='medium'>
          {bankAccount.lunchFlowBankAccount.institutionName}
        </Typography>
        <Typography ellipsis>{bankAccount.name}</Typography>
      </ItemContent>
      <ItemContent align='center' flex='grow' justify='end' shrink='none' width='fit'>
        <Typography ellipsis>{capitalize(data?.status ?? 'Waiting')}</Typography>
        {data?.status !== 'complete' && <LoaderCircle className='animate-spin' data-loading='true' />}
        {data?.status === 'complete' && <Check className='text-green-500' />}
        {data?.status === 'error' && <X className='text-red-500' />}
      </ItemContent>
    </Item>
  );
}
