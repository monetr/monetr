import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle } from '@material-ui/core';
import SpendingSelectionList from 'components/Spending/SpendingSelectionList';
import Spending from 'data/Spending';
import Transaction from 'data/Transaction';
import React, { Component } from 'react';

export interface PropTypes {
  isOpen: boolean;
  onClose: { (): void };
  transaction: Transaction;
}

interface State {
  spendingId: number|null;
}

export class EditSpentFromDialog extends Component<PropTypes, State> {

  state = {
    spendingId: null,
  };

  selectSpending = (spending: Spending|null) => {
    return this.setState({
      spendingId: spending === null ? null : spending.spendingId,
    });
  };

  renderErrorMaybe = () => {
    return null;
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
            onClick={ () => {} }
            color="primary"
          >
            Save
          </Button>
        </DialogActions>
      </Dialog>
    );
  }
}

