import { CreditCard, ExitToApp, LightMode } from '@mui/icons-material';
import MenuIcon from '@mui/icons-material/Menu';
import { AppBar, Button, IconButton, Menu, MenuItem, Toolbar } from '@mui/material';
import BalanceNavDisplay from 'components/Balance/BalanceNavDisplay';
import BankAccountSelector from 'components/BankAccounts/BankAccountSelector';
import React, { useState } from 'react';
import { useSelector } from 'react-redux';
import { Link as RouterLink } from 'react-router-dom';
import useLogout from 'shared/authentication/actions/logout';
import manageBilling from 'shared/billing/actions/manageBilling';
import { getBillingEnabled } from 'shared/bootstrap/selectors';

const NavigationBar = React.memo((): JSX.Element => {
  const billingEnabled = useSelector(getBillingEnabled);
  const logout = useLogout();

  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);

  function handleOpenMenu(event: React.MouseEvent<HTMLButtonElement>) {
    setAnchorEl(event.currentTarget);
  }

  function handleCloseMenu() {
    setAnchorEl(null);
  }

  function toggleDarkMode() {
    window.localStorage.setItem('darkMode', `${ window.localStorage.getItem('darkMode') !== 'true' }`);
  }

  return (
    <AppBar position="static">
      <Toolbar>
        <BankAccountSelector/>
        <Button to="/transactions" component={ RouterLink } color="inherit">Transactions</Button>
        <Button to="/expenses" component={ RouterLink } color="inherit">Expenses</Button>
        <Button to="/goals" component={ RouterLink } color="inherit">Goals</Button>
        <BalanceNavDisplay/>
        <div style={ { marginLeft: 'auto' } }/>
        <IconButton
          onClick={ handleOpenMenu }
          edge="start"
          color="inherit"
          aria-label="menu"
        >
          <MenuIcon/>
        </IconButton>
        <Menu
          id="user-menu"
          anchorEl={ anchorEl }
          keepMounted
          open={ open }
          onClose={ handleCloseMenu }
        >
          { billingEnabled &&
          <MenuItem
            onClick={ manageBilling }
          >
            <CreditCard className="mr-2"/>
            Billing
          </MenuItem>
          }

          <MenuItem
            onClick={ toggleDarkMode }
          >
            <LightMode className="mr-2"/>
            Toggle Dark Mode (Requires Reload)
          </MenuItem>

          <MenuItem
            onClick={ logout }
          >
            <ExitToApp className="mr-2"/>
            Logout
          </MenuItem>
        </Menu>
      </Toolbar>
    </AppBar>
  )
});

export default NavigationBar;
