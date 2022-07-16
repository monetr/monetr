import { useAppConfiguration } from 'hooks/useAppConfiguration';
import React from 'react';
import { useSelector } from 'react-redux';
import { Link as RouterLink } from 'react-router-dom';
import {
  AccountBalance,
  CreditCard,
  ExitToApp, Menu,
  PriceCheck,
  Savings,
  Settings,
  ShoppingCart,
} from '@mui/icons-material';
import { Button, IconButton } from '@mui/material';

import classnames from 'classnames';
import BankAccountSelector from 'components/BankAccounts/BankAccountSelector';
import SidebarButton from 'components/Layout/Sidebar/SidebarButton';

import 'components/Layout/Sidebar/styles/Sidebar.scss';

interface SidebarProps {
  closed?: boolean;
  onToggleSidebar?: () => void;
  closeSidebar?: () => void;
}

export default function Sidebar(props: SidebarProps): JSX.Element {
  const {
    billingEnabled,
  } = useAppConfiguration();

  return (
    <div className={ classnames('sidebar fixed top-0 bottom-0 left-0 lg:flex lg:flex-shrink-0 lg:w-64 w-full z-50', {
      'block': !!!props.closed,
      'hidden': !!props.closed,
    }) }>
      <div className="w-full h-full flex flex-col text-white">
        <div className="flex">
          <div className="basis-1/5 block lg:hidden flex justify-start items-center pl-2.5">
            <IconButton onClick={ props.onToggleSidebar } aria-label="menu" className="text-white">
              <Menu />
            </IconButton>
          </div>
          <div className="basis-4/5 lg:basis-full flex justify-start p-2.5 flex-shrink-0">
            <BankAccountSelector />
          </div>
        </div>
        <div className="flex-1 flex flex-col pl-2.5 pt-2.5 pr-2.5 lg:pr-0">
          <SidebarButton onClick={ props.closeSidebar } to="/transactions">
            <ShoppingCart className="mr-2.5" />
            Transactions
          </SidebarButton>
          <SidebarButton onClick={ props.closeSidebar } to="/expenses">
            <PriceCheck className="mr-2.5" />
            Expenses
          </SidebarButton>
          <SidebarButton onClick={ props.closeSidebar } to="/goals">
            <Savings className="mr-2.5" />
            Goals
          </SidebarButton>
          <SidebarButton onClick={ props.closeSidebar } to="/accounts">
            <AccountBalance className="mr-2.5" />
            Accounts
          </SidebarButton>
        </div>
        <div className="flex justify-start p-2.5 flex-col gap-2.5">
          { billingEnabled &&
            <Button
              onClick={ props.closeSidebar }
              className="justify-start"
              to="/subscription"
              component={ RouterLink }
              color="inherit"
            >
              <CreditCard className="mr-2" />
              Subscription
            </Button>
          }
          <Button
            onClick={ props.closeSidebar }
            className="justify-start"
            to="/settings"
            component={ RouterLink }
            color="inherit"
          >
            <Settings className="mr-2" />
            Settings
          </Button>
          <Button
            onClick={ props.closeSidebar }
            className="justify-start"
            to="/logout"
            component={ RouterLink }
            color="inherit"
          >
            <ExitToApp className="mr-2" />
            Logout
          </Button>
        </div>
      </div>
    </div>
  );
}
