/* eslint-disable max-len */
import React, { Fragment } from 'react';
import { AccountBalanceOutlined, AccountBalanceWalletOutlined, HomeOutlined, KeyboardArrowDown, LocalAtmOutlined, MoreVert, PriceCheckOutlined, SavingsOutlined, ShoppingCartOutlined, TodayOutlined, TollOutlined } from '@mui/icons-material';
import clsx from 'clsx';

import BankSidebar from './new/BankSidebar';
import ExpenseList from './new/ExpenseList';
import TransactionList from './new/TransactionList';

import MDivider from 'components/MDivider';
import { ReactElement } from 'components/types';

interface MonetrWrapperProps {
  children: ReactElement;
}

export default function MonetrWrapper(props: MonetrWrapperProps): JSX.Element {
  return (
    <div className='w-full h-full dark:bg-dark-monetr-background flex'>
      <BankSidebar />
      <div className='w-full h-full flex min-w-0'>
        { props.children }
      </div>
    </div>
  );
}

interface BankViewProps {
  children: ReactElement;
}

export function BankView(props: BankViewProps): JSX.Element {
  return (
    <Fragment>
      <BudgetingSideBar />
      <div className='w-full h-full min-w-0 flex flex-col'>
        { props.children}
      </div>
    </Fragment>
  );
}

export function TransactionsView(): JSX.Element {
  return (
    <TransactionList />
  );
}

export function ExpensesView(): JSX.Element {
  return (
    <ExpenseList />
  );
}

interface NavigationItemProps {
  children: ReactElement;
  active?: boolean;
}

function NavigationItem(props: NavigationItemProps): JSX.Element {
  const className = clsx({
    'bg-zinc-700': props.active,
    'dark:text-dark-monetr-content-emphasis': props.active,
    'dark:text-dark-monetr-content-subtle': !props.active,
    'font-semibold': props.active,
    'font-medium': !props.active,
  }, [
    'align-middle',
    'cursor-pointer',
    'flex',
    'text-lg',
    'gap-2',
    'dark:hover:bg-dark-monetr-background-emphasis',
    'dark:hover:text-dark-monetr-content-emphasis',
    'items-center',
    'px-2',
    'py-1',
    'rounded-md',
    'w-full',
  ]);

  return (
    <a className={ className }>
      {props.children}
    </a>
  );
}

function BudgetingSideBar(): JSX.Element {
  return (
    <div className='hidden lg:w-72 h-full flex-none lg:flex flex-col dark:border-r-dark-monetr-border border border-transparent items-center'>
      <div className='w-full dark:hover:bg-dark-monetr-background-emphasis dark:text-dark-monetr-content-emphasis h-12 flex items-center p-2'>
        <span className='font-semibold text-ellipsis whitespace-nowrap overflow-hidden shadow-'>
          Navy Federal Credit Union
        </span>
        <MoreVert className='ml-auto' />
      </div>
      <MDivider className='w-1/2' />
      <div className='h-full flex flex-col gap-4 px-2 py-4 w-full items-center'>
        <div className='w-full'>
          <span className='cursor-pointer dark:hover:bg-dark-monetr-background-emphasis dark:hover:text-dark-monetr-content-emphasis dark:text-dark-monetr-content-subtle text-lg flex items-center font-semibold gap-2 p-1 align-middle rounded-md'>
            <AccountBalanceOutlined />
            Checking
            <span className='ml-auto text-xs dark:bg-dark-monetr-brand dark:text-dark-monetr-content-emphasis rounded-sm py-0.5 px-1'>
              4567
            </span>
            <KeyboardArrowDown />
          </span>
        </div>
        <MDivider className='w-1/2' />

        <div className='w-full flex items-center flex-col gap-2 px-2'>
          <div className='w-full flex justify-between dark:text-monetr-dark-content'>
            <span className='flex gap-2 items-center text-lg font-semibold'>
              <AccountBalanceWalletOutlined />
              Free-To-Use:
            </span>
            &nbsp;
            <span className='text-lg font-semibold'>
               $154.65
            </span>
          </div>
          <div className='w-full flex justify-between dark:text-monetr-dark-content'>
            <span className='flex gap-2 items-center text-lg font-semibold'>
              <LocalAtmOutlined />
              Available:
            </span>
            &nbsp;
            <span className='text-lg font-semibold'>
              $4,241.30
            </span>
          </div>
          <div className='w-full flex justify-between dark:text-monetr-dark-content'>
            <span className='flex gap-2 items-center text-lg font-semibold'>
              <TollOutlined />
              Current:
            </span>
            &nbsp;
            <span className='text-lg font-semibold'>
              $4,241.30
            </span>
          </div>
        </div>
        <MDivider className='w-1/2' />

        <div className='h-full w-full flex flex-col gap-2 overflow-y-auto'>
          <NavigationItem>
            <HomeOutlined />
            Overview
          </NavigationItem>
          <NavigationItem active>
            <ShoppingCartOutlined />
            Transactions
          </NavigationItem>
          <NavigationItem>
            <PriceCheckOutlined />
            Expenses
            <span className='ml-auto text-sm bg-monetr-brand dark:text-dark-monetr-content-emphasis rounded-md py-0.5 px-1.5'>
              $1,554.43
            </span>
          </NavigationItem>
          <NavigationItem>
            <SavingsOutlined />
            Goals
            <span className='ml-auto text-sm bg-monetr-brand dark:text-dark-monetr-content-emphasis rounded-md py-0.5 px-1.5'>
              $2,549.43
            </span>
          </NavigationItem>
          <NavigationItem>
            <TodayOutlined />
            Funding Schedules
            <span className='ml-auto text-sm bg-monetr-brand dark:text-dark-monetr-content-emphasis rounded-md py-0.5 px-1.5'>
              7/15
            </span>
          </NavigationItem>
        </div>
      </div>
    </div>
  );
}

