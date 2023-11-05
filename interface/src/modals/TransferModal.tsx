import React, { useCallback, useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { SwapVertOutlined } from '@mui/icons-material';
import { AxiosError } from 'axios';
import { FormikErrors, FormikHelpers, useFormikContext } from 'formik';
import { useSnackbar } from 'notistack';

import MAmountField from '@monetr/interface/components/MAmountField';
import MFormButton from '@monetr/interface/components/MButton';
import MForm from '@monetr/interface/components/MForm';
import { MLabelDecoratorProps } from '@monetr/interface/components/MLabel';
import MModal, { MModalRef } from '@monetr/interface/components/MModal';
import MSelectSpending from '@monetr/interface/components/MSelectSpending';
import MSpan from '@monetr/interface/components/MSpan';
import { useCurrentBalance } from '@monetr/interface/hooks/balances';
import { useSpendingSink, useTransfer } from '@monetr/interface/hooks/spending';
import { amountToFriendly, formatAmount, friendlyToAmount } from '@monetr/interface/util/amounts';
import { ExtractProps } from '@monetr/interface/util/typescriptEvils';

export interface TransferModalProps {
  initialFromSpendingId?: number;
  initialToSpendingId?: number;
}

interface TransferValues {
  fromSpendingId: number | null;
  toSpendingId: number | null;
  amount: number;
}

function TransferModal(props: TransferModalProps): JSX.Element {
  const modal = useModal();
  const ref = useRef<MModalRef>(null);
  const transfer = useTransfer();
  const { enqueueSnackbar } = useSnackbar();
  const balance = useCurrentBalance();
  const { result: spending } = useSpendingSink();

  const initialValues: TransferValues = {
    fromSpendingId: props.initialFromSpendingId,
    toSpendingId: props.initialToSpendingId,
    amount: 0.00,
  };

  function validate(values: TransferValues): FormikErrors<TransferValues> {
    const errors: FormikErrors<TransferValues> = {};
    const amount = friendlyToAmount(values.amount);

    if (amount <= 0) {
      errors['amount'] = 'Amount must be greater than zero';
    }

    // If the source is null that means it is the free to use budget.
    if (values.fromSpendingId === null) {
      // But don't let them move more than is available in that budget.
      if (amount > balance?.free) {
        errors['amount'] = 'Cannot move more than is available from Free-To-Use';
      }
    } else {
      // Otherwise we are moving funds out of an actual budget. Find that budget.
      const from = spending?.find(item => item.spendingId === values.fromSpendingId);
      // And make sure that we are not moving more than that budget has.
      if (amount > from?.currentAmount) {
        errors['amount'] = `Cannot move more than is available from ${from?.name}`;
      }
    }

    return errors;
  }

  async function submit(values: TransferValues, helper: FormikHelpers<TransferValues>): Promise<void> {
    if (values.toSpendingId === null && values.fromSpendingId === null) {
      helper.setFieldError('toSpendingId', 'Must select a destination and a source');
      return Promise.resolve();
    }

    const check = validate(values);
    if (Object.keys(check).length > 0) {
      helper.setErrors(check);
      return Promise.resolve();
    }

    helper.setSubmitting(true);
    return transfer({
      fromSpendingId: values.fromSpendingId,
      toSpendingId: values.toSpendingId,
      amount: friendlyToAmount(values.amount),
    })
      .then(() => modal.remove())
      .catch((error: AxiosError) => void enqueueSnackbar(
        error.response.data['error'], {
          variant: 'error',
          disableWindowBlurListener: true,
        }))
      .finally(() => helper.setSubmitting(false));
  }

  return (
    <MModal open={ modal.visible } ref={ ref } className='md:max-w-sm'>
      <MForm
        onSubmit={ submit }
        initialValues={ initialValues }
        validate={ validate }
        className='h-full flex flex-col gap-2 p-2 justify-between'
        data-testid='transfer-modal'
      >
        <div className='flex flex-col gap-2'>
          <div className='flex flex-col items-center'>
            <MSpan size='2xl' weight='semibold'>
              Transfer
            </MSpan>
            <MSpan size='lg' weight='medium' color='subtle'>
              Move funds between your budgets
            </MSpan>
          </div>
          <MSelectSpending
            excludeFrom='toSpendingId'
            label='From'
            labelDecorator={ TransferSelectDecorator }
            menuPortalTarget={ document.body }
            name="fromSpendingId"
          />
          <ReverseTargetsButton />
          <MSelectSpending
            excludeFrom='fromSpendingId'
            label='To'
            labelDecorator={ TransferSelectDecorator }
            menuPortalTarget={ document.body }
            name="toSpendingId"
          />
          <MAmountField
            name='amount'
            label='Amount'
            placeholder='Amount to move...'
            step='0.01'
            allowNegative={ false }
          />
        </div>
        <div className='flex justify-end gap-2'>
          <MFormButton
            color='secondary'
            onClick={ modal.remove }
            data-testid='close-new-expense-modal'
          >
            Cancel
          </MFormButton>
          <MFormButton
            color='primary'
            type='submit'
          >
            Transfer
          </MFormButton>
        </div>
      </MForm>
    </MModal>
  );
}

const transferModal = NiceModal.create<TransferModalProps>(TransferModal);

export default transferModal;

export function showTransferModal(props: TransferModalProps): Promise<void> {
  return NiceModal.show<void, ExtractProps<typeof transferModal>, {}>(transferModal, props);
}

function ReverseTargetsButton(): JSX.Element {
  const formik = useFormikContext<TransferValues>();

  function onClick() {
    // Do nothing if we are currently submitting.
    if (formik.isSubmitting) return;

    const { fromSpendingId, toSpendingId, amount } = formik.values;
    formik.setValues({
      fromSpendingId: toSpendingId,
      toSpendingId: fromSpendingId,
      amount: amount,
    });
  }

  return (
    <a className='w-full flex justify-center mb-1'>
      <SwapVertOutlined
        onClick={ onClick }
        className='cursor-pointer text-4xl dark:text-dark-monetr-content-subtle hover:dark:text-dark-monetr-content'
      />
    </a>
  );
}

function TransferSelectDecorator(props: MLabelDecoratorProps): JSX.Element {
  const formik = useFormikContext<TransferValues>();
  const value = formik.values[props.name];
  const { result: spending } = useSpendingSink();
  const balances = useCurrentBalance();

  // If we are working with the free to use amount.
  if (!value || value === -1) {
    const amount = balances?.free;

    return (
      <AmountButton amount={ amount } />
    );
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

  if (remaining > 0 && remaining != target) {
    return (
      <MSpan className='gap-1'>
        <AmountButton amount={ current } />
        of
        <AmountButton amount={ target } />
        &nbsp;
        (<AmountButton amount={ remaining } />)
      </MSpan>
    );
  }

  return (
    <MSpan className='gap-1' color='subtle'>
      <AmountButton amount={ current } />
      of
      <AmountButton amount={ target } />
    </MSpan>
  );
}

interface AmountButtonProps {
  amount: number | null | undefined;
}

function AmountButton({ amount }: AmountButtonProps): JSX.Element {
  const formik = useFormikContext<TransferValues>();
  const onClick = useCallback(() => {
    if (typeof amount === 'number') {
      formik?.setFieldValue('amount', amountToFriendly(amount));
    }
  }, [formik, amount]);

  return (
    <MSpan
      size='sm'
      weight='medium'
      className='cursor-pointer hover:dark:text-dark-monetr-content-emphasis'
      onClick={ onClick }
    >
      {typeof amount === 'number' && formatAmount(amount)}
    </MSpan>
  );
}
