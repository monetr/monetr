import { Fragment } from 'react';
import { startOfDay, startOfTomorrow } from 'date-fns';
import type { FormikHelpers } from 'formik';
import { useLocation } from 'wouter';

import { flexVariants } from '@monetr/interface/components/Flex';
import FormAmountField from '@monetr/interface/components/FormAmountField';
import FormDatePicker from '@monetr/interface/components/FormDatePicker';
import MForm from '@monetr/interface/components/MForm';
import MSelectFrequency from '@monetr/interface/components/MSelectFrequency';
import type {
  ManualLinkSetupForm,
  ManualLinkSetupMetadata,
} from '@monetr/interface/components/setup/manual/ManualLinkSetup';
import ManualLinkSetupButtons from '@monetr/interface/components/setup/manual/ManualLinkSetupButtons';
import type { ManualLinkSetupSteps } from '@monetr/interface/components/setup/manual/ManualLinkSetupSteps';
import Typography from '@monetr/interface/components/Typography';
import { useViewContext } from '@monetr/interface/components/ViewManager';
import { useCreateBankAccount } from '@monetr/interface/hooks/useCreateBankAccount';
import { useCreateFundingSchedule } from '@monetr/interface/hooks/useCreateFundingSchedule';
import { useCreateLink } from '@monetr/interface/hooks/useCreateLink';
import useLocaleCurrency, { DefaultCurrency } from '@monetr/interface/hooks/useLocaleCurrency';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { BankAccountSubType, BankAccountType } from '@monetr/interface/models/BankAccount';

import styles from './ManualLinkSetupIncome.module.scss';

export type ManualLinkSetupIncomeValues = {
  nextPayday: Date;
  ruleset: string;
  paydayAmount: number;
  currency: string;
};

export default function ManualLinkSetupIncome(): React.JSX.Element {
  const { inTimezone } = useTimezone();
  const createLink = useCreateLink();
  const createBankAccount = useCreateBankAccount();
  const createFundingSchedule = useCreateFundingSchedule();
  const [, navigate] = useLocation();
  const viewContext = useViewContext<ManualLinkSetupSteps, ManualLinkSetupMetadata, ManualLinkSetupForm>();
  const { data: locale } = useLocaleCurrency(viewContext.formData.currency);
  const initialValues: ManualLinkSetupIncomeValues = {
    nextPayday: startOfTomorrow({
      in: inTimezone,
    }),
    ruleset: '',
    paydayAmount: 0.0,
    currency: locale?.currency ?? DefaultCurrency,
    ...(viewContext.formData as Partial<ManualLinkSetupForm>),
  };

  async function submit(values: ManualLinkSetupIncomeValues, helpers: FormikHelpers<ManualLinkSetupIncomeValues>) {
    // The locale is needed to convert the friendly amounts into stored amounts, it should always be loaded by the time
    // we can submit but bail just in case it is not ready yet.
    if (!locale) {
      return Promise.resolve();
    }

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
          description: null,
          nextRecurrence: startOfDay(values.nextPayday, {
            in: inTimezone,
          }),
          ruleset: values.ruleset,
          estimatedDeposit: values.paydayAmount === 0 ? null : locale.friendlyToAmount(values.paydayAmount),
          excludeWeekends: false,
          autoCreateTransaction: false,
        }),
      )
      .then(fundingSchedule => navigate(`/bank/${fundingSchedule.bankAccountId}/transactions`))
      .catch(error => {
        throw error;
      });
  }

  return (
    <MForm
      className={flexVariants({
        orientation: 'column',
        justify: 'center',
        align: 'center',
      })}
      initialValues={initialValues}
      onSubmit={submit}
    >
      {({ values: { currency } }) => (
        <Fragment>
          <Typography align='center' color='subtle' size='lg'>
            How often do you get paid and how much do you get paid typically? monetr uses this to forecast balances
            based on the budgets you create.
          </Typography>
          <FormDatePicker
            autoFocus
            className={styles.field}
            label='When do you get paid next?'
            min={startOfTomorrow({
              in: inTimezone,
            })}
            name='nextPayday'
            required
          />
          <MSelectFrequency
            className={styles.frequencyField}
            dateFrom='nextPayday'
            label='How often do you get paid?'
            name='ruleset'
            placeholder='Select a funding frequency...'
            required
          />
          <FormAmountField
            allowNegative={false}
            className={styles.field}
            currency={currency}
            label='How much do you usually get paid?'
            name='paydayAmount'
            required
          />
          <ManualLinkSetupButtons />
        </Fragment>
      )}
    </MForm>
  );
}
