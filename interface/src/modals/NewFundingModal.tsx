import { useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { startOfDay, startOfTomorrow } from 'date-fns';
import { type FormikHelpers, useFormikContext } from 'formik';
import { Repeat } from 'lucide-react';

import type { ApiError } from '@monetr/interface/api/client';
import { Button } from '@monetr/interface/components/Button';
import FormAmountField from '@monetr/interface/components/FormAmountField';
import FormButton from '@monetr/interface/components/FormButton';
import FormDatePicker from '@monetr/interface/components/FormDatePicker';
import FormSwitch from '@monetr/interface/components/FormSwitch';
import FormTextField from '@monetr/interface/components/FormTextField';
import MForm from '@monetr/interface/components/MForm';
import MModal, { type MModalRef } from '@monetr/interface/components/MModal';
import MSelectFrequency from '@monetr/interface/components/MSelectFrequency';
import Typography from '@monetr/interface/components/Typography';
import { useCreateFundingSchedule } from '@monetr/interface/hooks/useCreateFundingSchedule';
import { useCurrentLink } from '@monetr/interface/hooks/useCurrentLink';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { usePatchTransactionRecurring } from '@monetr/interface/hooks/usePatchTransactionRecurring';
import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { useTransactionCluster } from '@monetr/interface/hooks/useTransactionCluster';
import type FundingSchedule from '@monetr/interface/models/FundingSchedule';
import type Transaction from '@monetr/interface/models/Transaction';
import type TransactionRecurring from '@monetr/interface/models/TransactionRecurring';
import type { APIError } from '@monetr/interface/util/request';
import { useSnackbar } from '@monetr/notify';

import styles from './NewFundingModal.module.scss';

interface NewFundingValues {
  name: string;
  nextOccurrence: Date;
  ruleset: string;
  excludeWeekends: boolean;
  estimatedDeposit?: number | null;
  autoCreateTransaction: boolean;
}

export interface NewFundingModalProps {
  // recurring fills in the new funding schedule from a recurring deposit, like a paycheck
  recurring?: TransactionRecurring;
  // transaction is the one the user started from, its name is the fallback if there's no cluster name
  transaction?: Transaction;
}

function NewFundingModal(props: NewFundingModalProps): React.JSX.Element {
  const { inTimezone } = useTimezone();
  const modal = useModal();
  const ref = useRef<MModalRef>(null);
  const { enqueueSnackbar } = useSnackbar();
  const selectedBankAccountId = useSelectedBankAccountId();
  const createFundingSchedule = useCreateFundingSchedule();
  const patchTransactionRecurring = usePatchTransactionRecurring();
  const { data: link } = useCurrentLink();
  const isManual = Boolean(link?.getIsManual());
  const { data: locale } = useLocaleCurrency();
  // Use the similar transactions name instead of the transaction's own name, since that's the name for the whole group
  // of deposits and not just the one they happened to start from
  const { data: cluster, isLoading: clusterIsLoading } = useTransactionCluster(
    props.recurring?.transactionClusterId ?? null,
  );
  const name = cluster?.name || props.transaction?.getName();

  // The form only reads its initial values once, so wait for everything the recurring deposit fills in
  if (props.recurring && (!locale || clusterIsLoading)) {
    return (
      <MModal className={styles.modal} open={modal.visible} ref={ref}>
        One moment...
      </MModal>
    );
  }

  const tomorrow = startOfTomorrow({
    in: inTimezone,
  });
  const initialValues: NewFundingValues = {
    name: name ?? '',
    nextOccurrence: props.recurring?.getNextRecurrence(tomorrow) ?? tomorrow,
    ruleset: props.recurring?.ruleset ?? '',
    excludeWeekends: false,
    estimatedDeposit:
      props.recurring && locale ? locale.amountToFriendly(Math.abs(props.recurring.lastAmount)) : undefined,
    autoCreateTransaction: false,
  };

  async function submit(values: NewFundingValues, helpers: FormikHelpers<NewFundingValues>): Promise<void> {
    if (!selectedBankAccountId || !locale) {
      return Promise.resolve();
    }

    helpers.setSubmitting(true);
    const estimatedDeposit = values.estimatedDeposit ?? 0;
    return await createFundingSchedule({
      bankAccountId: selectedBankAccountId,
      name: values.name,
      description: null,
      nextRecurrence: startOfDay(new Date(values.nextOccurrence), {
        in: inTimezone,
      }),
      ruleset: values.ruleset,
      estimatedDeposit: estimatedDeposit > 0 ? locale.friendlyToAmount(estimatedDeposit) : null,
      excludeWeekends: values.excludeWeekends,
      // Auto create transaction requires a manual link and a non-zero
      // estimated deposit; force it off otherwise so the API will not reject
      // the create.
      autoCreateTransaction: isManual && estimatedDeposit > 0 && values.autoCreateTransaction,
    })
      .then(async created => {
        // The funding schedule is linked to the recurring deposit with a second request. If that fails the funding
        // schedule is still created, so keep it and just let the user know it isn't linked
        if (props.recurring) {
          await patchTransactionRecurring({
            transactionRecurringId: props.recurring.transactionRecurringId,
            bankAccountId: props.recurring.bankAccountId,
            fundingScheduleId: created.fundingScheduleId,
          }).catch(
            () =>
              void enqueueSnackbar(
                'Your funding schedule was created, but it could not be linked to the recurring deposit.',
                {
                  variant: 'warning',
                  disableWindowBlurListener: true,
                },
              ),
          );
        }
        return created;
      })
      .then(created => modal.resolve(created))
      .then(() => modal.remove())
      .catch(
        (error: ApiError<APIError>) =>
          void enqueueSnackbar(error.response.data.error, {
            variant: 'error',
            disableWindowBlurListener: true,
          }),
      )
      .finally(() => helpers.setSubmitting(false));
  }

  return (
    <MModal className={styles.modal} open={modal.visible} ref={ref}>
      <MForm className={styles.form} data-testid='new-funding-modal' initialValues={initialValues} onSubmit={submit}>
        <div className={styles.body}>
          <Typography className={styles.heading} size='xl' weight='bold'>
            Create A New Funding Schedule
          </Typography>
          {props.recurring && (
            <div className={styles.recurringBanner} data-testid='new-funding-recurring-banner'>
              <Repeat />
              <Typography color='inherit' size='sm'>
                Filled in from your {name ?? 'recurring'} deposits. Give it a once over before you create it.
              </Typography>
            </div>
          )}
          <FormTextField
            autoComplete='off'
            autoFocus
            data-1p-ignore
            label='What do you want to call your funding schedule?'
            name='name'
            placeholder='Example: Payday...'
            required
          />
          <FormDatePicker
            description={props.recurring && 'When the next deposit is expected.'}
            label='When do you get paid next?'
            min={startOfTomorrow({
              in: inTimezone,
            })}
            name='nextOccurrence'
            required
          />
          <MSelectFrequency
            dateFrom='nextOccurrence'
            description={props.recurring && `Matches when ${name ?? 'it'} pays you.`}
            label='How often do you get paid?'
            name='ruleset'
            placeholder='Select a funding frequency...'
            required
          />
          <FormAmountField
            allowNegative={false}
            description={props.recurring && 'Your last deposit.'}
            label='Estimated Deposit'
            name='estimatedDeposit'
            placeholder='Example: $ 1,000.00'
          />
          <FormSwitch
            description='If it were to land on a weekend, it is adjusted to the previous weekday instead.'
            label='Exclude Weekends'
            name='excludeWeekends'
          />
          {isManual && <AutoCreateTransactionToggle />}
        </div>
        <div className={styles.actions}>
          <Button data-testid='close-new-funding-modal' onClick={modal.remove} variant='secondary'>
            Cancel
          </Button>
          <FormButton type='submit' variant='primary'>
            Create
          </FormButton>
        </div>
      </MForm>
    </MModal>
  );
}

function AutoCreateTransactionToggle(): React.JSX.Element {
  const { values } = useFormikContext<NewFundingValues>();
  const hasDeposit = (values.estimatedDeposit ?? 0) > 0;

  return (
    <FormSwitch
      checked={hasDeposit && values.autoCreateTransaction}
      data-testid='new-funding-auto-create-transaction'
      description='Automatically add a deposit transaction for the estimated deposit each time the funding schedule would occur.'
      disabled={!hasDeposit}
      label='Auto create transaction'
      name='autoCreateTransaction'
    />
  );
}

const newFundingModal = NiceModal.create(NewFundingModal);

export default newFundingModal;

export function showNewFundingModal(props: NewFundingModalProps = {}): Promise<FundingSchedule | null> {
  return NiceModal.show(newFundingModal, props) as Promise<FundingSchedule | null>;
}
