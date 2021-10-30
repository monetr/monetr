import { DatePicker } from '@mui/lab';
import React, { Component, Fragment } from 'react';
import { Formik, FormikErrors, FormikHelpers } from 'formik';
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
  Stepper, TextField
} from '@mui/material';
import { connect } from 'react-redux';
import { getSelectedExpense } from 'shared/spending/selectors/getSelectedExpense';
import Spending from 'models/Spending';
import updateSpending from 'shared/spending/actions/updateSpending';
import moment from 'moment';
import { RecurrenceList } from 'components/Recurrence/RecurrenceList';
import Recurrence from 'components/Recurrence/Recurrence';

enum EditSpendingDueDateStep {
  NextRecurrence,
  Frequency,
}

export interface PropTypes {
  onClose: { (): void };
  isOpen: boolean;
}

interface WithConnectionPropTypes extends PropTypes {
  spending: Spending;
  updateSpending: { (spending: Spending): Promise<any> }
}

interface State {
  error?: string;
  step: EditSpendingDueDateStep;
}

interface editSpendingDueDateForm {
  dueDate: moment.Moment;
  recurrenceRule: Recurrence;
}

export class EditExpenseDueDateDialog extends Component<WithConnectionPropTypes, State> {

  state = {
    error: null,
    step: EditSpendingDueDateStep.NextRecurrence,
  };

  validateInput = (values: editSpendingDueDateForm): FormikErrors<any> => {
    return null;
  };

  submit = (values: editSpendingDueDateForm, { setSubmitting }: FormikHelpers<editSpendingDueDateForm>) => {
    const { spending, updateSpending, onClose } = this.props;

    const updatedSpending = new Spending({
      ...spending,
      nextRecurrence: values.dueDate.startOf('day'),
      recurrenceRule: values.recurrenceRule.ruleString(),
      description: values.recurrenceRule.name,
    });

    return updateSpending(updatedSpending)
      .then(() => {
        return onClose();
      })
      .catch(error => {
        setSubmitting(false);

        this.setState({
          error: error.response.data.error,
        });
      });
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

  nextStep = () => {
    return this.setState(prevState => ({
      step: Math.min(EditSpendingDueDateStep.Frequency, prevState.step + 1),
    }));
  };

  previousStep = () => {
    return this.setState(prevState => ({
      step: Math.max(EditSpendingDueDateStep.NextRecurrence, prevState.step - 1),
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
        Update
      </Button>
    );

    switch (step) {
      case EditSpendingDueDateStep.NextRecurrence:
        return (
          <Fragment>
            { cancelButton }
            { nextButton }
          </Fragment>
        );
      case EditSpendingDueDateStep.Frequency:
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
    const { isOpen } = this.props;

    const initial: editSpendingDueDateForm = {
      dueDate: this.props.spending.nextRecurrence,
      recurrenceRule: new Recurrence(),
    };

    const { step } = this.state;

    return (
      <Fragment>
        { this.renderErrorMaybe() }
        <Formik
          initialValues={ initial }
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
                  Edit expense due date
                </DialogTitle>
                <DialogContent>
                  <DialogContentText>
                    Edit when you need your recurring expense's money by. This will change how much is contributed to
                    the expense each time you get paid.
                  </DialogContentText>
                  <div>
                    <Stepper activeStep={ step } orientation="vertical">
                      <Step key="When do you need it next?">
                        <StepLabel>When do you need it next?</StepLabel>
                        <StepContent>
                          <DatePicker
                            minDate={ moment().startOf('day').add(1, 'day') }
                            onChange={ (value) => setFieldValue('dueDate', value.startOf('day')) }
                            inputFormat="MM/DD/yyyy"
                            value={ values.dueDate }
                            renderInput={ (params) => <TextField fullWidth { ...params } /> }
                          />
                        </StepContent>
                      </Step>
                      <Step key="How frequently do you need it?">
                        <StepLabel>How frequently do you need it?</StepLabel>
                        <StepContent>
                          { (step === EditSpendingDueDateStep.Frequency || values.dueDate) &&
                          <RecurrenceList
                            disabled={ isSubmitting }
                            date={ values.dueDate }
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
            </form>
          ) }
        </Formik>
      </Fragment>
    );
  }
}

export default connect(
  state => ({
    spending: getSelectedExpense(state),
  }),
  {
    updateSpending,
  }
)(EditExpenseDueDateDialog);
