import { CreditCard, ExitToApp } from '@mui/icons-material';
import MenuIcon from '@mui/icons-material/Menu';
import { AppBar, Button, IconButton, Menu, MenuItem, Toolbar } from '@mui/material';
import BalanceNavDisplay from 'components/Balance/BalanceNavDisplay';
import BankAccountSelector from 'components/BankAccounts/BankAccountSelector';
import React, { useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Link as RouterLink } from 'react-router-dom';
import logout from 'shared/authentication/actions/logout';
import manageBilling from 'shared/billing/actions/manageBilling';
import { getBillingEnabled } from 'shared/bootstrap/selectors';

const NavigationBar = React.memo((): JSX.Element => {
  const billingEnabled = useSelector(getBillingEnabled);

  const dispatch = useDispatch();
  const dispatchLogout = () => logout()(dispatch);

  const toggleDarkMode = () => window.localStorage.setItem('darkMode', `${ window.localStorage.getItem('darkMode') !== 'true' }`);

  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);
  const handleOpenMenu = (event: React.MouseEvent<HTMLButtonElement>) => setAnchorEl(event.currentTarget);
  const handleCloseMenu = () => setAnchorEl(null);

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
            Toggle Dark Mode (Requires Reload)
          </MenuItem>

          <MenuItem
            onClick={ dispatchLogout }
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