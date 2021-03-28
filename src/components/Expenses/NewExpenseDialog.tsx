import MomentUtils from '@date-io/moment';
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControl,
  Input,
  InputAdornment,
  InputLabel,
  Step,
  StepContent,
  StepLabel,
  Stepper,
  TextField
} from "@material-ui/core";
import { KeyboardDatePicker, MuiPickersUtilsProvider } from '@material-ui/pickers';
import FundingSchedule from "data/FundingSchedule";
import { Formik, FormikErrors } from "formik";
import { Map } from 'immutable';
import moment from "moment";
import React, { Component, Fragment } from "react";
import { connect } from "react-redux";

enum NewExpenseStep {
  Name,
  Amount,
  Date,
  Recurrence,
  Funding,
}

export interface PropTypes {
  onClose: { (): void };
  isOpen: boolean;
}

export interface WithConnectionPropTypes extends PropTypes {
  fundingSchedules: Map<number, FundingSchedule>;
}

export interface State {
  step: NewExpenseStep;
  canNextStep: boolean;
}

interface newExpenseForm {
  name: string;
  amount: number;
  nextOccurrence: moment.Moment;
  recurrenceRule: string;
  fundingScheduleId: number;
}

const initialValues: newExpenseForm = {
  name: '',
  amount: 0.00,
  nextOccurrence: moment(),
  recurrenceRule: '',
  fundingScheduleId: 0,
};

class NewExpenseDialog extends Component<WithConnectionPropTypes, State> {

  state = {
    step: NewExpenseStep.Name,
    canNextStep: true,
  };

  validateInput = (values: newExpenseForm): FormikErrors<any> => {
    const { step, canNextStep } = this.state;

    switch (step) {
      // case NewExpenseStep.Name:
      //   if (values.name.length === 0) {
      //     return {
      //       name: 'Name cannot be blank'
      //     };
      //   }
    }

    // if (!canNextStep) {
    //   this.setState({
    //     canNextStep: true
    //   });
    // }

    return {};
  };

  submit = (values: newExpenseForm, { setSubmitting }) => {

  };

  nextStep = () => {
    return this.setState(prevState => ({
      canNextStep: true,
      step: Math.min(NewExpenseStep.Funding, prevState.step + 1),
    }));
  };

  previousStep = () => {
    return this.setState(prevState => ({
      canNextStep: true, // Math.min(NewExpenseStep.Name, prevState.step - 1) < prevState.step,
      step: Math.max(NewExpenseStep.Name, prevState.step - 1),
    }));
  };

  renderActions = () => {
    const { onClose } = this.props;
    const { step, canNextStep } = this.state;

    const cancelButton = (
      <Button color="secondary" onClick={ onClose }>
        Cancel
      </Button>
    );

    const previousButton = (
      <Button color="secondary" onClick={ this.previousStep }>
        Previous
      </Button>
    );

    const nextButton = (
      <Button color="primary" onClick={ this.nextStep } disabled={ !canNextStep }>
        Next
      </Button>
    );

    const submitButton = (
      <Button color="primary" type="submit">
        Create
      </Button>
    );

    switch (step) {
      case NewExpenseStep.Name:
        return (
          <Fragment>
            { cancelButton }
            { nextButton }
          </Fragment>
        );
      case NewExpenseStep.Funding:
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
           }) => (
          <form onSubmit={ handleSubmit }>
            <MuiPickersUtilsProvider utils={ MomentUtils }>
              <Dialog open={ isOpen }>
                <DialogTitle>
                  Create a new expense
                </DialogTitle>
                <DialogContent>
                  <div className="w-96">
                    <Stepper activeStep={ step } orientation="vertical">
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
                          <KeyboardDatePicker
                            fullWidth
                            minDate={ moment().subtract('1 day') }
                            name="date"
                            margin="normal"
                            id="date-picker-dialog"
                            label="Date picker dialog"
                            format="MM/DD/yyyy"
                            value={ values.nextOccurrence }
                            onChange={ (value) => setFieldValue('nextOccurrence', value) }
                            KeyboardButtonProps={ {
                              'aria-label': 'change date',
                            } }
                          />
                        </StepContent>
                      </Step>
                      <Step key="How frequently do you need it?">
                        <StepLabel>How frequently do you need it?</StepLabel>
                        <StepContent>
                          <TextField id="new-expense-frequency" className="w-full" label="Frequency"/>
                        </StepContent>
                      </Step>
                      <Step key="How do you want to fund it?">
                        <StepLabel>How do you want to fund it?</StepLabel>
                        <StepContent>
                          <TextField id="new-expense-funding" className="w-full" label="Funding"/>
                        </StepContent>
                      </Step>
                    </Stepper>
                  </div>
                </DialogContent>
                <DialogActions>
                  { this.renderActions() }
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
  (state, props: PropTypes) => ({
    fundingSchedules: Map<number, FundingSchedule>(),
  }),
  {}
)(NewExpenseDialog)
