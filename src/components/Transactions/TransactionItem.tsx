import React, { Component } from 'react';
import { connect } from "react-redux";
import Transaction from "data/Transaction";
import { Avatar, Box, Grid, Typography } from "@material-ui/core";
import { getTransactionById } from "shared/transactions/selectors/getTransactionById";
import classnames from 'classnames';

import './styles/TransactionItem.scss';

interface PropTypes {
  transactionId: number;
  transaction: Transaction;
}

export class TransactionItem extends Component<PropTypes, {}> {

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
                  Spent From Safe-To-Spend
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
  (state, props: PropTypes) => ({
    transaction: getTransactionById(props.transactionId)(state),
  }),
  {}
)(TransactionItem)
