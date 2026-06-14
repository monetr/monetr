import { useCallback, useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { type FormikErrors, type FormikHelpers, useFormikContext } from 'formik';
import { ArrowUpDown } from 'lucide-react';

import type { ApiError } from '@monetr/interface/api/client';
import { Button } from '@monetr/interface/components/Button';
import FormAmountField from '@monetr/interface/components/FormAmountField';
import FormButton from '@monetr/interface/components/FormButton';
import type { LabelDecoratorProps } from '@monetr/interface/components/Label';
import MForm from '@monetr/interface/components/MForm';
import MModal, { type MModalRef } from '@monetr/interface/components/MModal';
import MSelectSpending from '@monetr/interface/components/MSelectSpending';
import Typography from '@monetr/interface/components/Typography';
import { useCurrentBalance } from '@monetr/interface/hooks/useCurrentBalance';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useSpendings } from '@monetr/interface/hooks/useSpendings';
import { useTransfer } from '@monetr/interface/hooks/useTransfer';
import { AmountType } from '@monetr/interface/util/amounts';
import type { APIError } from '@monetr/interface/util/request';
import { useSnackbar } from '@monetr/notify';

import styles from './TransferModal.module.scss';

export interface TransferModalProps {
  initialFromSpendingId?: string;
  initialToSpendingId?: string;
}

interface TransferValues {
  fromSpendingId: string | null;
  toSpendingId: string | null;
  amount: number;
}

function TransferModal(props: TransferModalProps): React.JSX.Element {
  const initialValues: TransferValues = {
    fromSpendingId: props.initialFromSpendingId ?? null,
    toSpendingId: props.initialToSpendingId ?? null,
    amount: 0.0,
  };

  const { data: locale } = useLocaleCurrency();
  const modal = useModal();
  const ref = useRef<MModalRef>(null);
  const transfer = useTransfer();
  const { enqueueSnackbar } = useSnackbar();
  const { data: spending } = useSpendings();

  const validate = useCallback(
    (values: TransferValues): FormikErrors<TransferValues> => {
      const errors: FormikErrors<TransferValues> = {};
      // Locale is always present once the currency data has loaded, but guard anyway so we don't try to convert the
      // amount before it is ready.
      if (!locale) {
        return errors;
      }
      const amount = locale.friendlyToAmount(values.amount);

      if (amount <= 0) {
        errors.amount = 'Amount must be greater than zero';
      }

      // If we are moving an allocation out of an existing budget, do not let us overdraw that budget. We can only really
      // overdraw free to use.
      if (values.fromSpendingId !== null) {
        // Otherwise we are moving funds out of an actual budget. Find that budget.
        const from = spending?.find(item => item.spendingId === values.fromSpendingId);
        // And make sure that we are not moving more than that budget has.
        if (from && amount > from.currentAmount) {
          errors.amount = `Cannot move more than is available from ${from.name}`;
        }
      }

      return errors;
    },
    [locale, spending],
  );

  const submit = useCallback(
    async (values: TransferValues, helper: FormikHelpers<TransferValues>): Promise<void> => {
      if (values.toSpendingId === null && values.fromSpendingId === null) {
        helper.setFieldError('toSpendingId', 'Must select a destination and a source');
        return Promise.resolve();
      }

      const check = validate(values);
      if (Object.keys(check).length > 0) {
        helper.setErrors(check);
        return Promise.resolve();
      }

      // We need the locale to convert the friendly amount before sending it along, it should always be present by the
      // time we can submit but bail just in case it has not loaded yet.
      if (!locale) {
        return Promise.resolve();
      }

      helper.setSubmitting(true);
      return await transfer({
        fromSpendingId: values.fromSpendingId,
        toSpendingId: values.toSpendingId,
        amount: locale.friendlyToAmount(values.amount),
      })
        .then(() => modal.remove())
        .then(
          () =>
            void enqueueSnackbar('Moved funds allocated successfully', {
              variant: 'success',
              disableWindowBlurListener: true,
            }),
        )
        .catch(
          (error: ApiError<APIError>) =>
            void enqueueSnackbar(error.response.data.error, {
              variant: 'error',
              disableWindowBlurListener: true,
            }),
        )
        .finally(() => helper.setSubmitting(false));
    },
    [enqueueSnackbar, locale, modal, transfer, validate],
  );

  return (
    <MModal className={styles.modal} open={modal.visible} ref={ref}>
      <MForm
        className={styles.form}
        data-testid='transfer-modal'
        initialValues={initialValues}
        onSubmit={submit}
        validate={validate}
      >
        <div className={styles.body}>
          <div className={styles.header}>
            <Typography size='2xl' weight='semibold'>
              Transfer
            </Typography>
            <Typography color='subtle' size='lg' weight='medium'>
              Move funds between your budgets
            </Typography>
          </div>
          <MSelectSpending
            excludeFrom='toSpendingId'
            label='From'
            labelDecorator={TransferSelectDecorator}
            name='fromSpendingId'
          />
          <ReverseTargetsButton />
          <MSelectSpending
            excludeFrom='fromSpendingId'
            label='To'
            labelDecorator={TransferSelectDecorator}
            name='toSpendingId'
          />
          <FormAmountField
            allowNegative={false}
            label='Amount'
            name='amount'
            placeholder='Amount to move...'
            step='0.01'
          />
        </div>
        <div className={styles.actions}>
          <Button data-testid='close-new-expense-modal' onClick={modal.remove} variant='secondary'>
            Cancel
          </Button>
          <FormButton type='submit' variant='primary'>
            Transfer
          </FormButton>
        </div>
      </MForm>
    </MModal>
  );
}

