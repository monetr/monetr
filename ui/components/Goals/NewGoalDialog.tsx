import React, { Fragment, useState } from 'react';
import { DatePicker } from '@mui/lab';
import {
  Alert,
  Button,
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
import { Formik, FormikHelpers } from 'formik';
import moment from 'moment';

import FundingScheduleSelectionList from 'components/FundingSchedules/FundingScheduleSelectionList';
import { useSelectedBankAccountId } from 'hooks/bankAccounts';
import { useCreateSpending } from 'hooks/spending';
import Spending, { SpendingType } from 'models/Spending';

enum DialogStep {
  Name,
  Amount,
  Date,
  Funding,
}

interface NewGoalForm {
  name: string;
  amount: number;
  byDate: moment.Moment;
  fundingScheduleId: number;
}

const initialValues: NewGoalForm = {
  name: '',
  amount: 0.00,
  byDate: moment().add(1, 'day'),
  fundingScheduleId: 0,
};

interface Props {
  onClose: { (): void };
  isOpen: boolean;
}

export default function NewGoalDialog(props: Props): JSX.Element {
  const [step, setStep] = useState<DialogStep>(DialogStep.Name);
  const [error, setError]  = useState<string|null>(null);
  const bankAccountId = useSelectedBankAccountId();
  const createSpending = useCreateSpending();

  async function submit(values: NewGoalForm, { setSubmitting }: FormikHelpers<NewGoalForm>): Promise<void> {
    const { onClose } = props;

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
      .then(onClose)
      .catch(error => setError(error?.response?.data?.error))
      .finally(() => setSubmitting(false));
  }

  function nextStep() {
    setStep(Math.min(DialogStep.Funding, step + 1));
  }

  function previousStep() {
    setStep(Math.max(DialogStep.Name, step - 1));
  }

  function renderActions(isSubmitting: boolean, submitForm: () => Promise<unknown>) {
    const { onClose } = props;

    const cancelButton = (
      <Button color="secondary" onClick={ onClose }>
        Cancel
      </Button>
    );

    const previousButton = (
      <Button
        disabled={ isSubmitting }
        color="secondary"
        onClick={ previousStep }
      >
        Previous
      </Button>
    );

    const nextButton = (
      <Button
        color="primary"
        onClick={ nextStep }
      >
        Next
      </Button>
    );

    const submitButton = (
      <Button
        disabled={ isSubmitting }
        onClick={ submitForm }
        color="primary"
        type="submit"
      >
        Create
      </Button>
    );

    switch (step) {
      case DialogStep.Name:
        return (
          <Fragment>
            { cancelButton }
            { nextButton }
          </Fragment>
        );
      case DialogStep.Funding:
        return (
          <Fragment>
            { previousButton }
            { submitButton }
          </Fragment>
        );
      default:
        return (
          <Fragment>
            { previousButton }
            { nextButton }
          </Fragment>
        );
    }
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

  const { isOpen } = props;

  return (
    <Formik
      initialValues={ initialValues }
      validate={ () => {
      } }
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
          <Dialog open={ isOpen } maxWidth="sm">
            <DialogTitle>
              Create a new goal
            </DialogTitle>
            <DialogContent>
              <ErrorMaybe />
              <DialogContentText>
                Goals let you save up to something specific over a longer period of time. They do not repeat but
                make it easy to put money aside for something you know you'll need or want later.
              </DialogContentText>
              <div>
                <Stepper activeStep={ step } orientation="vertical">
                  <Step key="What are you saving for?">
                    <StepLabel>
                      What are you saving for?
                    </StepLabel>
                    <StepContent>
                      <TextField
                        error={ touched.name && !!errors.name }
                        helperText={ (touched.name && errors.name) ? errors.name : null }
                        autoFocus
                        id="new-goal-name"
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
                        <InputLabel htmlFor="new-goal-amount">Amount</InputLabel>
                        <Input
                          id="new-goal-amount"
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
                  <Step key="When do you need it by?">
                    <StepLabel>When do you need it by</StepLabel>
                    <StepContent>
                      <DatePicker
                        minDate={ moment().startOf('day').add(1, 'day') }
                        onChange={ value => setFieldValue('byDate', value.startOf('day')) }
                        inputFormat="MM/DD/yyyy"
                        value={ values.byDate }
                        renderInput={ params => <TextField fullWidth { ...params } /> }
                      />
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
              { renderActions(isSubmitting, submitForm) }
            </DialogActions>
          </Dialog>
        </form>
      ) }
    </Formik>
  );
}

