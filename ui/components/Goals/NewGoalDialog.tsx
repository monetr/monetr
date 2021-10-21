import MomentUtils from '@date-io/moment';
import {
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
  TextField
} from '@material-ui/core';
import { Alert } from '@material-ui/lab';
import { KeyboardDatePicker, MuiPickersUtilsProvider } from '@material-ui/pickers';
import FundingScheduleSelectionList from 'components/FundingSchedules/FundingScheduleSelectionList';
import Spending, { SpendingType } from 'data/Spending';
import { Formik, FormikHelpers } from "formik";
import moment from 'moment';
import React, { Component, Fragment } from 'react';
import { connect } from 'react-redux';
import { getSelectedBankAccountId } from 'shared/bankAccounts/selectors/getSelectedBankAccountId';
import createSpending from 'shared/spending/actions/createSpending';

export interface PropTypes {
  onClose: { (): void };
  isOpen: boolean;
}

interface WithConnectionPropTypes extends PropTypes {
  bankAccountId: number;
  createSpending: { (spending: Spending): Promise<any> }
}

enum NewGoalStep {
  Name,
  Amount,
  Date,
  Funding,
}

interface State {
  step: NewGoalStep;
  error?: string;
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
  byDate: moment().add('1 day'),
  fundingScheduleId: 0,
}

export class NewGoalDialog extends Component<WithConnectionPropTypes, State> {

  state = {
    step: NewGoalStep.Name,
    error: null,
  };

  submit = (values: NewGoalForm, { setSubmitting }: FormikHelpers<NewGoalForm>) => {
    const { bankAccountId, createSpending, onClose } = this.props;

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
      .then(() => onClose())
      .catch(error => {
        setSubmitting(false);

        this.setState({
          error: error.response.data.error,
        });
      });
  };

  nextStep = () => {
    return this.setState(prevState => ({
      step: Math.min(NewGoalStep.Funding, prevState.step + 1),
    }));
  };

  previousStep = () => {
    return this.setState(prevState => ({
      step: Math.max(NewGoalStep.Name, prevState.step - 1),
    }));
  };

  renderActions = (isSubmitting: boolean, submitForm: { (): Promise<any> }) => {
    const { onClose } = this.props;
    const { step } = this.state;

    const cancelButton = (
      <Button color="secondary" onClick={ onClose }>
        Cancel
      </Button>
    );

    const previousButton = (
      <Button
        disabled={ isSubmitting }
        color="secondary"
        onClick={ this.previousStep }
      >
        Previous
      </Button>
    );

    const nextButton = (
      <Button
        color="primary"
        onClick={ this.nextStep }
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
      case NewGoalStep.Name:
        return (
          <Fragment>
            { cancelButton }
            { nextButton }
          </Fragment>
        );
      case NewGoalStep.Funding:
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
    const { isOpen } = this.props;
    const { step } = this.state;

    return (
      <Formik
        initialValues={ initialValues }
        validate={ () => {
        } }
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
              <Dialog open={ isOpen } maxWidth="sm">
                <DialogTitle>
                  Create a new goal
                </DialogTitle>
                <DialogContent>
                  { this.renderErrorMaybe() }
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
                          <KeyboardDatePicker
                            fullWidth
                            minDate={ moment().add('1 day') }
                            name="date"
                            margin="normal"
                            id="date-picker-dialog"
                            label="Date picker dialog"
                            format="MM/DD/yyyy"
                            value={ values.byDate }
                            onChange={ (value) => setFieldValue('byDate', value) }
                            KeyboardButtonProps={ {
                              'aria-label': 'change date',
                            } }
                          />
                        </StepContent>
                      </Step>
                      <Step key="How do you want to fund it?">
                        <StepLabel>How do you want to fund it?</StepLabel>
                        <StepContent>
                          <div className="mt-5"/>
                          <FundingScheduleSelectionList
                            disabled={ isSubmitting }
                            onChange={ (value) => setFieldValue('fundingScheduleId', value.fundingScheduleId) }
                          />
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
    );
  }
}

export default connect(
  (state) => ({
    bankAccountId: getSelectedBankAccountId(state),
  }),
  {
    createSpending,
  }
)(NewGoalDialog)
