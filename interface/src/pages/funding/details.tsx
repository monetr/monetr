import { useId } from 'react';
import { tz } from '@date-fns/tz';
import type { AxiosError } from 'axios';
import { format, isEqual, startOfDay, startOfTomorrow } from 'date-fns';
import type { FormikErrors, FormikHelpers } from 'formik';
import { CalendarSync, HeartCrack, Save, Trash } from 'lucide-react';
import { useSnackbar } from 'notistack';
import { useMatch, useNavigate } from 'react-router-dom';

import { Button } from '@monetr/interface/components/Button';
import FormButton from '@monetr/interface/components/FormButton';
import FormCheckbox from '@monetr/interface/components/FormCheckbox';
import FormDatePicker from '@monetr/interface/components/FormDatePicker';
import FormTextField from '@monetr/interface/components/FormTextField';
import FundingTimeline from '@monetr/interface/components/funding/FundingTimeline';
import MAmountField from '@monetr/interface/components/MAmountField';
import MDivider from '@monetr/interface/components/MDivider';
import MForm from '@monetr/interface/components/MForm';
import MSelectFrequency from '@monetr/interface/components/MSelectFrequency';
import MSpan from '@monetr/interface/components/MSpan';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import { useFundingSchedule } from '@monetr/interface/hooks/useFundingSchedule';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { usePatchFundingSchedule } from '@monetr/interface/hooks/usePatchFundingSchedule';
import { useRemoveFundingSchedule } from '@monetr/interface/hooks/useRemoveFundingSchedule';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import type FundingSchedule from '@monetr/interface/models/FundingSchedule';
import type { APIError } from '@monetr/interface/util/request';

interface FundingValues {
  name: string;
  nextRecurrence: Date;
  ruleset: string;
  excludeWeekends: boolean;
  estimatedDeposit: number | null;
}

