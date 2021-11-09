import { Typography } from '@mui/material';
import React from 'react';
import { useSelector } from 'react-redux';
import { getBalance } from 'shared/balances/selectors/getBalance';

const BalanceNavDisplay = (): JSX.Element => {
  const balance = useSelector(getBalance);

  if (!balance) {
    return null;
  }

  return (
    <div className="flex-1 flex justify-center gap-2">
      <Typography>
        <b>Safe-To-Spend:</b> { balance.getSafeToSpendString() }
      </Typography>
      <Typography variant="body2">
        <b>Expenses:</b> { balance.getExpensesString() }
      </Typography>
      <Typography variant="body2">
        <b>Goals:</b> { balance.getGoalsString() }
      </Typography>
      <Typography variant="body2">
        <b>Available:</b> { balance.getAvailableString() }
      </Typography>
    </div>
  )
}

export default BalanceNavDisplay;