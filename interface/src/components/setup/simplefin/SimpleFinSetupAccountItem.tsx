import { useCallback, useMemo } from 'react';
import { useFormikContext } from 'formik';

import { Avatar, AvatarFallback } from '@monetr/interface/components/Avatar';
import { Item, ItemContent } from '@monetr/interface/components/Item';
import { Switch } from '@monetr/interface/components/Switch';
import type { SimpleFinSetupAccountsForm } from '@monetr/interface/components/setup/simplefin/SimpleFinSetupAccounts';
import Typography from '@monetr/interface/components/Typography';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import type SimpleFinBankAccount from '@monetr/interface/models/SimpleFinBankAccount';
import { SimpleFinBankAccountStatus } from '@monetr/interface/models/SimpleFinBankAccount';
import { AmountType } from '@monetr/interface/util/amounts';
import avatarLetter from '@monetr/interface/util/avatarLetter';

export interface SimpleFinSetupAccountItemProps {
  data: SimpleFinBankAccount;
}

export default function SimpleFinSetupAccountItem(props: SimpleFinSetupAccountItemProps): React.JSX.Element {
  const { data: locale } = useLocaleCurrency(props.data.currency);
  const formik = useFormikContext<SimpleFinSetupAccountsForm>();

  const checked = useMemo(
    () => formik.values.items[props.data.simpleFinBankAccountId] ?? false,
    [props.data.simpleFinBankAccountId, formik],
  );

  const onChange = useCallback(
    (checked: boolean) => {
      const state: { [key: string]: boolean } = formik.values.items;
      formik.setFieldValue('items', {
        ...state,
        [props.data.simpleFinBankAccountId]: checked,
      });
    },
    [props.data.simpleFinBankAccountId, formik],
  );

  return (
    <Item>
      <Avatar>
        <AvatarFallback>{avatarLetter(props.data.name)}</AvatarFallback>
      </Avatar>
      <ItemContent align='default' flex='shrink' gap='none' justify='start' orientation='column' shrink='default'>
        <Typography ellipsis weight='medium'>
          {props.data.institutionName}
        </Typography>
        <Typography ellipsis>{props.data.name}</Typography>
      </ItemContent>
      <ItemContent align='center' flex='grow' justify='end' shrink='none' width='fit'>
        <Typography>{locale?.formatAmount(props.data.currentBalance, AmountType.Stored, false)}</Typography>
        <Switch
          checked={checked}
          disabled={props.data.status === SimpleFinBankAccountStatus.Active || formik.isSubmitting}
          onCheckedChange={onChange}
        />
      </ItemContent>
    </Item>
  );
}
