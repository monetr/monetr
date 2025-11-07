import { Fragment } from 'react';
import { tz } from '@date-fns/tz';
import { startOfDay, startOfTomorrow } from 'date-fns';
import type { FormikHelpers } from 'formik';
import { useNavigate } from 'react-router-dom';

import { flexVariants } from '@monetr/interface/components/Flex';
import FormAmountField from '@monetr/interface/components/FormAmountField';
import FormDatePicker from '@monetr/interface/components/FormDatePicker';
import MForm from '@monetr/interface/components/MForm';
import MSelectFrequency from '@monetr/interface/components/MSelectFrequency';
import type { ManualLinkSetupForm } from '@monetr/interface/components/setup/manual/ManualLinkSetup';
import ManualLinkSetupButtons from '@monetr/interface/components/setup/manual/ManualLinkSetupButtons';
import type { ManualLinkSetupSteps } from '@monetr/interface/components/setup/manual/ManualLinkSetupSteps';
import Typography from '@monetr/interface/components/Typography';
import { useViewContext } from '@monetr/interface/components/ViewManager';
import { useCreateBankAccount } from '@monetr/interface/hooks/useCreateBankAccount';
import { useCreateFundingSchedule } from '@monetr/interface/hooks/useCreateFundingSchedule';
import { useCreateLink } from '@monetr/interface/hooks/useCreateLink';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { BankAccountSubType, BankAccountType } from '@monetr/interface/models/BankAccount';

export type ManualLinkSetupIncomeValues = {
  nextPayday: Date;
  ruleset: string;
  paydayAmount: number;
};

export default function ManualLinkSetupIncome(): JSX.Element {
  const { data: timezone } = useTimezone();
  const createLink = useCreateLink();
  const createBankAccount = useCreateBankAccount();
  const createFundingSchedule = useCreateFundingSchedule();
  const navigate = useNavigate();
  const viewContext = useViewContext<ManualLinkSetupSteps, unknown, ManualLinkSetupForm>();
  const { data: locale } = useLocaleCurrency(viewContext.formData.currency);
  const initialValues: ManualLinkSetupIncomeValues = {
    nextPayday: startOfTomorrow({
      in: tz(timezone),
    }),
    ruleset: '',
    paydayAmount: 0.0,
    ...viewContext.formData,
  };

  async function submit(values: ManualLinkSetupIncomeValues, helpers: FormikHelpers<ManualLinkSetupIncomeValues>) {
    helpers.setSubmitting(true);
    const data = {
      ...viewContext.formData,
      ...values,
    };
    return createLink({
      institutionName: data.budgetName,
    })
      .then(link =>
        createBankAccount({
          linkId: link.linkId,
          name: data.accountName,
          availableBalance: locale.friendlyToAmount(data.startingBalance),
          currentBalance: locale.friendlyToAmount(data.startingBalance),
          accountType: BankAccountType.Depository,
          accountSubType: BankAccountSubType.Checking,
          currency: data.currency,
        }),
      )
      .then(bankAccount =>
        createFundingSchedule({
          bankAccountId: bankAccount.bankAccountId,
          name: 'Payday',
          nextRecurrence: startOfDay(values.nextPayday, {
            in: tz(timezone),
          }),
          ruleset: values.ruleset,
          estimatedDeposit: locale.friendlyToAmount(values.paydayAmount),
          excludeWeekends: false,
        }),
      )
      .then(fundingSchedule => navigate(`/bank/${fundingSchedule.bankAccountId}/transactions`))
      .catch(error => {
        throw error;
      });
  }

  return (
    <MForm
      initialValues={initialValues}
      onSubmit={submit}
      className={flexVariants({
        orientation: 'column',
        justify: 'center',
        align: 'center',
      })}
    >
      {({ values: { currency } }) => (
        <Fragment>
          <Typography size='lg' color='subtle' align='center'>
            How often do you get paid and how much do you get paid typically? monetr uses this to forecast balances
            based on the budgets you create.
          </Typography>
          <FormDatePicker
            name='nextPayday'
            label='When do you get paid next?'
            className='w-full'
            required
            min={startOfTomorrow({
              in: tz(timezone),
            })}
            autoFocus
          />
          <MSelectFrequency
            dateFrom='nextPayday'
            label='How often do you get paid?'
            placeholder='Select a funding frequency...'
            required
            className='w-full text-start'
            name='ruleset'
          />
          <FormAmountField
            name='paydayAmount'
            label='How much do you usually get paid?'
            className='w-full'
            required
            allowNegative={false}
            currency={currency}
          />
          <ManualLinkSetupButtons />
        </Fragment>
      )}
    </MForm>
  );
}
