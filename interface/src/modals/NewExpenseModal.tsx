import { useId, useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { startOfDay, startOfTomorrow } from 'date-fns';
import { type FormikHelpers, useFormikContext } from 'formik';

import type { ApiError } from '@monetr/interface/api/client';
import { Button } from '@monetr/interface/components/Button';
import FormAmountField from '@monetr/interface/components/FormAmountField';
import FormButton from '@monetr/interface/components/FormButton';
import FormDatePicker from '@monetr/interface/components/FormDatePicker';
import FormTextField from '@monetr/interface/components/FormTextField';
import MForm from '@monetr/interface/components/MForm';
import MModal, { type MModalRef } from '@monetr/interface/components/MModal';
import MSelectFrequency from '@monetr/interface/components/MSelectFrequency';
import MSelectFunding from '@monetr/interface/components/MSelectFunding';
import { Switch } from '@monetr/interface/components/Switch';
import Typography from '@monetr/interface/components/Typography';
import { useCreateSpending } from '@monetr/interface/hooks/useCreateSpending';
import { useCurrentLink } from '@monetr/interface/hooks/useCurrentLink';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import Spending, { SpendingType } from '@monetr/interface/models/Spending';
import type { APIError } from '@monetr/interface/util/request';
import type { ExtractProps } from '@monetr/interface/util/typescriptEvils';
import { useSnackbar } from '@monetr/notify';

import styles from './NewExpenseModal.module.scss';

interface NewExpenseValues {
  name: string;
  amount: number;
  nextOccurrence: Date;
  ruleset: string;
  fundingScheduleId: string;
  autoCreateTransaction: boolean;
}

function NewExpenseModal(): JSX.Element {
  const { inTimezone } = useTimezone();
  const {
    data: { friendlyToAmount },
  } = useLocaleCurrency();
  const modal = useModal();
  const { enqueueSnackbar } = useSnackbar();
  const { data: selectedBankAccount } = useSelectedBankAccount();
  const { data: link } = useCurrentLink();
  const isManual = Boolean(link?.getIsManual());
  const createSpending = useCreateSpending();

  const ref = useRef<MModalRef>(null);

  if (!selectedBankAccount) {
    return (
      <MModal className={styles.modal} open={modal.visible} ref={ref}>
        One moment...
      </MModal>
    );
  }

  const initialValues: NewExpenseValues = {
    name: '',
    amount: 0.0,
    nextOccurrence: startOfTomorrow({
      in: inTimezone,
    }),
    ruleset: '',
    fundingScheduleId: '',
    autoCreateTransaction: false,
  };

  async function submit(values: NewExpenseValues, helper: FormikHelpers<NewExpenseValues>): Promise<void> {
    const newSpending = new Spending({
      bankAccountId: selectedBankAccount.bankAccountId,
      name: values.name.trim(),
      nextRecurrence: startOfDay(new Date(values.nextOccurrence), {
        in: inTimezone,
      }),
      spendingType: SpendingType.Expense,
      fundingScheduleId: values.fundingScheduleId,
      targetAmount: friendlyToAmount(values.amount),
      ruleset: values.ruleset,
      // Auto create transaction requires a manual link and a non-zero target
      // amount; force it off otherwise so the API will not reject the create.
      autoCreateTransaction: isManual && values.amount > 0 && values.autoCreateTransaction,
    });

    helper.setSubmitting(true);
    return createSpending(newSpending)
      .then(created => modal.resolve(created))
      .then(() => modal.remove())
      .catch(
        (error: ApiError<APIError>) =>
          void enqueueSnackbar(error.response.data.error, {
            variant: 'error',
            disableWindowBlurListener: true,
          }),
      )
      .finally(() => helper.setSubmitting(false));
  }

  return (
    <MModal className={styles.modal} open={modal.visible} ref={ref}>
      <MForm className={styles.form} data-testid='new-expense-modal' initialValues={initialValues} onSubmit={submit}>
        <div className={styles.body}>
          <Typography className={styles.heading} size='xl' weight='bold'>
            Create A New Expense
          </Typography>
          <FormTextField
            autoComplete='off'
            autoFocus
            data-1p-ignore
            label='What are you budgeting for?'
            name='name'
            placeholder='Amazon, Netflix...'
            required
          />
          <div className={styles.fieldRow}>
            <FormAmountField
              allowNegative={false}
              className={styles.fieldRowItem}
              label='How much do you need?'
              name='amount'
              required
            />
            <FormDatePicker
              className={styles.fieldRowItem}
              label='When do you need it next?'
              min={startOfTomorrow({
                in: inTimezone,
              })}
              name='nextOccurrence'
              required
            />
          </div>
          <MSelectFunding
            label='When do you want to fund the expense?'
            menuPortalTarget={document.body}
            name='fundingScheduleId'
            required
          />
          <MSelectFrequency
            dateFrom='nextOccurrence'
            label='How frequently do you need this expense?'
            name='ruleset'
            placeholder='Select a spending frequency...'
            required
          />
          {isManual && <AutoCreateTransactionToggle />}
        </div>
        <div className={styles.actions}>
          <Button data-testid='close-new-expense-modal' onClick={modal.remove} variant='secondary'>
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

function AutoCreateTransactionToggle(): JSX.Element {
  const autoCreateSwitchId = useId();
  const { setFieldValue, values } = useFormikContext<NewExpenseValues>();
  const hasAmount = (values.amount ?? 0) > 0;

  return (
    <div className={styles.optionRow} data-testid='new-expense-auto-create-transaction'>
      <div className={styles.optionText}>
        <label aria-disabled={!hasAmount} className={styles.optionLabel} htmlFor={autoCreateSwitchId}>
          Auto create transaction
        </label>
        <p aria-disabled={!hasAmount} className={styles.optionDescription}>
          Automatically add a transaction for this expense each time it is due, deducting from your balance.
        </p>
      </div>
      <Switch
        checked={hasAmount && values.autoCreateTransaction}
        disabled={!hasAmount}
        id={autoCreateSwitchId}
        onCheckedChange={() => setFieldValue('autoCreateTransaction', !values.autoCreateTransaction)}
      />
    </div>
  );
}

const newExpenseModal = NiceModal.create(NewExpenseModal);

export default newExpenseModal;

export function showNewExpenseModal(): Promise<Spending | null> {
  return NiceModal.show<Spending | null, ExtractProps<typeof newExpenseModal>, unknown>(newExpenseModal);
}
