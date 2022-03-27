import { AppBar, Toolbar } from '@mui/material';
import BalanceNavDisplay from 'components/Balance/BalanceNavDisplay';
import React from 'react';

import 'components/Layout/NavigationBar/styles/NavigationBar.scss';

export default function NavigationBar(): JSX.Element {
  return (
    <AppBar position="static">
      <Toolbar>
        <BalanceNavDisplay/>
      </Toolbar>
    </AppBar>
  )
}
