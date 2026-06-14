import { useId } from 'react';
import { format, isEqual, startOfDay, startOfTomorrow } from 'date-fns';
import { type FormikErrors, type FormikHelpers, useFormikContext } from 'formik';
import { CalendarSync, HeartCrack, Save, Trash } from 'lucide-react';
import { useLocation, useRoute } from 'wouter';

import type { ApiError } from '@monetr/interface/api/client';
import { Button } from '@monetr/interface/components/Button';
import Divider from '@monetr/interface/components/Divider';
import FormAmountField from '@monetr/interface/components/FormAmountField';
import FormButton from '@monetr/interface/components/FormButton';
import FormCheckbox from '@monetr/interface/components/FormCheckbox';
import FormDatePicker from '@monetr/interface/components/FormDatePicker';
import FormTextField from '@monetr/interface/components/FormTextField';
import FundingTimeline from '@monetr/interface/components/funding/FundingTimeline';
import { layoutVariants } from '@monetr/interface/components/Layout';
import MForm from '@monetr/interface/components/MForm';
import MSelectFrequency from '@monetr/interface/components/MSelectFrequency';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import Typography from '@monetr/interface/components/Typography';
import { useCurrentLink } from '@monetr/interface/hooks/useCurrentLink';
import { useFundingSchedule } from '@monetr/interface/hooks/useFundingSchedule';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { usePatchFundingSchedule } from '@monetr/interface/hooks/usePatchFundingSchedule';
import { useRemoveFundingSchedule } from '@monetr/interface/hooks/useRemoveFundingSchedule';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import type FundingSchedule from '@monetr/interface/models/FundingSchedule';
import type { APIError } from '@monetr/interface/util/request';
import { useSnackbar } from '@monetr/notify';

import styles from './details.module.scss';

interface FundingValues {
  name: string;
  nextRecurrence: Date;
  ruleset: string;
  excludeWeekends: boolean;
  estimatedDeposit: number | null;
  autoCreateTransaction: boolean;
}

