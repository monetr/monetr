import React, { useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { DatePicker } from '@mui/lab';
import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, Divider, InputAdornment, TextField } from '@mui/material';
import { AxiosError } from 'axios';
import { Formik, FormikErrors, FormikHelpers } from 'formik';
import moment from 'moment';
import { useSnackbar } from 'notistack';

import FundingScheduleSelect from 'components/FundingSchedules/FundingScheduleSelect';
import Recurrence from 'components/Recurrence/Recurrence';
import RecurrenceSelect from 'components/Recurrence/RecurrenceSelect';
import { useSelectedBankAccountId } from 'hooks/bankAccounts';
import { useCreateSpending } from 'hooks/spending';
import Spending, { SpendingType } from 'models/Spending';

interface CreateExpenseForm {
  name: string;
  amount: number;
  nextOccurrence: moment.Moment;
  recurrenceRule: Recurrence;
  fundingScheduleId: number;
}

function CreateExpenseDialog(): JSX.Element {
  const modal = useModal();
  const selectedBankAccountId = useSelectedBankAccountId();
  const createSpending = useCreateSpending();
  const { enqueueSnackbar } = useSnackbar();

  const ref = useRef<HTMLDivElement>(null);

  function validateInput(input: CreateExpenseForm): FormikErrors<CreateExpenseForm> {
    const errors: FormikErrors<CreateExpenseForm> = {};

    if (input.name.trim().length > 120) {
      errors['name'] = 'Cannot be longer than 120 characters.';
    }

    if (input.amount <= 0) {
      errors['amount'] = 'Must be greater than 0.';
    }

    if ((input.amount % 1) != 0 && input.amount.toString().split('.')[1].length > 2) {
      errors['amount'] = 'Can only have up to 2 decimal places.';
    }

    return errors;
  }

  async function submit(values: CreateExpenseForm, helper: FormikHelpers<CreateExpenseForm>): Promise<void> {
    if (values.name.trim().length === 0) {
      helper.setFieldError('name', 'Required to create an expense.');
      return Promise.reject();
    }

    helper.setSubmitting(true);
    const newSpending = new Spending({
      bankAccountId: selectedBankAccountId,
      name: values.name.trim(),
      description: values.recurrenceRule.name.trim(),
      nextRecurrence: values.nextOccurrence.startOf('day'),
      spendingType: SpendingType.Expense,
      fundingScheduleId: values.fundingScheduleId,
      targetAmount: Math.ceil(values.amount * 100), // Convert to an integer.
      recurrenceRule: values.recurrenceRule.ruleString(),
    });

    return createSpending(newSpending)
      .then(() => modal.remove())
      .catch((error: AxiosError) => void enqueueSnackbar(error.response.data['error'], {
        variant: 'error',
        disableWindowBlurListener: true,
      }))
      .finally(() => helper.setSubmitting(false));
  }

  const initialValues: CreateExpenseForm = {
    name: '',
    amount: 0.00,
    nextOccurrence: moment().add(1, 'day'),
    recurrenceRule: new Recurrence(),
    fundingScheduleId: 0,
  };

  return (
    <Formik
      initialValues={ initialValues }
      validate={ validateInput }
      onSubmit={ submit }
    >
      { ({
        values,
        errors,
        touched,
        handleChange,
        handleBlur,
        handleSubmit,
        setFieldValue,
        isSubmitting,
        submitForm,
        isValid,
      }) => (
        <form onSubmit={ handleSubmit }>
          <Dialog open={ modal.visible } maxWidth="sm" ref={ ref }>
            <DialogTitle>
              Create A New Expense
            </DialogTitle>
            <DialogContent>
              <DialogContentText>
                Expenses let you budget for things that happen on a regular basis automatically. Money is allocated
                to expenses whenever you get paid so that you don't have to pay something from a single paycheck.
              </DialogContentText>
              <div className='grid sm:grid-cols-12 md:grid-cols-12 mt-5 md:gap-x-5 md:gap-y-5 gap-y-2'>
                <div className='col-span-12'>
                  <span className='font-normal ml-3'>
                    What are you budgeting for?
                  </span>
                  <TextField
                    error={ touched.name && !!errors.name }
                    placeholder="Example: Amazon..."
                    helperText={ (touched.name && errors.name) ? errors.name : ' ' }
                    autoFocus
                    name="name"
                    className="w-full"
                    onChange={ handleChange }
                    onBlur={ handleBlur }
                    value={ values.name }
                    disabled={ isSubmitting }
                    required
                  />
                </div>
                <div className='col-span-12 md:col-span-6'>
                  <span className='font-normal ml-3'>
                    How much do you need?
                  </span>
                  <TextField
                    error={ touched.amount && !!errors.amount }
                    helperText={ (touched.amount && errors.amount) ? errors.amount : ' ' }
                    name="amount"
                    className="w-full"
                    type="number"
                    onChange={ handleChange }
                    onBlur={ handleBlur }
                    value={ values.amount }
                    disabled={ isSubmitting }
                    required
                    InputProps={ {
                      startAdornment: <InputAdornment position="start">$</InputAdornment>,
                      inputProps: { min: 0 },
                    } }
                  />
                </div>
                <div className='col-span-12 md:col-span-6'>
                  <span className='font-normal ml-3'>
                    When do you need it next?
                  </span>
                  <DatePicker
                    disabled={ isSubmitting }
                    minDate={ moment().startOf('day').add(1, 'day') }
                    onChange={ value => setFieldValue('nextOccurrence', value.startOf('day')) }
                    inputFormat="MM/DD/yyyy"
                    value={ values.nextOccurrence }
                    renderInput={ params => (
                      <TextField label="When do you need it next?"  fullWidth { ...params } />
                    ) }
                  />
                </div>
                <Divider className='col-span-12 mt-4' />
                <div className='col-span-12'>
                  <span className='font-normal ml-3'>
                    How often do you need to pay for { values.name || 'your expense' }?
                  </span>
                  <RecurrenceSelect
                    menuRef={ ref.current }
                    disabled={ isSubmitting }
                    date={ values.nextOccurrence }
                    onChange={ value => setFieldValue('recurrenceRule', value) }
                  />
                </div>
                <div className='col-span-12'>
                  <span className='font-normal ml-3'>
                    How do you want to fund your expense?
                  </span>
                  <FundingScheduleSelect
                    className='w-full'
                    menuRef={ ref.current }
                    disabled={ isSubmitting }
                    onChange={ value => setFieldValue('fundingScheduleId', value) }
                    value={ values.fundingScheduleId }
                  />
                </div>
              </div>
            </DialogContent>
            <DialogActions>
              <Button
                color="secondary"
                disabled={ isSubmitting }
                onClick={ modal.remove }
              >
                Cancel
              </Button>
              <Button
                disabled={ isSubmitting || !isValid }
                onClick={ submitForm }
                color="primary"
                type="submit"
              >
                Create
              </Button>
            </DialogActions>
          </Dialog>
        </form>
      ) }
    </Formik>
  );
}

const createExpenseModal = NiceModal.create(CreateExpenseDialog);

export default createExpenseModal;

export function showCreateExpenseDialog(): void {
  NiceModal.show(createExpenseModal);
}
