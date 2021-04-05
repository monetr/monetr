import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  Snackbar
} from '@material-ui/core';
import { Alert } from '@material-ui/lab';
import SpendingSelectionList from 'components/Spending/SpendingSelectionList';
import Spending from 'data/Spending';
import Transaction from 'data/Transaction';
import React, { Component } from 'react';
import { connect } from 'react-redux';
import updateTransaction from 'shared/transactions/actions/updateTransaction';

export interface PropTypes {
  isOpen: boolean;
  onClose: { (): void };
  transaction: Transaction;
}

interface WithConnectionPropTypes extends PropTypes {
  updateTransaction: { (transaction: Transaction): Promise<any> }
}

interface State {
  spendingId: number | null;
  error: string | null;
}

export class EditSpentFromDialog extends Component<WithConnectionPropTypes, State> {

  state = {
    error: null,
    spendingId: null,
  };

  componentDidMount() {
    this.setState({
      spendingId: this.props.transaction.spendingId,
    });
  }

  selectSpending = (spending: Spending | null) => {
    return this.setState({
      spendingId: spending === null ? null : spending.spendingId,
    });
  };

  save = () => {
    const { transaction, updateTransaction, onClose } = this.props;
    const { spendingId } = this.state;

    // If nothing has actually changed then we don't need to do anything, return a resolved promise.
    if (transaction.spendingId === spendingId) {
      return Promise.resolve();
    }

    transaction.spendingId = spendingId;

    return updateTransaction(transaction)
      .then(() => {
        return onClose();
      })
      .catch(error => {
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
    const { isOpen, onClose } = this.props;
    const { spendingId } = this.state;

    return (
      <Dialog open={ isOpen }>
        <DialogTitle>
          Choose where to spend from
        </DialogTitle>
        <DialogContent>
          { this.renderErrorMaybe() }
          <DialogContentText>
            Select an expense or a goal where you'd like to spend this transaction from. This will deduct the amount of
            the transaction from the expense or goal rather than from your Safe To Spend.
          </DialogContentText>
          <SpendingSelectionList
            value={ spendingId }
            onChange={ this.selectSpending }
          />
        </DialogContent>
        <DialogActions>
          <Button
            color="secondary"
            onClick={ onClose }
          >
            Cancel
          </Button>
          <Button
            onClick={ this.save }
            color="primary"
          >
            Save
          </Button>
        </DialogActions>
      </Dialog>
    );
  }
}


export default connect(
  state => ({}),
  {
    updateTransaction,
  }
)(EditSpentFromDialog);