export default function FundingDetails(): React.JSX.Element | null {
  const nameId = useId();
  const { inTimezone } = useTimezone();
  const { data: locale } = useLocaleCurrency();
  // I don't want to do it this way, but it seems like it's the only way to do it for tests without having the entire
  // router also present in the test?
  const [, params] = useRoute<{ bankId: string; fundingId: string }>('/bank/:bankId/funding/:fundingId/details');
  const fundingId = params?.fundingId;
  const { data: funding, isLoading, isError } = useFundingSchedule(fundingId);
  const { data: link } = useCurrentLink();
  const isManual = Boolean(link?.getIsManual());
  const [, navigate] = useLocation();
  const patchFundingSchedule = usePatchFundingSchedule();
  const removeFundingSchedule = useRemoveFundingSchedule();
  const { enqueueSnackbar } = useSnackbar();

  if (!fundingId) {
    return (
      <div className={styles.centerState}>
        <HeartCrack className={styles.errorIcon} />
        <Typography size='5xl'>Something isn't right...</Typography>
        <Typography size='2xl'>There wasn't a funding schedule specified...</Typography>
      </div>
    );
  }

  // Treat the locale still loading the same as the funding schedule still loading, otherwise we fall all the way through
  // to the null return below and flash a blank page while the currency formatting catches up.
  if (isLoading || !locale) {
    return (
      <div className={styles.centerState}>
        <Typography size='5xl'>One moment...</Typography>
      </div>
    );
  }

  if (isError) {
    return (
      <div className={styles.centerState}>
        <HeartCrack className={styles.errorIcon} />
        <Typography size='5xl'>Something isn't right...</Typography>
        <Typography size='2xl'>Couldn't find the funding schedule you specified...</Typography>
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
    if (!funding || !locale) {
      return Promise.resolve();
    }

    helpers.setSubmitting(true);
    return patchFundingSchedule({
      fundingScheduleId: funding.fundingScheduleId,
      bankAccountId: funding.bankAccountId,
      name: values.name,
      nextRecurrence: startOfDay(values.nextRecurrence, {
        in: inTimezone,
      }),
      ruleset: values.ruleset,
      excludeWeekends: values.excludeWeekends,
      estimatedDeposit: locale.friendlyToAmount(values.estimatedDeposit ?? 0),
      // Auto create transaction requires a manual link and a non-zero estimated
      // deposit; force it off otherwise so the API will not reject the update.
      autoCreateTransaction: isManual && (values.estimatedDeposit ?? 0) > 0 && values.autoCreateTransaction,
    })
      .then(
        () =>
          void enqueueSnackbar('Updated funding schedule successfully', {
            variant: 'success',
            disableWindowBlurListener: true,
          }),
      )
      .catch(
        (error: ApiError<APIError>) =>
          void enqueueSnackbar(error.response.data.error || 'Failed to update funding schedule', {
            variant: 'error',
            disableWindowBlurListener: true,
          }),
      )
      .finally(() => helpers.setSubmitting(false));
  }

  function backToFunding() {
    navigate(`/bank/${funding?.bankAccountId}/funding`);
  }

  async function removeFunding() {
    if (!funding) {
      return Promise.resolve();
    }

    if (window.confirm(`Are you sure you want to delete funding schedule: ${funding.name}`)) {
      return removeFundingSchedule(funding)
        .then(() => backToFunding())
        .catch(
          (error: ApiError<APIError>) =>
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
    estimatedDeposit: locale.amountToFriendly(funding.estimatedDeposit ?? 0),
    autoCreateTransaction: funding.autoCreateTransaction,
  };

  return (
    <MForm className={styles.form} initialValues={initialValues} onSubmit={submit} validate={validate}>
      <MTopNavigation
        base={`/bank/${funding.bankAccountId}/funding`}
        breadcrumb={funding.name}
        icon={CalendarSync}
        title='Funding Schedules'
      >
        <Button onClick={removeFunding} variant='destructive'>
          <Trash />
          Remove
        </Button>
        <FormButton className={styles.saveButton} role='form' type='submit' variant='primary'>
          <Save />
          Save
        </FormButton>
      </MTopNavigation>
      <div className={styles.body}>
        <div className={styles.columns}>
          <div className={styles.column}>
            <FormTextField
              className={layoutVariants({ width: 'full' })}
              id={`${nameId}-funding-name-search`}
              label='Name'
              name='name'
              required
            />
            <FormDatePicker
              className={layoutVariants({ width: 'full' })}
              data-testid='funding-details-date-picker'
              label='Next Recurrence'
              labelDecorator={() => <NextOccurrenceDecorator fundingSchedule={funding} />}
              min={startOfTomorrow({
                in: inTimezone,
              })}
              name='nextRecurrence'
              required
            />
            <MSelectFrequency
              className={layoutVariants({ width: 'full' })}
              dateFrom='nextRecurrence'
              label='How often does this funding happen?'
              name='ruleset'
              placeholder='Select a funding frequency...'
              required
            />
            <FormCheckbox
              data-testid='funding-details-exclude-weekends'
              description='If it were to land on a weekend, it is adjusted to the previous weekday instead.'
              label='Exclude weekends'
              name='excludeWeekends'
            />
            <FormAmountField
              allowNegative={false}
              label='Estimated Deposit'
              name='estimatedDeposit'
              placeholder='Example: $ 1,000.00'
            />
            {isManual && <AutoCreateTransactionToggle />}
          </div>
          <Divider className={styles.dividerMobile} />
          <div className={styles.columnTimeline}>
            <Typography className={styles.timelineTitle} size='xl'>
              Funding Timeline
            </Typography>
            <FundingTimeline fundingScheduleId={funding.fundingScheduleId} />
          </div>
        </div>
      </div>
    </MForm>
  );
}

function AutoCreateTransactionToggle(): React.JSX.Element {
  const { values } = useFormikContext<FundingValues>();
  // The toggle is always visible on a manual link but is only usable once a
  // non-zero estimated deposit has been provided.
  const hasDeposit = (values.estimatedDeposit ?? 0) > 0;

  return (
    <FormCheckbox
      data-testid='funding-details-auto-create-transaction'
      description='Automatically add a deposit transaction for the estimated deposit each time the funding schedule would occur.'
      disabled={!hasDeposit}
      label='Auto create transaction'
      name='autoCreateTransaction'
    />
  );
}

interface NextOccurrenceDecoratorProps {
  fundingSchedule: FundingSchedule;
}

function NextOccurrenceDecorator({ fundingSchedule: funding }: NextOccurrenceDecoratorProps): React.JSX.Element | null {
  if (!funding.excludeWeekends) {
    return null;
  }

  if (isEqual(funding.nextRecurrence, funding.nextRecurrenceOriginal)) {
    return null;
  }

  return (
    <Typography data-testid='funding-schedule-weekend-notice' size='sm' weight='medium'>
      Actual occurrence avoids weekend ({format(funding.nextRecurrence, 'M/dd')})
    </Typography>
  );
}
