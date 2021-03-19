import { Box, Grid, Typography } from "@material-ui/core";
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

  render() {
    const { transaction } = this.props;

    return (
      <Box className="transactions-item">
        <Grid container spacing={ 2 }>
          <Grid item xs={ 12 } sm container>
            <Grid item xs container spacing={ 2 } direction="column">
              <Grid item xs>
                <Typography>
                  { transaction.name }
                </Typography>
              </Grid>
              <Grid item>
                <Typography>
                  { this.getSpentFromString() }
                </Typography>
              </Grid>
            </Grid>
            <Grid item>
              { transaction.isPending &&
              <Typography>Pending</Typography>
              }
              <Typography align="right" className={ classnames('amount', {
                'addition': transaction.getIsAddition(),
              }) }>
                { transaction.getAmountString() }
              </Typography>
              <Typography>
                { transaction.categories.join(', ') }
              </Typography>
            </Grid>
          </Grid>
        </Grid>
      </Box>
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
