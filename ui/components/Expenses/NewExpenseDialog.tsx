import React, { useState } from 'react';
import { DatePicker } from '@mui/lab';
import {
  Alert,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  FormControl,
  Input,
  InputAdornment,
  InputLabel,
  Snackbar,
  Step,
  StepContent,
  StepLabel,
  Stepper,
  TextField,
} from '@mui/material';
import { Formik, FormikErrors, FormikHelpers } from 'formik';
import moment from 'moment';

import FundingScheduleSelectionList from 'components/FundingSchedules/FundingScheduleSelectionList';
import Recurrence from 'components/Recurrence/Recurrence';
import RecurrenceList from 'components/Recurrence/RecurrenceList';
import StepperDialogActionButtons, { StepperStep } from 'components/StepperDialogActionButtons';
import { useSelectedBankAccountId } from 'hooks/bankAccounts';
import { useCreateSpending } from 'hooks/spending';
import Spending, { SpendingType } from 'models/Spending';

interface newExpenseForm {
  name: string;
  amount: number;
  nextOccurrence: moment.Moment;
  recurrenceRule: Recurrence;
  fundingScheduleId: number;
}

const initialValues: newExpenseForm = {
  name: '',
  amount: 0.00,
  nextOccurrence: moment().add(1, 'day'),
  recurrenceRule: new Recurrence(),
  fundingScheduleId: 0,
};

interface Props {
  onClose: { (): void };
  isOpen: boolean;
}

export default function NewExpenseDialog(props: Props): JSX.Element {
  enum NewExpenseStep {
    Name,
    Amount,
    Date,
    Recurrence,
    Funding,
  }
  interface StepState {
    step: NewExpenseStep;
    canNextStep: boolean;
  }

  const selectedBankAccountId = useSelectedBankAccountId();
  const createSpending = useCreateSpending();
  const [stepState, setStepState] = useState<StepState>({
    step: NewExpenseStep.Name,
    canNextStep: true,
  });
  const [error, setError] = useState<string | null>(null);

  function validateInput(_: newExpenseForm): FormikErrors<newExpenseForm> {
    return null;
  }

  async function submit(values: newExpenseForm, { setSubmitting }: FormikHelpers<newExpenseForm>): Promise<void> {
    setSubmitting(true);
    const newSpending = new Spending({
      bankAccountId: selectedBankAccountId,
      name: values.name,
      description: values.recurrenceRule.name,
      nextRecurrence: values.nextOccurrence.startOf('day'),
      spendingType: SpendingType.Expense,
      fundingScheduleId: values.fundingScheduleId,
      targetAmount: Math.ceil(values.amount * 100), // Convert to an integer.
      recurrenceRule: values.recurrenceRule.ruleString(),
    });

    return createSpending(newSpending)
      .then(() => props.onClose())
      .catch(error => setError(error.response.data.error))
      .finally(() => setSubmitting(false));
  }

  function nextStep() {
    return setStepState({
      canNextStep: true,
      step: Math.min(NewExpenseStep.Funding, stepState.step + 1),
    });
  }

  function previousStep() {
    return setStepState({
      canNextStep: true, // Math.min(NewExpenseStep.Name, prevState.step - 1) < prevState.step,
      step: Math.max(NewExpenseStep.Name, stepState.step - 1),
    });
  }

  function ErrorMaybe(): JSX.Element {
    if (!error) {
      return null;
    }

    const onClose = () => setError(null);

    return (
      <Snackbar open autoHideDuration={ 6000 } onClose={ onClose }>
        <Alert onClose={ onClose } severity="error">
          { error }
        </Alert>
      </Snackbar>
    );
  }

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
      }) => (
        <form onSubmit={ handleSubmit }>
          <Dialog open={ props.isOpen } maxWidth="sm">
            <DialogTitle>
              Create a new expense
            </DialogTitle>
            <DialogContent>
              <ErrorMaybe />
              <DialogContentText>
                Expenses let you budget for things that happen on a regular basis automatically. Money is allocated
                to expenses whenever you get paid so that you don't have to pay something from a single paycheck.
              </DialogContentText>
              <div>
                <Stepper activeStep={ stepState.step } orientation="vertical">
                  <Step key="What is your expense for?">
                    <StepLabel>What is your expense for?</StepLabel>
                    <StepContent>
                      <TextField
                        error={ touched.name && !!errors.name }
                        helperText={ (touched.name && errors.name) ? errors.name : null }
                        autoFocus
                        id="new-expense-name"
                        name="name"
                        className="w-full"
                        label="Name"
                        onChange={ handleChange }
                        onBlur={ handleBlur }
                        value={ values.name }
                        disabled={ isSubmitting }
                      />
                    </StepContent>
                  </Step>
                  <Step key="How much do you need?">
                    <StepLabel>How much do you need?</StepLabel>
                    <StepContent>
                      <FormControl fullWidth>
                        <InputLabel htmlFor="new-expense-amount">Amount</InputLabel>
                        <Input
                          id="new-expense-amount"
                          name="amount"
                          value={ values.amount }
                          onBlur={ handleBlur }
                          onChange={ handleChange }
                          disabled={ isSubmitting }
                          startAdornment={ <InputAdornment position="start">$</InputAdornment> }
                        />
                      </FormControl>
                    </StepContent>
                  </Step>
                  <Step key="When do you need it next?">
                    <StepLabel>When do you need it next?</StepLabel>
                    <StepContent>
                      <DatePicker
                        minDate={ moment().startOf('day').add(1, 'day') }
                        onChange={ value => setFieldValue('nextOccurrence', value.startOf('day')) }
                        inputFormat="MM/DD/yyyy"
                        value={ values.nextOccurrence }
                        renderInput={ params => <TextField fullWidth { ...params } /> }
                      />
                    </StepContent>
                  </Step>
                  <Step key="How frequently do you need it?">
                    <StepLabel>How frequently do you need it?</StepLabel>
                    <StepContent>
                      { (stepState.step === NewExpenseStep.Recurrence || values.nextOccurrence) &&
                        <RecurrenceList
                          disabled={ isSubmitting }
                          date={ values.nextOccurrence }
                          onChange={ value => setFieldValue('recurrenceRule', value) }
                        />
                      }
                    </StepContent>
                  </Step>
                  <Step key="How do you want to fund it?">
                    <StepLabel>How do you want to fund it?</StepLabel>
                    <StepContent>
                      <div className="mt-5" />
                      <FundingScheduleSelectionList
                        disabled={ isSubmitting }
                        onChange={ value => setFieldValue('fundingScheduleId', value.fundingScheduleId) }
                      />
                    </StepContent>
                  </Step>
                </Stepper>
              </div>
            </DialogContent>
            <DialogActions>
              <StepperDialogActionButtons
                isSubmitting={ isSubmitting }
                submitForm={ submitForm }
                onClose={ props.onClose }
                previousStep={ previousStep }
                nextStep={ nextStep }
                canNextStep={ stepState.canNextStep }
                step={ stepState.step === NewExpenseStep.Name ? StepperStep.First :
                  stepState.step === NewExpenseStep.Funding ? StepperStep.Last : StepperStep.Other }
              />
            </DialogActions>
          </Dialog>
        </form>
      ) }
    </Formik>
  );
}
