import { Chip, Divider, ListItem, Paper, Popover, Typography } from '@mui/material';
import classnames from 'classnames';
import TransactionNameEditor from 'components/Transactions/TransactionNameEditor';
import Spending from 'models/Spending';
import Transaction from 'models/Transaction';
import React, { Component, Fragment } from 'react';
import { connect } from 'react-redux';
import { getSpendingById } from 'shared/spending/selectors/getSpendingById';
import { getTransactionById } from 'shared/transactions/selectors/getTransactionById';
import SelectButton from 'components/SelectyBoi/SelectButton';
import SpendingSelectionList from 'components/Spending/SpendingSelectionList';
import updateTransaction from 'shared/transactions/actions/updateTransaction';

import './styles/TransactionItem.scss';
import { AppState } from 'store';

interface PropTypes {
  transactionId: number;
}

interface WithConnectionPropTypes extends PropTypes {
  transaction: Transaction;
  spending?: Spending;
  updateTransaction: (transaction: Transaction) => Promise<void>;
}

interface State {
  spentFromAnchorEl: Element | null;
  spentFromWidth: number | null;
}

export class TransactionItem extends Component<WithConnectionPropTypes, State> {

  state = {
    spentFromAnchorEl: null,
    spentFromWidth: 0,
  };

  getSpentFromString() {
    const { spending, transaction, updateTransaction } = this.props;
    const { spentFromAnchorEl } = this.state;

    if (transaction.getIsAddition()) {
      return null;
    }

    const updateSpentFrom = (selection: Spending | null) => {
      const spendingId = selection ? selection.spendingId : null;

      if (spendingId === transaction.spendingId) {
        return Promise.resolve();
      }

      const updatedTransaction = new Transaction({
        ...transaction,
        spendingId: spendingId,
      });

      return updateTransaction(updatedTransaction)
        .catch(error => alert(error));
    };

    const openPopover = (event: { currentTarget: Element }) => {
      this.setState({
        spentFromAnchorEl: event.currentTarget,
        spentFromWidth: event.currentTarget.clientWidth,
      });
    };

    const closePopover = () => this.setState({
      spentFromAnchorEl: null,
      spentFromWidth: null,
    });

    return (
      <Fragment>
        <SelectButton
          open={ Boolean(spentFromAnchorEl) }
          onClick={ openPopover }
        >
          <span className="mr-1 opacity-50">
            Spent From
          </span>
          <span className={ classnames('overflow-ellipsis overflow-hidden flex-nowrap whitespace-nowrap', {
            'opacity-50': !spending,
          }) }>
            { spending ? spending.name : 'Safe-To-Spend' }
          </span>
        </SelectButton>
        <Popover
          id={ `transaction-spent-from-popover-${ transaction.transactionId }` }
          open={ Boolean(spentFromAnchorEl) }
          anchorEl={ spentFromAnchorEl }
          onClose={ closePopover }
          anchorOrigin={ {
            vertical: 'bottom',
            horizontal: 'left',
          } }
          transformOrigin={ {
            vertical: 'top',
            horizontal: 'left',
          } }
        >
          <Paper
            style={ { width: `${ this.state.spentFromWidth }px` } }
            className="p-0 overflow-auto min-w-96 max-h-96"
          >
            <SpendingSelectionList
              value={ transaction.spendingId }
              onChange={ updateSpentFrom }
            />
          </Paper>
        </Popover>
      </Fragment>
    )
  }

  render() {
    const { transaction } = this.props;

    return (
      <Fragment>
        <ListItem className="transactions-item h-12" role="transaction-row">
          <div className="flex flex-row w-full">
            <div
              className="flex-shrink w-2/5 pr-1 font-semibold transaction-item-name place-self-center"
            >
              <TransactionNameEditor transactionId={ transaction.transactionId }/>
            </div>

            <p
              className="flex-auto w-2/5 pr-1 overflow-hidden transaction-expense-name overflow-ellipsis flex-nowrap whitespace-nowrap"
            >
              { this.getSpentFromString() }
            </p>
            <div className="flex items-center flex-none w-1/5">
              { transaction.isPending && <Chip label="Pending" className="self-center align-middle"/> }
              <div className="flex justify-end w-full">
                <Typography className={ classnames('amount align-middle self-center place-self-center', {
                  'addition': transaction.getIsAddition(),
                }) }>
                  <b>{ transaction.getAmountString() }</b>
                </Typography>
              </div>
            </div>
          </div>
        </ListItem>
        <Divider/>
      </Fragment>
    )
  }
}

export default connect(
  (state: AppState, props: PropTypes) => {
    const transaction = getTransactionById(props.transactionId)(state);

    return {
      transaction,
      spending: getSpendingById(transaction.spendingId)(state),
    }
  },
  {
    updateTransaction,
  }
)(TransactionItem)
