import MomentUtils from "@date-io/moment";
import {
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
  TextField
} from "@material-ui/core";
import { Alert } from "@material-ui/lab";
import { KeyboardDatePicker, MuiPickersUtilsProvider } from '@material-ui/pickers';
import Recurrence from "components/Recurrence/Recurrence";
import { RecurrenceList } from "components/Recurrence/RecurrenceList";
import FundingSchedule from "data/FundingSchedule";
import { Formik, FormikErrors } from "formik";
import moment from "moment";
import React, { Component, Fragment } from "react";
import { connect } from "react-redux";
import { getSelectedBankAccountId } from "shared/bankAccounts/selectors/getSelectedBankAccountId";
import createFundingSchedule from "shared/fundingSchedules/actions/createFundingSchedule";
import { AppState } from 'store';

enum NewFundingScheduleStep {
  Name,
  Date,
  Recurrence,
}

export interface PropTypes {
  onClose: { (): void };
  isOpen: boolean;
}

interface WithConnectionPropTypes extends PropTypes {
  bankAccountId: number;
  createFundingSchedule: { (fundingSchedule: FundingSchedule): Promise<FundingSchedule> }
}

interface ComponentState {
  step: NewFundingScheduleStep;
  error?: string;
}

interface newFundingScheduleForm {
  name: string;
  nextOccurrence: moment.Moment;
  recurrenceRule: Recurrence;
}

const initialValues: newFundingScheduleForm = {
  name: '',
  nextOccurrence: moment(),
  recurrenceRule: new Recurrence(),
};

export class NewFundingScheduleDialog extends Component<WithConnectionPropTypes, ComponentState> {

  state = {
    step: NewFundingScheduleStep.Name,
    error: null,
  };

  validateInput = (values: newFundingScheduleForm): FormikErrors<any> => {
    return {};
  };

  submit = (values: newFundingScheduleForm, { setSubmitting }) => {
    const { bankAccountId, createFundingSchedule } = this.props;

    const newFundingSchedule = new FundingSchedule({
      bankAccountId: bankAccountId,
      name: values.name,
      description: values.recurrenceRule.name,
      nextOccurrence: values.nextOccurrence.startOf('day'),
      rule: values.recurrenceRule.ruleString(),
    });

    return createFundingSchedule(newFundingSchedule)
      .then(result => {
        // Close the dialog.
        this.props.onClose();
      }).catch(error => {
        setSubmitting(false);

        this.setState({
          error: error.response.data.error,
        });
      });
  };


  nextStep = () => {
    return this.setState(prevState => ({
      step: Math.min(NewFundingScheduleStep.Recurrence, prevState.step + 1),
    }));
  };

  previousStep = () => {
    return this.setState(prevState => ({
      step: Math.max(NewFundingScheduleStep.Name, prevState.step - 1),
    }));
  };

  renderActions = (isSubmitting: boolean, submitForm: { (): Promise<any> }) => {
    const { onClose } = this.props;
    const { step } = this.state;

    const cancelButton = (
      <Button color="secondary" onClick={ onClose } disabled={ isSubmitting }>
        Cancel
      </Button>
    );

    const previousButton = (
      <Button color="secondary" onClick={ this.previousStep } disabled={ isSubmitting }>
        Previous
      </Button>
    );

    const nextButton = (
      <Button
        data-testid="new-funding-schedule-next-button"
        color="primary"
        onClick={ this.nextStep }
        disabled={ isSubmitting }
      >
        Next
      </Button>
    );

    const submitButton = (
      <Button color="primary" type="submit" disabled={ isSubmitting } onClick={ submitForm }>
        Create
      </Button>
    );

    switch (step) {
      case NewFundingScheduleStep.Name:
        return (
          <Fragment>
            { cancelButton }
            { nextButton }
          </Fragment>
        );
      case NewFundingScheduleStep.Recurrence:
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
  };

  renderErrorMaybe = () => {
    const { error } = this.state;

    if (!error) {
      return null;
    }

    const onClose = () => this.setState({ error: null });

    return (
      <Snackbar open autoHideDuration={ 6000 } onClose={ onClose }>
        <Alert onClose={ onClose } severity="error">
          { error }
        </Alert>
      </Snackbar>
    )
  };

  render() {
    const { onClose, isOpen } = this.props;
    const { step } = this.state;

    return (
      <Formik
        initialValues={ initialValues }
        validate={ this.validateInput }
        onSubmit={ this.submit }
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
            <MuiPickersUtilsProvider utils={ MomentUtils }>
              <Dialog open={ isOpen } maxWidth="sm" className="new-funding-schedule">
                <DialogTitle>
                  Create a new funding schedule
                </DialogTitle>
                <DialogContent>
                  <DialogContentText>
                    Funding schedules let us know when you will get paid so we can automatically allocate money towards
                    your budgets.
                  </DialogContentText>
                  { this.renderErrorMaybe() }
                  <div>
                    <Stepper activeStep={ step } orientation="vertical">
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
                          <KeyboardDatePicker
                            data-testid="new-funding-schedule-date-picker"
                            fullWidth
                            minDate={ moment().subtract('1 day') }
                            name="date"
                            margin="normal"
                            id="date-picker-dialog"
                            label="Date"
                            format="MM/DD/yyyy"
                            value={ values.nextOccurrence }
                            onChange={ (value) => setFieldValue('nextOccurrence', value.startOf('day')) }
                            KeyboardButtonProps={ {
                              'aria-label': 'change date',
                            } }
                          />
                        </StepContent>
                      </Step>
                      <Step key="How often do you get paid?">
                        <StepLabel>How often do you get paid?</StepLabel>
                        <StepContent>
                          { (step === NewFundingScheduleStep.Recurrence || values.nextOccurrence) &&
                          <RecurrenceList
                            disabled={ isSubmitting }
                            date={ values.nextOccurrence }
                            onChange={ (value) => setFieldValue('recurrenceRule', value) }
                          />
                          }
                        </StepContent>
                      </Step>
                    </Stepper>
                  </div>
                </DialogContent>
                <DialogActions>
                  { this.renderActions(isSubmitting, submitForm) }
                </DialogActions>
              </Dialog>
            </MuiPickersUtilsProvider>
          </form>
        ) }
      </Formik>
    )
  }
}

export default connect(
  (state: AppState) => ({
    bankAccountId: getSelectedBankAccountId(state),
  }),
  {
    createFundingSchedule,
  }
)(NewFundingScheduleDialog);
