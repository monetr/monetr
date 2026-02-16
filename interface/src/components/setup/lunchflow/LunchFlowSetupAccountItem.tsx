import { useCallback, useMemo } from 'react';
import { useFormikContext } from 'formik';

import { Avatar, AvatarFallback } from '@monetr/interface/components/Avatar';
import { flexVariants } from '@monetr/interface/components/Flex';
import { Item } from '@monetr/interface/components/Item';
import { Switch } from '@monetr/interface/components/Switch';
import type { LunchFlowSetupAccountsForm } from '@monetr/interface/components/setup/lunchflow/LunchFlowSetupAccounts';
import Typography from '@monetr/interface/components/Typography';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import type LunchFlowBankAccount from '@monetr/interface/models/LunchFlowBankAccount';
import { AmountType } from '@monetr/interface/util/amounts';
import { LunchFlowBankAccountStatus } from '@monetr/interface/models/LunchFlowBankAccount';

export interface LunchFlowSetupAccountItemProps {
  data: LunchFlowBankAccount;
}

export default function LunchFlowSetupAccountItem(props: LunchFlowSetupAccountItemProps): React.JSX.Element {
  const { data: locale } = useLocaleCurrency(props.data.currency);
  const formik = useFormikContext<LunchFlowSetupAccountsForm>();

  const checked = useMemo(
    () => formik.values.items[props.data.lunchFlowBankAccountId] ?? false,
    [props.data.lunchFlowBankAccountId, formik],
  );

  const onChange = useCallback(
    (checked: boolean) => {
      const state: { [key: string]: boolean } = formik.values.items;
      formik.setFieldValue('items', {
        ...state,
        [props.data.lunchFlowBankAccountId]: checked,
      });
    },
    [props.data.lunchFlowBankAccountId, formik],
  );

  return (
    <li>
      <Item>
        <Avatar>
          <AvatarFallback>{props.data.name.toUpperCase().charAt(0) || '?'}</AvatarFallback>
        </Avatar>
        <div className={flexVariants({ flex: 'shrink', orientation: 'column', gap: 'none' })}>
          <Typography ellipsis weight='medium'>
            {props.data.institutionName}
          </Typography>
          <Typography ellipsis>{props.data.name}</Typography>
        </div>
        <Typography>{locale.formatAmount(props.data.currentBalance, AmountType.Stored, false)}</Typography>
        <Switch
          checked={checked}
          disabled={props.data.status === LunchFlowBankAccountStatus.Active}
          onCheckedChange={onChange}
        />
      </Item>
    </li>
  );
}
