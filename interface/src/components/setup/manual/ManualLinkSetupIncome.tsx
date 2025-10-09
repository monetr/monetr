import React, { Fragment } from 'react';
import { useNavigate } from 'react-router-dom';
import { tz } from '@date-fns/tz';
import { startOfDay, startOfTomorrow } from 'date-fns';
import { FormikHelpers } from 'formik';

import MAmountField from '@monetr/interface/components/MAmountField';
import MDatePicker from '@monetr/interface/components/MDatePicker';
import MForm from '@monetr/interface/components/MForm';
import MSelectFrequency from '@monetr/interface/components/MSelectFrequency';
import MSpan from '@monetr/interface/components/MSpan';
import ManualLinkSetupButtons from '@monetr/interface/components/setup/manual/ManualLinkSetupButtons';
import { ManualLinkSetupSteps } from '@monetr/interface/components/setup/manual/ManualLinkSetupSteps';
import { useViewContext } from '@monetr/interface/components/ViewManager';
import { useCreateBankAccount } from '@monetr/interface/hooks/useCreateBankAccount';
import { useCreateFundingSchedule } from '@monetr/interface/hooks/useCreateFundingSchedule';
import { useCreateLink } from '@monetr/interface/hooks/useCreateLink';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { BankAccountSubType, BankAccountType } from '@monetr/interface/models/BankAccount';

interface Values {
  nextPayday: Date;
  ruleset: string;
  paydayAmount: number;
}

export default function ManualLinkSetupIncome(): JSX.Element {
  const { data: timezone } = useTimezone();
  const createLink = useCreateLink();
  const createBankAccount = useCreateBankAccount();
  const createFundingSchedule = useCreateFundingSchedule();
  const navigate = useNavigate();
  const viewContext = useViewContext<ManualLinkSetupSteps, {}>();
  const { data: locale } = useLocaleCurrency(viewContext.formData['currency']);
  const initialValues: Values = {
    nextPayday: startOfTomorrow({
      in: tz(timezone),
    }),
    ruleset: '',
    paydayAmount: 0.00,
    ...viewContext.formData,
  };

  async function submit(values: Values, helpers: FormikHelpers<Values>) {
    helpers.setSubmitting(true);
    const data = {
      ...viewContext.formData,
      ...values,
    };
    return createLink({
      institutionName: data['budgetName'],
    })
      .then(link => createBankAccount({
        linkId: link.linkId,
        name: data['accountName'],
        availableBalance: locale.friendlyToAmount(data['startingBalance']),
        currentBalance: locale.friendlyToAmount(data['startingBalance']),
        accountType: BankAccountType.Depository,
        accountSubType: BankAccountSubType.Checking,
        currency: data['currency'],
      }))
      .then(bankAccount => createFundingSchedule({
        bankAccountId: bankAccount.bankAccountId,
        name: 'Payday',
        nextRecurrence: startOfDay(values.nextPayday, {
          in: tz(timezone),
        }),
        ruleset: values.ruleset,
        estimatedDeposit: locale.friendlyToAmount(values.paydayAmount),
        excludeWeekends: false,
      }))
      .then(fundingSchedule => navigate(`/bank/${fundingSchedule.bankAccountId}/transactions`))
      .catch(error => {
        throw error;
      });
  }

  return (
    <MForm
      initialValues={ initialValues }
      onSubmit={ submit }
      className='w-full flex flex-col justify-center items-center gap-2'
    >
      { ({ values: { currency } }) => (
        <Fragment>
          <MSpan size='lg' color='subtle' className='text-center'>
            How often do you get paid and how much do you get paid typically? monetr uses this to forecast balances
            based on the budgets you create.
          </MSpan>
          <MDatePicker
            name='nextPayday'
            label='When do you get paid next?'
            className='w-full'
            required
            min={ startOfTomorrow({
              in: tz(timezone),
            }) }
            autoFocus
          />
          <MSelectFrequency
            dateFrom='nextPayday'
            menuPosition='fixed'
            menuShouldScrollIntoView={ false }
            menuShouldBlockScroll={ true }
            menuPortalTarget={ document.body }
            menuPlacement='bottom'
            label='How often do you get paid?'
            placeholder='Select a funding frequency...'
            required
            className='w-full text-start'
            name='ruleset'
          />
          <MAmountField
            name='paydayAmount'
            label='How much do you usually get paid?'
            className='w-full'
            required
            allowNegative={ false }
            currency={ currency }
          />
          <ManualLinkSetupButtons />
        </Fragment>
      ) }
    </MForm>
  );
}
