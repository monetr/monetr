import {
  AccountBalance,
  CreditCard,
  ExitToApp,
  PriceCheck,
  Savings,
  Settings,
  ShoppingCart
} from '@mui/icons-material';
import { Button } from '@mui/material';
import BankAccountSelector from 'components/BankAccounts/BankAccountSelector';
import React from 'react';
import { useSelector } from 'react-redux';
import { Link as RouterLink } from 'react-router-dom';
import { getBillingEnabled } from 'shared/bootstrap/selectors';

import 'components/Layout/Sidebar/styles/Sidebar.scss'

export default function Sidebar(): JSX.Element {
  const billingEnabled = useSelector(getBillingEnabled);

  return (
    <div className="sidebar fixed top-0 bottom-0 left-0 hidden lg:flex lg:flex-shrink-0 lg:w-64">
      <div className="w-full h-full flex flex-col text-white">
        <div className="flex justify-start p-2.5 flex-shrink-0">
          <BankAccountSelector/>
        </div>
        <div className="flex-1 flex flex-col gap-2.5 p-2.5">
          <Button
            className="justify-start text-lg"
            to="/transactions"
            component={ RouterLink }
            color="inherit"
          >
            <ShoppingCart className="mr-2.5"/>
            Transactions
          </Button>
          <Button
            className="justify-start text-lg"
            to="/expenses"
            component={ RouterLink }
            color="inherit"
          >
            <PriceCheck className="mr-2.5"/>
            Expenses
          </Button>
          <Button
            className="justify-start text-lg"
            to="/goals"
            component={ RouterLink }
            color="inherit"
          >
            <Savings className="mr-2.5"/>
            Goals
          </Button>
          <Button
            className="justify-start text-lg"
            to="/accounts"
            component={ RouterLink }
            color="inherit"
          >
            <AccountBalance className="mr-2.5"/>
            Accounts
          </Button>
        </div>
        <div className="flex justify-start p-2.5 flex-col gap-2.5">
          { billingEnabled &&
            <Button
              className="justify-start"
              to="/subscription"
              component={ RouterLink }
              color="inherit"
            >
              <CreditCard className="mr-2"/>
              Subscription
            </Button>
          }
          <Button
            className="justify-start"
            to="/settings"
            component={ RouterLink }
            color="inherit"
          >
            <Settings className="mr-2"/>
            Settings
          </Button>
          <Button

            className="justify-start"
            to="/logout"
            component={ RouterLink }
            color="inherit"
          >
            <ExitToApp className="mr-2"/>
            Logout
          </Button>
        </div>
      </div>
    </div>
  )
}
