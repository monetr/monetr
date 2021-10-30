import React, { Component, Fragment } from "react";
import { Formik, FormikErrors, FormikHelpers } from "formik";
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
  Snackbar
} from "@mui/material";
import { connect } from "react-redux";
import { getSelectedExpense } from "shared/spending/selectors/getSelectedExpense";
import Spending from "models/Spending";
import updateSpending from "shared/spending/actions/updateSpending";


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
  initialValues: editSpendingAmountForm;
}

interface editSpendingAmountForm {
  amount: number;
}

export class EditExpenseAmountDialog extends Component<WithConnectionPropTypes, State> {

  state = {
    error: null,
    initialValues: {
      amount: 0.00,
    },
  };

  validateInput = (values: editSpendingAmountForm): FormikErrors<any> => {
    return null;
  };

  submit = (values: editSpendingAmountForm, { setSubmitting }: FormikHelpers<editSpendingAmountForm>) => {
    const { spending, updateSpending, onClose } = this.props;

    const updatedSpending = new Spending({
      ...spending,
      targetAmount: values.amount * 100,
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

  render() {
    const { isOpen } = this.props;

    const initial: editSpendingAmountForm = {
      amount: this.props.spending.getTargetAmountDollars(),
    };

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
              <Dialog open={ isOpen } maxWidth='xs'>
                <DialogTitle>
                  Edit amount
                </DialogTitle>
                <DialogContent>
                  <DialogContentText>
                    Edit the amount you want to put aside for this goal or expense. This will change how much is put
                    aside
                    each time you are paid as well to make sure you have enough by your target date.
                  </DialogContentText>
                  <FormControl fullWidth>
                    <InputLabel htmlFor="edit-expense-amount">Target Amount</InputLabel>
                    <Input
                      autoFocus={ true }
                      id="edit-expense-amount"
                      name="amount"
                      value={ values.amount }
                      onBlur={ handleBlur }
                      onChange={ handleChange }
                      disabled={ isSubmitting }
                      startAdornment={ <InputAdornment position="start">$</InputAdornment> }
                    />
                  </FormControl>
                </DialogContent>
                <DialogActions>
                  <Button
                    disabled={ isSubmitting }
                    onClick={ this.props.onClose }
                  >
                    Cancel
                  </Button>
                  <Button
                    disabled={ isSubmitting }
                    color="primary"
                    onClick={ submitForm }
                  >
                    Save
                  </Button>
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
    // TODO Edit amount dialog doesn't work with goals. But it should
    spending: getSelectedExpense(state),
  }),
  {
    updateSpending,
  }
)(EditExpenseAmountDialog);
