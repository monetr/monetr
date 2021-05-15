import React, { Component, Fragment } from "react";
import { connect } from "react-redux";
import { getSpendingById } from "shared/spending/selectors/getSpendingById";
import Spending from "data/Spending";
import { Dialog, Snackbar } from "@material-ui/core";
import { Alert } from "@material-ui/lab";
import { Formik, FormikErrors } from "formik";

export interface Props {
  expenseId: number;
}

interface WithConnectionPropTypes extends Props {
  expense: Spending;
}

interface State {
  error?: string;
}

export class EditExpenseFundingScheduleDialog extends Component<WithConnectionPropTypes, State> {

  state = {
    error: null,
  };

  submit = () => {

  };

  validateInput = (values: {}): FormikErrors<any> => {
    return null;
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
    return (
      <Fragment>
        { this.renderErrorMaybe() }
        <Formik
          initialValues={ {} }
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
              <Dialog open={ true } maxWidth='xs'>

              </Dialog>
            </form>
          ) }
        </Formik>
      </Fragment>
    )
  }
}

export default connect(
  (state, props: Props) => ({
    expense: getSpendingById(props.expenseId)(state),
  }),
  {}
)(EditExpenseFundingScheduleDialog);
