import React, { Fragment } from 'react';
import { AccountBalanceOutlined, CreditCardOutlined, ExitToAppOutlined, HelpOutlined, HomeOutlined, MenuOpenOutlined, PriceCheckOutlined, SavingsOutlined, SettingsOutlined, ShoppingCartOutlined, TodayOutlined } from '@mui/icons-material';

import MSidebarButton from './MSidebarButton';

import clsx from 'clsx';
import MLogo from 'components/MLogo';
import { useAppConfiguration } from 'hooks/useAppConfiguration';

export interface MSidebarProps {
  open?: boolean;
  setClosed?: () => void;
}

export default function MSidebar(props: MSidebarProps): JSX.Element {
  const config = useAppConfiguration();

  function SubscriptionButton(): JSX.Element {
    if (!config.billingEnabled) return null;

    return (
      <MSidebarButton to="/subscription">
        <CreditCardOutlined />
        Subscription
      </MSidebarButton>
    );
  }

  function SidebarButtons(): JSX.Element {
    return (
      <nav className="h-full flex">
        <ul role="list" className="flex flex-col flex-1 gap-y-3">
          <li>
            <ul className="flex gap-y-3 flex-col">
              <MSidebarButton to="/">
                <HomeOutlined />
                Overview
              </MSidebarButton>
              <MSidebarButton to="/transactions">
                <ShoppingCartOutlined />
                Transactions
              </MSidebarButton>
              <MSidebarButton to="/expenses">
                <PriceCheckOutlined />
                Expenses
              </MSidebarButton>
              <MSidebarButton to="/goals">
                <SavingsOutlined />
                Goals
              </MSidebarButton>
              <MSidebarButton to="/funding">
                <TodayOutlined />
                Funding Schedules
              </MSidebarButton>
              <MSidebarButton to="/accounts">
                <AccountBalanceOutlined />
                Accounts
              </MSidebarButton>
            </ul>
          </li>
          <li className="mt-auto">
            <ul className="flex gap-y-3 flex-col">
              <SubscriptionButton />
              <MSidebarButton to="/settings">
                <SettingsOutlined />
                Settings
              </MSidebarButton>
              <MSidebarButton to="https://monetr.app/help">
                <HelpOutlined />
                Help
              </MSidebarButton>
              <MSidebarButton to="/logout">
                <ExitToAppOutlined />
                Logout
              </MSidebarButton>
            </ul>
          </li>
        </ul>
      </nav>
    );
  }

  return (
    <Fragment>
      <div className="lg:hidden">
        <div className={ clsx({ 'hidden': !props.open }) }>
          <Fragment>
            <div
              className="absolute z-40 w-screen h-screen left-0 top-0 bottom-0 right-0 bg-purple-800 opacity-50"
            />
            <div className="flex flex-col w-64 z-50 left-0 top-0 bottom-0 fixed bg-purple-800">
              <div className="pb-4 px-6 overflow-y-auto flex-col gap-y-5 flex-grow flex">
                <div className="items-center flex-shrink-0 h-16 flex gap-x-3">
                  <MLogo className="h-8 w-auto" />
                  <span className="text-gray-50 text-2xl">
                    monetr
                  </span>
                  <button
                    onClick={ () => props.setClosed && props.setClosed() }
                    className="ml-auto text-white"
                  >
                    <MenuOpenOutlined />
                  </button>
                </div>
                <SidebarButtons />
              </div>
            </div>
          </Fragment>
        </div>
      </div>

      <div className="lg:flex lg:flex-col lg:w-64 lg:z-50 left-0 top-0 bottom-0 lg:fixed bg-purple-800 hidden">
        <div className="pb-4 px-6 overflow-y-auto flex-col gap-y-5 flex-grow flex">
          <div className="items-center flex-shrink-0 h-16 flex gap-x-3">
            <MLogo className="h-8 w-auto" />
            <span className="text-gray-50 text-2xl">
              monetr
            </span>
          </div>
          <SidebarButtons />
        </div>
      </div>
    </Fragment>
  );
}
