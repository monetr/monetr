import { Typography } from '@material-ui/core';
import Balance from 'data/Balance';
import React, { Component } from 'react';
import { connect } from 'react-redux';
import { getBalance } from 'shared/balances/selectors/getBalance';

interface WithConnectionPropTypes {
  balance: Balance;
}

export class BalanceNavDisplay extends Component<WithConnectionPropTypes, any> {

  render() {
    return (
      <div className="flex-1 flex justify-center gap-2">
        <Typography> <b>Available:</b> { this.props.balance.getAvailableString() }</Typography>
        <Typography> <b>Safe-To-Spend:</b> { this.props.balance.getSafeToSpendString() }</Typography>
        <Typography> <b>Expenses:</b> { this.props.balance.getExpensesString() }</Typography>
        <Typography> <b>Goals:</b> { this.props.balance.getGoalsString() }</Typography>
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
