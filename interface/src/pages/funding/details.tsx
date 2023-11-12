import React from 'react';
import { useMatch, useNavigate } from 'react-router-dom';
import { DeleteOutlined, HeartBroken, SaveOutlined, TodayOutlined } from '@mui/icons-material';
import { AxiosError } from 'axios';
import { format, isEqual, startOfDay } from 'date-fns';
import { FormikErrors, FormikHelpers } from 'formik';
import { useSnackbar } from 'notistack';

import MAmountField from '@monetr/interface/components/MAmountField';
import MFormButton, { MBaseButton } from '@monetr/interface/components/MButton';
import MCheckbox from '@monetr/interface/components/MCheckbox';
import MDatePicker from '@monetr/interface/components/MDatePicker';
import MForm from '@monetr/interface/components/MForm';
import MSelectFrequency from '@monetr/interface/components/MSelectFrequency';
import MSpan from '@monetr/interface/components/MSpan';
import MTextField from '@monetr/interface/components/MTextField';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import { useFundingSchedule, useRemoveFundingSchedule, useUpdateFundingSchedule } from '@monetr/interface/hooks/fundingSchedules';
import FundingSchedule from '@monetr/interface/models/FundingSchedule';
import { amountToFriendly, friendlyToAmount } from '@monetr/interface/util/amounts';
import { APIError } from '@monetr/interface/util/request';

interface FundingValues {
  name: string;
  nextOccurrence: Date;
  rule: string;
  excludeWeekends: boolean;
  estimatedDeposit: number | null;
}

export default function FundingDetails(): JSX.Element {
  // I don't want to do it this way, but it seems like it's the only way to do it for tests without having the entire
  // router also present in the test?
  const match = useMatch('/bank/:bankId/funding/:fundingId/details');
  const fundingId = +match?.params?.fundingId || null;
  const { data: funding } = useFundingSchedule(fundingId && +fundingId);
  const navigate = useNavigate();
  const updateFundingSchedule = useUpdateFundingSchedule();
  const removeFundingSchedule = useRemoveFundingSchedule();
  const { enqueueSnackbar } = useSnackbar();

  if (!fundingId) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <HeartBroken className='dark:text-dark-monetr-content h-24 w-24' />
        <MSpan className='text-5xl'>
          Something isn't right...
        </MSpan>
        <MSpan className='text-2xl'>
          There wasn't a funding schedule specified...
        </MSpan>
      </div>
    );
  }

  if (!funding) {
    return null;
  }

  function validate(values: FundingValues): FormikErrors<FundingValues> {
    const errors: FormikErrors<FundingValues> = {};

    if (values.rule === '' || !values.rule) {
      errors['rule'] = 'Frequency is required for funding schedules.';
    }

    return errors;
  }

  async function submit(values: FundingValues, helpers: FormikHelpers<FundingValues>) {
    helpers.setSubmitting(true);
    const updatedFunding = new FundingSchedule({
      ...funding,
      name: values.name,
      nextOccurrence: startOfDay(values.nextOccurrence),
      ruleset: values.rule,
      excludeWeekends: values.excludeWeekends,
      estimatedDeposit: friendlyToAmount(values.estimatedDeposit),
    });

    return updateFundingSchedule(updatedFunding)
      .catch((error: AxiosError<APIError>) => {
        const message = error.response.data.error || 'Failed to update funding schedule.';
        enqueueSnackbar(message, {
          variant: 'error',
          disableWindowBlurListener: true,
        });
      })
      .finally(() => helpers.setSubmitting(false));
  }

  function backToFunding() {
    navigate(`/bank/${funding.bankAccountId}/funding`);
  }

  async function removeFunding() {
    if (!funding) {
      return Promise.resolve();
    }

    if (window.confirm(`Are you sure you want to delete funding schedule: ${ funding.name }`)) {
      return removeFundingSchedule(funding)
        .then(() => backToFunding())
        .catch((error: AxiosError<APIError>) => void enqueueSnackbar(error.response.data['error'], {
          variant: 'error',
          disableWindowBlurListener: true,
        }));
    }

    return Promise.resolve();
  }

  const initialValues: FundingValues = {
    name: funding.name,
    nextOccurrence: funding.nextOccurrenceOriginal,
    rule: funding.ruleset,
    excludeWeekends: funding.excludeWeekends,
    // Because we store all amounts in cents, in order to use them in the UI we need to convert them back to dollars.
    estimatedDeposit: amountToFriendly(funding.estimatedDeposit),
  };

  const NextOccurrenceDecorator = () => {
    if (isEqual(funding.nextOccurrence, funding.nextOccurrenceOriginal)) return null;

    return (
      <MSpan data-testid='funding-schedule-weekend-notice' size='sm' weight='medium'>
        Actual occurrence avoids weekend ({ format(funding.nextOccurrence, 'M/dd') })
      </MSpan>
    );
  };

  return (
    <MForm
      className='flex w-full h-full flex-col'
      initialValues={ initialValues }
      onSubmit={ submit }
      validate={ validate }
    >
      <MTopNavigation
        title='Funding Schedules'
        icon={ TodayOutlined }
        breadcrumb={ funding.name }
        base={ `/bank/${funding.bankAccountId}/funding` }
      >
        <MBaseButton color='cancel' className='gap-1 py-1 px-2' onClick={ removeFunding } >
          <DeleteOutlined />
          Remove
        </MBaseButton>
        <MFormButton color='primary' className='gap-1 py-1 px-2' type='submit' role='form'>
          <SaveOutlined />
          Save
        </MFormButton>
      </MTopNavigation>
      <div className='w-full h-full overflow-y-auto min-w-0 p-4'>
        <div className='flex flex-col md:flex-row w-full gap-8 items-center md:items-stretch'>
          <div className='w-full md:w-1/2 flex flex-col'>
            <MTextField className='w-full' label='Name' name='name' id='funding-name-search' required />
            <MDatePicker
              className='w-full'
              label='Next Occurrence'
              name='nextOccurrence'
              labelDecorator={ NextOccurrenceDecorator }
              required
              data-testid='funding-details-date-picker'
            />
            <MSelectFrequency
              className='w-full'
              dateFrom='nextOccurrence'
              label='How often does this funding happen?'
              name='rule'
              placeholder='Select a funding frequency...'
              required
            />
            <MCheckbox
              id='funding-details-exclude-weekends'
              data-testid='funding-details-exclude-weekends'
              name='excludeWeekends'
              label='Exclude weekends'
              description='If it were to land on a weekend, it is adjusted to the previous weekday instead.'
            />
            <MAmountField
              allowNegative={ false }
              label='Estimated Deposit'
              name='estimatedDeposit'
              placeholder='Example: $ 1,000.00'
            />
          </div>
        </div>
      </div>
    </MForm>
  );
}
