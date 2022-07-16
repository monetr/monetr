import React, {  useState } from 'react';
import { DatePicker } from '@mui/lab';
import {
  Alert,
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
import StepperDialogActionButtons, { StepperStep } from 'components/StepperDialogActionButtons';
import { useSelectedBankAccountId } from 'hooks/bankAccounts';
import { useCreateFundingSchedule } from 'hooks/fundingSchedules';
import FundingSchedule from 'models/FundingSchedule';

interface newFundingScheduleForm {
  name: string;
  nextOccurrence: moment.Moment;
  recurrenceRule: Recurrence;
}

const initialValues: newFundingScheduleForm = {
  name: '',
  nextOccurrence: moment().add(1, 'day').startOf('day'),
  recurrenceRule: new Recurrence(),
};

interface Props {
  onClose: () => void;
  isOpen: boolean;
}

export default function NewFundingScheduleDialog(props: Props): JSX.Element {
  enum NewFundingScheduleStep {
    Name,
    Date,
    Recurrence,
  }
  const [currentStep, setCurrentStep] = useState<NewFundingScheduleStep>(NewFundingScheduleStep.Name);
  const [error, setError] = useState<string|null>(null);
  const bankAccountId = useSelectedBankAccountId();
  const createFundingSchedule = useCreateFundingSchedule();

  function validateInput(_: newFundingScheduleForm): FormikErrors<newFundingScheduleForm> {
    return null;
  }

  async function submit(
    values: newFundingScheduleForm,
    { setSubmitting }: FormikHelpers<newFundingScheduleForm>,
  ): Promise<void> {
    setSubmitting(false);
    const newFundingSchedule = new FundingSchedule({
      bankAccountId: bankAccountId,
      name: values.name,
      description: values.recurrenceRule.name,
      nextOccurrence: values.nextOccurrence.startOf('day'),
      rule: values.recurrenceRule.ruleString(),
    });

    return createFundingSchedule(newFundingSchedule)
      .then(() => props.onClose())
      .catch(error => setError(error.response.data.error))
      .finally(() => setSubmitting(false));
  }
  const nextStep = () => setCurrentStep(Math.min(NewFundingScheduleStep.Recurrence, currentStep + 1));
  const previousStep = () => setCurrentStep(Math.max(NewFundingScheduleStep.Name, currentStep - 1));

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
          <Dialog open={ props.isOpen } maxWidth="sm" className="new-funding-schedule">
            <DialogTitle>
              Create a new funding schedule
            </DialogTitle>
            <DialogContent>
              <DialogContentText>
                Funding schedules let us know when you will get paid so we can automatically allocate money towards
                your budgets.
              </DialogContentText>
              <Error />
              <div>
                <Stepper activeStep={ currentStep } orientation="vertical">
                  <Step key="What do you want to call this funding schedule?">
                    <StepLabel>What do you want to call this funding schedule?</StepLabel>
                    <StepContent>
                      <TextField
                        autoFocus
                        className="w-full"
                        disabled={ isSubmitting }
                        error={ touched.name && !!errors.name }
                        helperText={ (touched.name && errors.name) ? errors.name : null }
                        id="new-funding-schedule-name"
                        label="Name"
                        name="name"
                        onBlur={ handleBlur }
                        onChange={ handleChange }
                        value={ values.name }
                      />
                    </StepContent>
                  </Step>
                  <Step key="When do you get paid next?">
                    <StepLabel>When do you get paid next?</StepLabel>
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
                  <Step key="How often do you get paid?">
                    <StepLabel>How often do you get paid?</StepLabel>
                    <StepContent>
                      { (currentStep === NewFundingScheduleStep.Recurrence || values.nextOccurrence) &&
                        <RecurrenceList
                          disabled={ isSubmitting }
                          date={ values.nextOccurrence }
                          onChange={ value => setFieldValue('recurrenceRule', value) }
                        />
                      }
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
                canNextStep={ true }
                step={ currentStep === NewFundingScheduleStep.Name ? StepperStep.First :
                  currentStep === NewFundingScheduleStep.Recurrence ? StepperStep.Last : StepperStep.Other }
              />
            </DialogActions>
          </Dialog>
        </form>
      ) }
    </Formik>
  );
}
