import React, { Fragment, useCallback, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { startOfDay, startOfTomorrow } from 'date-fns';
import { FormikErrors, FormikHelpers } from 'formik';

import MAmountField from '@monetr/interface/components/MAmountField';
import MFormButton from '@monetr/interface/components/MButton';
import MDatePicker from '@monetr/interface/components/MDatePicker';
import MForm from '@monetr/interface/components/MForm';
import MLink from '@monetr/interface/components/MLink';
import MLogo from '@monetr/interface/components/MLogo';
import MSelectFrequency from '@monetr/interface/components/MSelectFrequency';
import MSpan from '@monetr/interface/components/MSpan';
import MStepper from '@monetr/interface/components/MStepper';
import MTextField from '@monetr/interface/components/MTextField';
import { useCreateBankAccount } from '@monetr/interface/hooks/bankAccounts';
import { useCreateFundingSchedule } from '@monetr/interface/hooks/fundingSchedules';
import { useCreateLink } from '@monetr/interface/hooks/links';
import FundingSchedule from '@monetr/interface/models/FundingSchedule';
import { friendlyToAmount } from '@monetr/interface/util/amounts';

interface SetupManualValues {
  budgetName: string;
  accountName: string;
  startingBalance: number;
  nextPayday: Date;
  ruleset: string;
  paydayAmount: number;
}

const initialValues: SetupManualValues = {
  budgetName: '',
  accountName: '',
  startingBalance: 0.00,
  nextPayday: startOfTomorrow(),
  ruleset: '',
  paydayAmount: 0.00,
};

export default function SetupManual(): JSX.Element {
  const createLink = useCreateLink();
  const createBankAccount = useCreateBankAccount();
  const createFundingSchedule = useCreateFundingSchedule();
  const navigate = useNavigate();
  const [step, setStep] = useState(0);

  const nextStep = useCallback(() => {
    setStep(step => Math.min(step + 1, 3));
  }, [setStep]);

  const previousStep = useCallback(() => {
    setStep(step => Math.max(step - 1, 0));
  }, [setStep]);

  const validate = useCallback((values: SetupManualValues): FormikErrors<SetupManualValues> => {
    const errors: FormikErrors<SetupManualValues> = {};
    if (values.budgetName.trim() === '') {
      errors['budgetName'] = 'You must provide a name for your budget.';
    }

    if (values.accountName.trim() === '') {
      errors['accountName'] = 'You must provide a name for your account.';
    }

    if (!values.nextPayday) {
      errors['nextPayday'] = 'You must provide a date.';
    }



    return errors;
  }, []);

  const submit = useCallback(async (values: SetupManualValues, helper: FormikHelpers<SetupManualValues>) => {
    if (step < 3) {
      nextStep();
      return;
    }
    helper.setSubmitting(true);
    return createLink({
      institutionName: values.budgetName,
      customInstitutionName: values.budgetName,
    })
      .then(link => createBankAccount({
        linkId: link.linkId,
        name: values.accountName,
        availableBalance: friendlyToAmount(values.startingBalance),
        currentBalance: friendlyToAmount(values.startingBalance),
        accountType: 'depository',
        accountSubType: 'checking',
      }))
      .then(bankAccount => createFundingSchedule(new FundingSchedule({
        bankAccountId: bankAccount.bankAccountId,
        name: 'Payday',
        nextOccurrence: startOfDay(values.nextPayday),
        ruleset: values.ruleset,
        estimatedDeposit: friendlyToAmount(values.paydayAmount),
        excludeWeekends: false,
      })))
      .then(fundingSchedule => navigate(`/bank/${fundingSchedule.bankAccountId}/transactions`))
      .catch(error => {
        throw error;
      });
  }, [createLink, createBankAccount, createFundingSchedule, navigate, step, nextStep]);

  function CurrentStep(): JSX.Element {
    switch (step) {
      case 0:
        return <IntroAndName />;
      case 1:
        return <AccountSetup />;
      case 2:
        return <BalancesSetup />;
      case 3:
        return <IncomeSetup />;
      default:
        return null;
    }
  }

  function SetupButtons(): JSX.Element {
    switch (step) {
      case 0:
        return (
          <MFormButton color='primary' onClick={ nextStep }>
            Next
          </MFormButton>
        );
      case 3:
        return (
          <div className='flex gap-4'>
            <MFormButton color='primary' onClick={ previousStep }>
              Back
            </MFormButton>
            <MFormButton color='primary' type='submit'>
              Finish
            </MFormButton>
          </div>
        );
      default:
        return (
          <div className='flex gap-4'>
            <MFormButton color='primary' onClick={ previousStep }>
              Back
            </MFormButton>
            <MFormButton color='primary' onClick={ nextStep }>
              Next
            </MFormButton>
          </div>
        );
    }
  }


  return (
    <div
      className='w-full h-full flex justify-between items-center gap-8 flex-col p-4 md:p-2 overflow-auto'
    >
      <div className='p-0 md:p-8 w-full'>
        <MStepper steps={ ['Intro', 'Account', 'Balances', 'Income'] } activeIndex={ step } />
      </div>
      <MForm
        initialValues={ initialValues }
        validate={ validate }
        onSubmit={ submit }
        className='flex flex-col md:justify-center items-center max-w-sm h-full'
      >
        <MLogo className='w-24 h-24' />
        <div className='w-full flex flex-col justify-center items-center gap-2'>
          <CurrentStep />
          <SetupButtons />
        </div>
        <LogoutFooter />
      </MForm>
      <div className='md:h-28' />
    </div>
  );
}

function IntroAndName(): JSX.Element {
  return (
    <Fragment>
      <MSpan size='2xl' weight='medium'>
        Welcome to monetr!
      </MSpan>
      <MSpan size='lg' color='subtle' className='text-center'>
        Let's create a new budget to get started. What do you want to call this budget?
      </MSpan>
      <MTextField
        name='budgetName'
        label='Budget Name'
        className='w-full'
        placeholder='My Primary Bank'
        required
      />
    </Fragment>
  );
}

function AccountSetup(): JSX.Element {
  return (
    <Fragment>
      <MSpan size='lg' color='subtle' className='text-center'>
        What do you want to call the primary account you want to use for budgeting? For example; your checking account?
      </MSpan>
      <MTextField
        name='accountName'
        label='Account Name'
        className='w-full'
        placeholder='My Checking Account'
        required
      />
    </Fragment>
  );
}

function BalancesSetup(): JSX.Element {
  return (
    <Fragment>
      <MSpan size='lg' color='subtle' className='text-center'>
        What is your current available balance? monetr will use this as a starting point, you can modify this at any
        time later on.
      </MSpan>
      <MAmountField
        name='startingBalance'
        label='Starting Balance'
        className='w-full'
        required
        allowNegative={ false }
      />
    </Fragment>
  );
}

function IncomeSetup(): JSX.Element {
  return (
    <Fragment>
      <MSpan size='lg' color='subtle' className='text-center'>
        How often do you get paid and how much do you get paid typically? monetr uses this to forecast balances based on
        the budgets you create.
      </MSpan>
      <MDatePicker
        name='nextPayday'
        label='When do you get paid next?'
        className='w-full'
        required
        min={ startOfTomorrow() }
      />
      <MSelectFrequency
        dateFrom="nextPayday"
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
      />
    </Fragment>
  );
}

function LogoutFooter(): JSX.Element {
  return (
    <div className='flex justify-center gap-1 mt-4'>
      <MSpan color="subtle" size='sm'>Not ready to continue?</MSpan>
      <MLink to="/logout" size="sm">Logout for now</MLink>
    </div>
  );
}
