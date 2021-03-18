import React, { Component } from 'react';
import { connect } from "react-redux";
import Transaction from "data/Transaction";
import { Avatar, Box, Grid, Typography } from "@material-ui/core";
import { getTransactionById } from "shared/transactions/selectors/getTransactionById";
import classnames from 'classnames';
import Expense from "data/Expense";

import './styles/TransactionItem.scss';
import { getExpenseById } from "shared/expenses/selectors/getExpenseById";

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

    return `Spent From ${expense.name}`;
  }

  render() {
    const { transaction } = this.props;

    return (
      <Box className="transactions-item">
        <Grid container spacing={ 2 }>
          <Grid item>
            <Box bgcolor="primary.main" clone>
              <Avatar>

              </Avatar>
            </Box>
          </Grid>
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
              <Typography className={ classnames('amount', {
                'addition': transaction.getIsAddition(),
              }) }>
                { transaction.getAmountString() }
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
      transaction: transaction,
      expense: getExpenseById(transaction.expenseId)(state),
    }
  },
  {}
)(TransactionItem)
