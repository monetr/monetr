import { Fragment, useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { startOfDay, startOfToday } from 'date-fns';
import type { FormikHelpers } from 'formik';

import type { ApiError } from '@monetr/interface/api/client';
import { Button } from '@monetr/interface/components/Button';
import FormAmountField from '@monetr/interface/components/FormAmountField';
import FormButton from '@monetr/interface/components/FormButton';
import FormDatePicker from '@monetr/interface/components/FormDatePicker';
import FormTextField from '@monetr/interface/components/FormTextField';
import MForm from '@monetr/interface/components/MForm';
import MModal, { type MModalRef } from '@monetr/interface/components/MModal';
import MSelectSpending from '@monetr/interface/components/MSelectSpending';
import { Switch } from '@monetr/interface/components/Switch';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@monetr/interface/components/Tabs';
import Typography from '@monetr/interface/components/Typography';
import { type CreateTransactionRequest, useCreateTransaction } from '@monetr/interface/hooks/useCreateTransaction';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import mergeClasses from '@monetr/interface/util/mergeClasses';
import type { APIError } from '@monetr/interface/util/request';
import type { ExtractProps } from '@monetr/interface/util/typescriptEvils';
import { useSnackbar } from '@monetr/notify';

import styles from './NewTransactionModal.module.scss';

interface NewTransactionValues {
  name: string;
  date: Date;
  spendingId: string | null;
  amount: number;
  kind: 'debit' | 'credit';
  // TODO Just keep this false for now, monetr does not allow pending to be modified.
  isPending: boolean;
  adjustsBalance: boolean;
}

function NewTransactionModal(): JSX.Element {
  const { inTimezone } = useTimezone();
  const { data: locale } = useLocaleCurrency();
  const modal = useModal();
  const ref = useRef<MModalRef>(null);
  const { enqueueSnackbar } = useSnackbar();
  const { data: selectedBankAccount } = useSelectedBankAccount();
  const createTransaction = useCreateTransaction();

  const initialValues: NewTransactionValues = {
    name: '',
    date: startOfToday({
      in: inTimezone,
    }),
    amount: 0,
    spendingId: null,
    kind: 'debit',
    isPending: false,
    adjustsBalance: false,
  };

  async function submit(values: NewTransactionValues, helper: FormikHelpers<NewTransactionValues>): Promise<void> {
    const newTransactionRequest: CreateTransactionRequest = {
      bankAccountId: selectedBankAccount.bankAccountId,
      amount: locale.friendlyToAmount(values.kind === 'credit' ? values.amount * -1 : values.amount),
      name: values.name,
      merchantName: null,
      date: startOfDay(new Date(values.date), {
        in: inTimezone,
      }),
      isPending: values.isPending,
      spendingId: values.spendingId,
      adjustsBalance: values.adjustsBalance,
    };

    helper.setSubmitting(true);

    return (
      createTransaction(newTransactionRequest)
        // TODO Show toast that the transaction was created, include button to "view transaction".
        .then(() => modal.remove())
        .catch(
          (error: ApiError<APIError>) =>
            void enqueueSnackbar(error.response.data.error, {
              variant: 'error',
              disableWindowBlurListener: true,
            }),
        )
        .finally(() => helper.setSubmitting(false))
    );
  }

  return (
    <MModal className={styles.modal} open={modal.visible} ref={ref}>
      <MForm className={styles.form} initialValues={initialValues} onSubmit={submit}>
        {({ setFieldValue, values }) => (
          <Fragment>
            <div className={styles.body}>
              <Typography className={styles.heading} size='xl' weight='bold'>
                Create A New Transaction
              </Typography>

              {/* 
              TODO I'm like 99% sure there is going to be a bug here where someone could do something like select a
              spending ID while on the debit tab, then switch to the credit tab and create a deposit with a spending ID?
              */}

              <Tabs
                className={styles.tabs}
                defaultValue='debit'
                onValueChange={value => setFieldValue('kind', value as unknown)}
              >
                <TabsList className={styles.fullWidth}>
                  <TabsTrigger className={styles.fullWidth} value='debit'>
                    Debit
                  </TabsTrigger>
                  <TabsTrigger className={styles.fullWidth} value='credit'>
                    Credit
                  </TabsTrigger>
                </TabsList>
                <TabsContent value='debit'>
                  <FormTextField
                    autoComplete='off'
                    autoFocus
                    data-1p-ignore
                    label='Name / Description'
                    name='name'
                    placeholder='Amazon, Netflix...'
                    required
                  />
                  <div className={styles.fieldRow}>
                    <FormAmountField
                      allowNegative={false}
                      className={styles.fieldRowItem}
                      label='Amount'
                      name='amount'
                      required
                    />
                    <FormDatePicker className={styles.fieldRowItem} label='Date' name='date' required />
                  </div>
                  <MSelectSpending className={styles.spendingSelect} name='spendingId' />
                  <div className={mergeClasses(styles.optionRow, styles.optionRowSpaced)}>
                    <div className={styles.optionText}>
                      <label className={mergeClasses(styles.optionLabel, styles.optionLabelClickable)}>
                        Adjust Balance
                      </label>
                      <p className={styles.optionDescription}>Update your account balance for this transaction?</p>
                    </div>
                    <Switch
                      checked={values.adjustsBalance}
                      onCheckedChange={() => setFieldValue('adjustsBalance', !values.adjustsBalance)}
                    />
                  </div>
                </TabsContent>
                <TabsContent value='credit'>
                  <FormTextField
                    autoComplete='off'
                    data-1p-ignore
                    label='Name / Description'
                    name='name'
                    placeholder='Paycheck, Deposit...'
                    required
                  />
                  <div className={styles.fieldRow}>
                    <FormAmountField
                      allowNegative={false}
                      className={styles.fieldRowItem}
                      label='Amount'
                      name='amount'
                      required
                    />
                    <FormDatePicker className={styles.fieldRowItem} label='Date' name='date' required />
                  </div>
                  <div className={styles.optionRow}>
                    <div className={styles.optionText}>
                      <label className={styles.optionLabel}>Adjust Balance</label>
                      <p className={styles.optionDescription}>Update your account balance for this transaction?</p>
                    </div>
                    <Switch
                      checked={values.adjustsBalance}
                      onCheckedChange={() => setFieldValue('adjustsBalance', !values.adjustsBalance)}
                    />
                  </div>
                </TabsContent>
              </Tabs>
            </div>
            <div className={styles.actions}>
              <Button data-testid='close-new-transaction-modal' onClick={modal.remove} variant='secondary'>
                Cancel
              </Button>
              <FormButton type='submit' variant='primary'>
                Create
              </FormButton>
            </div>
          </Fragment>
        )}
      </MForm>
    </MModal>
  );
}

const newTransactionModal = NiceModal.create(NewTransactionModal);

export default newTransactionModal;

export function showNewTransactionModal(): Promise<void> {
  return NiceModal.show<void, ExtractProps<typeof newTransactionModal>, unknown>(newTransactionModal);
}
