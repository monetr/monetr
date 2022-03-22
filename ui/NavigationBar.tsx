import { AppBar, Toolbar } from '@mui/material';
import BalanceNavDisplay from 'components/Balance/BalanceNavDisplay';
import React from 'react';

const NavigationBar = React.memo((): JSX.Element => {
  return (
    <AppBar position="static">
      <Toolbar>
        <BalanceNavDisplay/>
      </Toolbar>
    </AppBar>
  )
});

export default NavigationBar;