export default function FundingDetails(): JSX.Element {
  const nameId = useId();
  const { data: timezone } = useTimezone();
  const { data: locale } = useLocaleCurrency();
  // I don't want to do it this way, but it seems like it's the only way to do it for tests without having the entire
  // router also present in the test?
  const match = useMatch('/bank/:bankId/funding/:fundingId/details');
  const fundingId = match?.params?.fundingId || null;
  const { data: funding } = useFundingSchedule(fundingId);
  const navigate = useNavigate();
  const patchFundingSchedule = usePatchFundingSchedule();
  const removeFundingSchedule = useRemoveFundingSchedule();
  const { enqueueSnackbar } = useSnackbar();

  if (!fundingId) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <HeartCrack className='dark:text-dark-monetr-content size-24' />
        <MSpan className='text-5xl'>Something isn't right...</MSpan>
        <MSpan className='text-2xl'>There wasn't a funding schedule specified...</MSpan>
      </div>
    );
  }

  if (!funding) {
    return null;
  }

  function validate(values: FundingValues): FormikErrors<FundingValues> {
    const errors: FormikErrors<FundingValues> = {};

    if (!values.name) {
      errors.name = 'Name cannot be blank.';
    }

    if (!values.ruleset) {
      errors.ruleset = 'Frequency is required for funding schedules.';
    }

    return errors;
  }

  async function submit(values: FundingValues, helpers: FormikHelpers<FundingValues>) {
    helpers.setSubmitting(true);
    return patchFundingSchedule({
      fundingScheduleId: funding.fundingScheduleId,
      bankAccountId: funding.bankAccountId,
      name: values.name,
      nextRecurrence: startOfDay(values.nextRecurrence, {
        in: tz(timezone),
      }),
      ruleset: values.ruleset,
      excludeWeekends: values.excludeWeekends,
      estimatedDeposit: locale.friendlyToAmount(values.estimatedDeposit),
    })
      .then(
        () =>
          void enqueueSnackbar('Updated funding schedule successfully', {
            variant: 'success',
            disableWindowBlurListener: true,
          }),
      )
      .catch(
        (error: AxiosError<APIError>) =>
          void enqueueSnackbar(error.response.data.error || 'Failed to update funding schedule', {
            variant: 'error',
            disableWindowBlurListener: true,
          }),
      )
      .finally(() => helpers.setSubmitting(false));
  }

  function backToFunding() {
    navigate(`/bank/${funding.bankAccountId}/funding`);
  }

  async function removeFunding() {
    if (!funding) {
      return Promise.resolve();
    }

    if (window.confirm(`Are you sure you want to delete funding schedule: ${funding.name}`)) {
      return removeFundingSchedule(funding)
        .then(() => backToFunding())
        .catch(
          (error: AxiosError<APIError>) =>
            void enqueueSnackbar(error.response.data.error, {
              variant: 'error',
              disableWindowBlurListener: true,
            }),
        );
    }

    return Promise.resolve();
  }

  const initialValues: FundingValues = {
    name: funding.name,
    nextRecurrence: funding.nextRecurrenceOriginal,
    ruleset: funding.ruleset,
    excludeWeekends: funding.excludeWeekends,
    // Because we store all amounts in cents, in order to use them in the UI we need to convert them back to dollars.
    estimatedDeposit: locale.amountToFriendly(funding.estimatedDeposit),
  };

  return (
    <MForm className='flex w-full h-full flex-col' initialValues={initialValues} onSubmit={submit} validate={validate}>
      <MTopNavigation
        title='Funding Schedules'
        icon={CalendarSync}
        breadcrumb={funding.name}
        base={`/bank/${funding.bankAccountId}/funding`}
      >
        <Button variant='destructive' onClick={removeFunding}>
          <Trash />
          Remove
        </Button>
        <FormButton variant='primary' className='gap-1 py-1 px-2' type='submit' role='form'>
          <Save />
          Save
        </FormButton>
      </MTopNavigation>
      <div className='w-full h-full overflow-y-auto min-w-0 p-4 pb-16 md:pb-4'>
        <div className='flex flex-col md:flex-row w-full gap-8 items-center md:items-stretch'>
          <div className='w-full md:w-1/2 flex flex-col'>
            <FormTextField className='w-full' label='Name' name='name' id={`${nameId}-funding-name-search`} required />
            <FormDatePicker
              className='w-full'
              label='Next Recurrence'
              name='nextRecurrence'
              labelDecorator={() => <NextOccurrenceDecorator fundingSchedule={funding} />}
              required
              data-testid='funding-details-date-picker'
              min={startOfTomorrow({
                in: tz(timezone),
              })}
            />
            <MSelectFrequency
              className='w-full'
              dateFrom='nextRecurrence'
              label='How often does this funding happen?'
              name='ruleset'
              placeholder='Select a funding frequency...'
              required
            />
            <FormCheckbox
              data-testid='funding-details-exclude-weekends'
              name='excludeWeekends'
              label='Exclude weekends'
              description='If it were to land on a weekend, it is adjusted to the previous weekday instead.'
            />
            <MAmountField
              allowNegative={false}
              label='Estimated Deposit'
              name='estimatedDeposit'
              placeholder='Example: $ 1,000.00'
            />
          </div>
          <MDivider className='block md:hidden w-1/2' />
          <div className='w-full md:w-1/2 flex flex-col gap-2'>
            <MSpan className='text-xl my-2'>Funding Timeline</MSpan>
            <FundingTimeline fundingScheduleId={funding.fundingScheduleId} />
          </div>
        </div>
      </div>
    </MForm>
  );
}

interface NextOccurrenceDecoratorProps {
  fundingSchedule: FundingSchedule;
}

function NextOccurrenceDecorator({ fundingSchedule: funding }: NextOccurrenceDecoratorProps): React.JSX.Element {
  if (!funding.excludeWeekends) {
    return null;
  }

  if (isEqual(funding.nextRecurrence, funding.nextRecurrenceOriginal)) {
    return null;
  }

  return (
    <MSpan data-testid='funding-schedule-weekend-notice' size='sm' weight='medium'>
      Actual occurrence avoids weekend ({format(funding.nextRecurrence, 'M/dd')})
    </MSpan>
  );
}
