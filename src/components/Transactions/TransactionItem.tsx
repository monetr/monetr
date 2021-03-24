import { Checkbox, ListItem, ListItemIcon, Typography } from "@material-ui/core";
import classnames from 'classnames';
import Expense from "data/Expense";
import Transaction from "data/Transaction";
import React, { Component } from 'react';
import { connect } from "react-redux";
import { getExpenseById } from "shared/expenses/selectors/getExpenseById";
import { getTransactionById } from "shared/transactions/selectors/getTransactionById";

import './styles/TransactionItem.scss';

interface PropTypes {
  transactionId: number;
  selected: boolean;
  onClick: { (transactionId: number): void };
}

interface WithConnectionPropTypes extends PropTypes {
  transaction: Transaction;
  expense?: Expense;
}

class TransactionItem extends Component<WithConnectionPropTypes, {}> {

  getSpentFromString(): string {
    const { expense } = this.props;

    if (!expense) {
      return 'Spent From Safe-To-Spend';
    }

    return `Spent From ${ expense.name }`;
  }

  handleClick = () => {
    return this.props.onClick(this.props.transactionId);
  }

  render() {
    const { transaction, selected } = this.props;

    return (
      <ListItem button onClick={ this.handleClick } className="transactions-item">
        <ListItemIcon>
          <Checkbox
            edge="start"
            checked={ selected }
            tabIndex={ -1 }
          />
        </ListItemIcon>
        <div className="grid grid-cols-4 grid-rows-2 grid-flow-col gap-1 w-full">
          <div className="col-span-3">
            <Typography>{ transaction.name }</Typography>
          </div>
          <div className="col-span-3 opacity-75">
            <Typography>{ this.getSpentFromString() }</Typography>
          </div>
          <div className="row-span-2 col-span-1 flex justify-end">
            <Typography className={ classnames('amount align-middle self-center', {
              'addition': transaction.getIsAddition(),
            }) }>
              { transaction.getAmountString() }
            </Typography>
          </div>
        </div>
      </ListItem>
    )
  }
}

export default connect(
  (state, props: PropTypes) => {
    const transaction = getTransactionById(props.transactionId)(state);

    return {
      transaction,
      expense: getExpenseById(transaction.expenseId)(state),
    }
  },
  {}
)(TransactionItem)
