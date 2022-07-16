import React, { Component, Fragment } from 'react';
import { connect } from 'react-redux';
import { Alert, Dialog, DialogContent, DialogContentText, DialogTitle, Snackbar } from '@mui/material';

import FundingScheduleSelectionList from 'components/FundingSchedules/FundingScheduleSelectionList';
import { Formik, FormikErrors } from 'formik';
import Spending from 'models/Spending';
import { getSelectedExpense } from 'shared/spending/selectors/getSelectedExpense';

export interface Props {
  onClose: () => void;
  isOpen: boolean;
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

  validateInput = (_: {}): FormikErrors<any> => {
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
    );
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
            handleSubmit,
            setFieldValue,
            isSubmitting,
          }) => (
            <form onSubmit={ handleSubmit }>
              <Dialog open={ true } maxWidth='xs'>
                <DialogTitle>
                  Edit funding schedule
                </DialogTitle>
                <DialogContent>
                  <DialogContentText>
                    Change how frequently you want to contribute to this expense. This might also change how much is
                    contributed to the expense depending on the new frequency.
                  </DialogContentText>
                  <FundingScheduleSelectionList
                    disabled={ isSubmitting }
                    onChange={ value => setFieldValue('fundingScheduleId', value.fundingScheduleId) }
                  />
                </DialogContent>
              </Dialog>
            </form>
          ) }
        </Formik>
      </Fragment>
    );
  }
}

export default connect(
  (state, _: Props) => ({
    expense: getSelectedExpense(state),
  }),
  {}
)(EditExpenseFundingScheduleDialog);
