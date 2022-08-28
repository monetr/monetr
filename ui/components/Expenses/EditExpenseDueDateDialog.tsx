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
  Snackbar,
  Step,
  StepContent,
  StepLabel,
  Stepper,
  TextField,
} from '@mui/material';
import { Formik, FormikErrors, FormikHelpers } from 'formik';
import moment from 'moment';

import Recurrence from 'components/Recurrence/Recurrence';
import RecurrenceList from 'components/Recurrence/RecurrenceList';
import { useUpdateSpending } from 'hooks/spending';
import Spending from 'models/Spending';

interface editSpendingDueDateForm {
  dueDate: moment.Moment;
  recurrenceRule: Recurrence;
}

interface Props {
  spending: Spending;
  onClose: { (): void };
  isOpen: boolean;
}

export default function EditExpenseDueDateDialog(props: Props): JSX.Element {
  enum EditSpendingDueDateStep {
    NextRecurrence,
    Frequency,
  }
  const updateSpending = useUpdateSpending();
  const [currentStep, setCurrentStep] = useState<EditSpendingDueDateStep>(EditSpendingDueDateStep.NextRecurrence);
  const [error, setError] = useState<string | null>(null);

  function validateInput(_: editSpendingDueDateForm): FormikErrors<editSpendingDueDateForm> {
    return null;
  }

  async function submit(
    values: editSpendingDueDateForm,
    { setSubmitting }: FormikHelpers<editSpendingDueDateForm>,
  ): Promise<void> {
    setSubmitting(true);
    const updatedSpending = new Spending({
      ...props.spending,
      nextRecurrence: values.dueDate.startOf('day'),
      recurrenceRule: values.recurrenceRule.ruleString(),
      description: values.recurrenceRule.name,
    });

    return updateSpending(updatedSpending)
      .then(() => props.onClose())
      .catch(error => setError(error.response.data.error))
      .finally(() => setSubmitting(false));
  }

  function Error(): JSX.Element {
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

  function StepActions({
    isSubmitting,
    submitForm,
  }: { isSubmitting: boolean, submitForm: () => Promise<void> }): JSX.Element {
    const { onClose } = props;
    function nextStep() {
      setCurrentStep(prevState => Math.min(EditSpendingDueDateStep.Frequency, prevState + 1));
    }

    function previousStep() {
      setCurrentStep(prevState => Math.min(EditSpendingDueDateStep.NextRecurrence, prevState - 1));
    }

    const CancelButton = () => (
      <Button color="secondary" onClick={ onClose }>
        Cancel
      </Button>
    );

    const PreviousButton = () => (
      <Button disabled={ isSubmitting } color="secondary" onClick={ previousStep }>
        Previous
      </Button>
    );

    const NextButton = () => (
      <Button color="primary" onClick={ nextStep }>
        Next
      </Button>
    );

    const SubmitButton = () => (
      <Button
        disabled={ isSubmitting }
        onClick={ submitForm }
        color="primary"
        type="submit"
      >
        Update
      </Button>
    );

    switch (currentStep) {
      case EditSpendingDueDateStep.NextRecurrence:
        return (
          <Fragment>
            <CancelButton />
            <NextButton />
          </Fragment>
        );
      case EditSpendingDueDateStep.Frequency:
        return (
          <Fragment>
            <PreviousButton />
            <SubmitButton />
          </Fragment>
        );
      default:
        return (
          <Fragment>
            <PreviousButton />
            <NextButton />
          </Fragment>
        );
    }
  }

  const initial: editSpendingDueDateForm = {
    dueDate: props.spending.nextRecurrence,
    recurrenceRule: new Recurrence(),
  };

  return (
    <Fragment>
      <Error />
      <Formik
        initialValues={ initial }
        validate={ validateInput }
        onSubmit={ submit }
      >
        { ({
          values,
          handleSubmit,
          setFieldValue,
          isSubmitting,
          submitForm,
        }) => (
          <form onSubmit={ handleSubmit }>
            <Dialog open={ props.isOpen } maxWidth="sm">
              <DialogTitle>
                Edit expense due date
              </DialogTitle>
              <DialogContent>
                <DialogContentText>
                  Edit when you need your recurring expense's money by. This will change how much is contributed to
                  the expense each time you get paid.
                </DialogContentText>
                <div>
                  <Stepper activeStep={ currentStep } orientation="vertical">
                    <Step key="When do you need it next?">
                      <StepLabel>When do you need it next?</StepLabel>
                      <StepContent>
                        <DatePicker
                          minDate={ moment().startOf('day').add(1, 'day') }
                          onChange={ value => setFieldValue('dueDate', value.startOf('day')) }
                          inputFormat="MM/DD/yyyy"
                          value={ values.dueDate }
                          renderInput={ params => <TextField fullWidth { ...params } /> }
                        />
                      </StepContent>
                    </Step>
                    <Step key="How frequently do you need it?">
                      <StepLabel>How frequently do you need it?</StepLabel>
                      <StepContent>
                        { (currentStep === EditSpendingDueDateStep.Frequency || values.dueDate) &&
                          <RecurrenceList
                            disabled={ isSubmitting }
                            date={ values.dueDate }
                            onChange={ value => setFieldValue('recurrenceRule', value) }
                          />
                        }
                      </StepContent>
                    </Step>
                  </Stepper>
                </div>
              </DialogContent>
              <DialogActions>
                <StepActions isSubmitting={ isSubmitting } submitForm={ submitForm } />
              </DialogActions>
            </Dialog>
          </form>
        ) }
      </Formik>
    </Fragment>
  );
}
