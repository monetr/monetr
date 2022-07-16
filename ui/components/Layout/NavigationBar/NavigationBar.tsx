import React from 'react';
import { Menu } from '@mui/icons-material';
import { AppBar, IconButton, Toolbar } from '@mui/material';

import BalanceNavDisplay from 'components/Balance/BalanceNavDisplay';

import 'components/Layout/NavigationBar/styles/NavigationBar.scss';

interface NavigationBarProps {
  onToggleSidebar?: () => void;
}

export default function NavigationBar(props: NavigationBarProps): JSX.Element {
  return (
    <AppBar position="static">
      <Toolbar>
        <IconButton onClick={ props.onToggleSidebar } aria-label="menu" className="text-white block lg:hidden">
          <Menu />
        </IconButton>
        <BalanceNavDisplay />
      </Toolbar>
    </AppBar>
  );
}
