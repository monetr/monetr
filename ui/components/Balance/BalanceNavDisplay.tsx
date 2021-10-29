import { Typography } from '@material-ui/core';
import Balance from 'models/Balance';
import React, { Component } from 'react';
import { connect } from 'react-redux';
import { getBalance } from 'shared/balances/selectors/getBalance';

interface WithConnectionPropTypes {
  balance: Balance;
}

export class BalanceNavDisplay extends Component<WithConnectionPropTypes, any> {

  render() {
    if (!this.props.balance) {
      return null;
    }

    return (
      <div className="flex-1 flex justify-center gap-2">
        <Typography>
          <b>Safe-To-Spend:</b> { this.props.balance.getSafeToSpendString() }
        </Typography>
        <Typography variant="body2">
          <b>Expenses:</b> { this.props.balance.getExpensesString() }
        </Typography>
        <Typography variant="body2">
          <b>Goals:</b> { this.props.balance.getGoalsString() }
        </Typography>
        <Typography variant="body2">
          <b>Available:</b> { this.props.balance.getAvailableString() }
        </Typography>
      </div>
    )
  }
}

export default connect(
  (state) => ({
    balance: getBalance(state),
  }),
  {}
)(BalanceNavDisplay);
