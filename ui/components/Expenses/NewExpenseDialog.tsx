import MomentUtils from '@date-io/moment';
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
  TextField
} from '@mui/material';
import FundingScheduleSelectionList from 'components/FundingSchedules/FundingScheduleSelectionList';
import Recurrence from 'components/Recurrence/Recurrence';
import { RecurrenceList } from 'components/Recurrence/RecurrenceList';
import Spending, { SpendingType } from 'models/Spending';
import { Formik, FormikErrors } from 'formik';
import moment from 'moment';
import React, { Component, Fragment } from 'react';
import { connect } from 'react-redux';
import { getSelectedBankAccountId } from 'shared/bankAccounts/selectors/getSelectedBankAccountId';
import createSpending from 'shared/spending/actions/createSpending';
import { AppState } from 'store';

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
  bankAccountId: number;
  createSpending: { (spending: Spending): Promise<any> }
}

interface ComponentState {
  step: NewExpenseStep;
  canNextStep: boolean;
  error?: string;
}

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

class NewExpenseDialog extends Component<WithConnectionPropTypes, ComponentState> {

  state = {
    step: NewExpenseStep.Name,
    canNextStep: true,
    error: null,
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
    const { bankAccountId, createSpending, onClose } = this.props;

    const newSpending = new Spending({
      bankAccountId: bankAccountId,
      name: values.name,
      description: values.recurrenceRule.name,
      nextRecurrence: values.nextOccurrence.startOf('day'),
      spendingType: SpendingType.Expense,
      fundingScheduleId: values.fundingScheduleId,
      targetAmount: Math.ceil(values.amount * 100), // Convert to an integer.
      recurrenceRule: values.recurrenceRule.ruleString(),
    });

    return createSpending(newSpending)
      .then(() => {
        onClose();
      })
      .catch(error => {
        setSubmitting(false);

        this.setState({
          error: error.response.data.error,
        });
      })
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

  renderActions = (isSubmitting: boolean, submitForm: { (): Promise<any> }) => {
    const { onClose } = this.props;
    const { step, canNextStep } = this.state;

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
        disabled={ !canNextStep }
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
            <Dialog open={ isOpen } maxWidth="sm">
              <DialogTitle>
                Create a new expense
              </DialogTitle>
              <DialogContent>
                { this.renderErrorMaybe() }
                <DialogContentText>
                  Expenses let you budget for things that happen on a regular basis automatically. Money is allocated
                  to expenses whenever you get paid so that you don't have to pay something from a single paycheck.
                </DialogContentText>
                <div>
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
                        <DatePicker
                          minDate={ moment().startOf('day').add(1, 'day') }
                          onChange={ (value) => setFieldValue('nextOccurrence', value.startOf('day')) }
                          inputFormat="MM/DD/yyyy"
                          value={ values.nextOccurrence }
                          renderInput={ (params) => <TextField fullWidth { ...params } /> }
                        />
                      </StepContent>
                    </Step>
                    <Step key="How frequently do you need it?">
                      <StepLabel>How frequently do you need it?</StepLabel>
                      <StepContent>
                        { (step === NewExpenseStep.Recurrence || values.nextOccurrence) &&
                        <RecurrenceList
                          disabled={ isSubmitting }
                          date={ values.nextOccurrence }
                          onChange={ (value) => setFieldValue('recurrenceRule', value) }
                        />
                        }
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
          </form>
        ) }
      </Formik>
    )
  }
}

export default connect(
  (state: AppState, props: PropTypes) => ({
    bankAccountId: getSelectedBankAccountId(state),
  }),
  {
    createSpending,
  }
)(NewExpenseDialog)