const transferModal = NiceModal.create<TransferModalProps>(TransferModal);

export default transferModal;

export function showTransferModal(props: TransferModalProps): Promise<void> {
  return NiceModal.show(transferModal, props) as Promise<void>;
}

function ReverseTargetsButton(): React.JSX.Element {
  const formik = useFormikContext<TransferValues>();
  const swap = useCallback(() => {
    // Do nothing if we are currently submitting.
    if (formik.isSubmitting) {
      return;
    }

    const { fromSpendingId, toSpendingId, amount } = formik.values;
    formik.setValues({
      fromSpendingId: toSpendingId,
      toSpendingId: fromSpendingId,
      amount: amount,
    });
  }, [formik]);

  return (
    <button className={styles.reverseButton} onClick={swap} type='button'>
      <ArrowUpDown className={styles.reverseIcon} />
    </button>
  );
}

function TransferSelectDecorator(props: LabelDecoratorProps): React.JSX.Element | null {
  const formik = useFormikContext<TransferValues>();
  // The decorator is wired up to one of the spending id fields by name, so look that field up dynamically.
  const value = props.name ? formik.values[props.name as keyof TransferValues] : undefined;
  const { data: spending } = useSpendings();
  const { data: balances } = useCurrentBalance();

  // If we are working with the free to use amount.
  if (!value || value === -1) {
    const amount = balances?.free;

    return <AmountButton amount={amount} />;
  }

  // If we aren't dealing with the free to use, then we are working with a spending item. Find it and find out what its
  // amounts are.
  const spendingSubject = spending?.find(item => item.spendingId === value);
  if (!spendingSubject) {
    return null;
  }

  const current = spendingSubject.currentAmount;
  const target = spendingSubject.targetAmount;
  const remaining = Math.max(spendingSubject.targetAmount - spendingSubject.currentAmount, 0);

  if (remaining > 0 && remaining !== target) {
    return (
      <Typography className={styles.decorator} size='inherit'>
        <AmountButton amount={current} />
        of
        <AmountButton amount={target} />
        &nbsp; (<AmountButton amount={remaining} />)
      </Typography>
    );
  }

  return (
    <Typography className={styles.decorator} color='subtle' size='inherit'>
      <AmountButton amount={current} />
      of
      <AmountButton amount={target} />
    </Typography>
  );
}

interface AmountButtonProps {
  amount: number | null | undefined;
}

function AmountButton({ amount }: AmountButtonProps): React.JSX.Element {
  const { data: locale } = useLocaleCurrency();
  const formik = useFormikContext<TransferValues>();
  const onClick = useCallback(() => {
    if (typeof amount === 'number' && locale) {
      formik?.setFieldValue('amount', locale.amountToFriendly(amount));
    }
  }, [amount, formik, locale]);

  return (
    <Typography className={styles.amountButton} onClick={onClick} size='sm' weight='medium'>
      {typeof amount === 'number' && locale && locale.formatAmount(amount, AmountType.Stored)}
    </Typography>
  );
}
