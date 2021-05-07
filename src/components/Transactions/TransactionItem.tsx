import { Checkbox, Chip, Divider, ListItem, ListItemIcon, Typography } from "@material-ui/core";
import classnames from 'classnames';
import Spending from "data/Spending";
import Transaction from "data/Transaction";
import React, { Component, Fragment } from 'react';
import { connect } from "react-redux";
import { getSpendingById } from "shared/spending/selectors/getSpendingById";
import selectTransaction from "shared/transactions/actions/selectTransaction";
import { getTransactionById } from "shared/transactions/selectors/getTransactionById";

import './styles/TransactionItem.scss';
import { getTransactionIsSelected } from "shared/transactions/selectors/getTransactionIsSelected";

interface PropTypes {
  transactionId: number;
}

interface WithConnectionPropTypes extends PropTypes {
  transaction: Transaction;
  spending?: Spending;
  isSelected: boolean;
  selectTransaction: { (transactionId: number): void }
}

export class TransactionItem extends Component<WithConnectionPropTypes, {}> {

  getSpentFromString() {
    const { spending, transaction } = this.props;

    if (transaction.getIsAddition()) {
      return (
        <Fragment>
          Deposited Into Safe-To-Spend
        </Fragment>
      );
    }

    if (!spending) {
      return (
        <Fragment>
          Spent From Safe-To-Spend
        </Fragment>
      );
    }

    return (
      <Fragment>
        Spent From <b>{ spending.name }</b>
      </Fragment>
    )
  }

  handleClick = () => {
    return this.props.selectTransaction(this.props.transactionId);
  }

  render() {
    const { transaction, isSelected } = this.props;

    return (
      <Fragment>
        <ListItem button onClick={ this.handleClick } className="transactions-item" role="transaction-row">
          <ListItemIcon>
            <Checkbox
              edge="start"
              checked={ isSelected }
              tabIndex={ -1 }
              color="primary"
            />
          </ListItemIcon>
          <div className="grid grid-cols-8 grid-rows-2 grid-flow-col gap-1 w-full">
            <div className="col-span-6">
              <Typography className="transaction-item-name"><b>{ transaction.getName() }</b></Typography>
            </div>
            <div className="col-span-1">
              <Typography className="opacity-80">
                { transaction.date.format('MMMM Do') }
              </Typography>
            </div>
            <div className="col-span-5 opacity-75">
              <Typography className="transaction-expense-name">
                { this.getSpentFromString() }
              </Typography>
            </div>
            <div className="row-span-2 col-span-1 flex justify-end">
              { transaction.isPending && <Chip label="Pending" className="align-middle self-center"/> }
            </div>
            <div className="row-span-2 col-span-1 flex justify-end">
              <Typography className={ classnames('amount align-middle self-center', {
                'addition': transaction.getIsAddition(),
              }) }>
                <b>{ transaction.getAmountString() }</b>
              </Typography>
            </div>
          </div>
        </ListItem>
        <Divider />
      </Fragment>
    )
  }
}

export default connect(
  (state, props: PropTypes) => {
    const transaction = getTransactionById(props.transactionId)(state);
    const isSelected = getTransactionIsSelected(props.transactionId)(state);

    return {
      transaction,
      isSelected,
      spending: getSpendingById(transaction.spendingId)(state),
    }
  },
  {
    selectTransaction,
  }
)(TransactionItem)
