import React from 'react';
import { useSelector } from 'react-redux';
import { Typography } from '@mui/material';

import { getBalance } from 'shared/balances/selectors/getBalance';

const BalanceNavDisplay = React.memo((): JSX.Element => {
  const balance = useSelector(getBalance);

  if (!balance) {
    return null;
  }

  return (
    <div className="flex-1 flex justify-center gap-2 items-center">
      <Typography data-testid="safe-to-spend">
        <b>Safe-To-Spend:</b> { balance.getSafeToSpendString() }
      </Typography>
      <Typography variant="body2" className="hidden lg:block">
        <b>Expenses:</b> { balance.getExpensesString() }
      </Typography>
      <Typography variant="body2" className="hidden lg:block">
        <b>Goals:</b> { balance.getGoalsString() }
      </Typography>
      <Typography variant="body2" className="hidden lg:block">
        <b>Available:</b> { balance.getAvailableString() }
      </Typography>
    </div>
  );
});

export default BalanceNavDisplay;
