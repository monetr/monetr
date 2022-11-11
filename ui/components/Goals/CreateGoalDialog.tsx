import useIsMobile from 'hooks/useIsMobile';
import React, { useRef } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { DatePicker } from '@mui/lab';
import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, Divider, InputAdornment, TextField } from '@mui/material';
import { AxiosError } from 'axios';
import { Formik, FormikErrors, FormikHelpers } from 'formik';
import moment from 'moment';
import { useSnackbar } from 'notistack';

import FundingScheduleSelect from 'components/FundingSchedules/FundingScheduleSelect';
import { useSelectedBankAccountId } from 'hooks/bankAccounts';
import { useCreateSpending } from 'hooks/spending';
import Spending, { SpendingType } from 'models/Spending';

interface CreateGoalForm {
  name: string;
  amount: number;
  byDate: moment.Moment;
  fundingScheduleId: number;
}

function CreateGoalDialog(): JSX.Element {
  const modal = useModal();
  const isMobile = useIsMobile();
  const bankAccountId = useSelectedBankAccountId();
  const createSpending = useCreateSpending();
  const { enqueueSnackbar } = useSnackbar();
  const ref = useRef<HTMLDivElement>(null);

  function validateInput(input: CreateGoalForm): FormikErrors<CreateGoalForm> {
    const errors: FormikErrors<CreateGoalForm> = {};

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

  async function submit(values: CreateGoalForm, { setSubmitting }: FormikHelpers<CreateGoalForm>): Promise<void> {
    const newSpending = new Spending({
      bankAccountId: bankAccountId,
      name: values.name,
      description: null,
      nextRecurrence: values.byDate.startOf('day'),
      spendingType: SpendingType.Goal,
      fundingScheduleId: values.fundingScheduleId,
      targetAmount: Math.ceil(values.amount * 100), // Convert to an integer.
      recurrenceRule: null,
    });


    return createSpending(newSpending)
      .then(() => modal.remove())
      .catch((error: AxiosError) => void enqueueSnackbar(error.response.data['error'], {
        variant: 'error',
        disableWindowBlurListener: true,
      }))
      .finally(() => setSubmitting(false));
  }

  const initialValues: CreateGoalForm = {
    name: '',
    amount: 0.00,
    byDate: moment().add(1, 'day'),
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
          <Dialog open={ modal.visible } maxWidth="sm" ref={ ref } fullScreen={ isMobile }>
            <DialogTitle>
              Create A New Goal
            </DialogTitle>
            <DialogContent>
              <DialogContentText>
                Goals let you budget for things that don't repeat on a regular basis. Like saving up for something or
                paying something off. Once a goal reaches its target amount no more contributions to it will be made.
              </DialogContentText>
              <div className='grid sm:grid-cols-12 md:grid-cols-12 mt-5 md:gap-x-5 md:gap-y-5 gap-y-2'>
                <div className='col-span-12'>
                  <span className='font-normal ml-3'>
                    What are you budgeting for?
                  </span>
                  <TextField
                    error={ touched.name && !!errors.name }
                    placeholder="Example: Vacation..."
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
                    When do you want to have it by?
                  </span>
                  <DatePicker
                    disabled={ isSubmitting }
                    minDate={ moment().startOf('day').add(1, 'day') }
                    onChange={ value => setFieldValue('byDate', value.startOf('day')) }
                    inputFormat="MM/DD/yyyy"
                    value={ values.byDate }
                    renderInput={ params => (
                      <TextField label="When do you want to have it by?"  fullWidth { ...params } />
                    ) }
                  />
                </div>
                <Divider className='col-span-12 mt-4' />
                <div className='col-span-12'>
                  <span className='font-normal ml-3'>
                    How do you want to fund your goal?
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
                disabled={ isSubmitting || !isValid || values.fundingScheduleId === null }
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

const createGoalModal = NiceModal.create(CreateGoalDialog);

export default createGoalModal;

export function showCreateGoalDialog(): void {
  NiceModal.show(createGoalModal);
}
