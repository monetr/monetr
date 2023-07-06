/* eslint-disable max-len */
import React from 'react';
import { AccountBalanceOutlined, AccountBalanceWalletOutlined, HomeOutlined, KeyboardArrowDown, LocalAtmOutlined, MoreVert, PriceCheckOutlined, SavingsOutlined, ShoppingCartOutlined, TodayOutlined, TollOutlined } from '@mui/icons-material';

import MDivider from 'components/MDivider';
import { ReactElement } from 'components/types';
import { useCurrentBalance } from 'hooks/balances';
import { useSelectedBankAccount } from 'hooks/bankAccounts';
import mergeTailwind from 'util/mergeTailwind';

export default function BudgetingSidebar(): JSX.Element {
  const { bankAccount, isLoading, isError } = useSelectedBankAccount();
  const balance = useCurrentBalance();

  if (isLoading)  {
    return null;
  }

  if (isError) {
    return null;
  }

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
            <span className='text-ellipsis whitespace-nowrap overflow-hidden'>
              { bankAccount.name }
            </span>
            <span className='ml-auto text-xs dark:bg-dark-monetr-brand dark:text-dark-monetr-content-emphasis rounded-sm py-0.5 px-1'>
              { bankAccount.mask }
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
              { balance?.getFreeToUseString() }
            </span>
          </div>
          <div className='w-full flex justify-between dark:text-monetr-dark-content'>
            <span className='flex gap-2 items-center text-lg font-semibold'>
              <LocalAtmOutlined />
              Available:
            </span>
            &nbsp;
            <span className='text-lg font-semibold'>
              { balance?.getAvailableString() }
            </span>
          </div>
          <div className='w-full flex justify-between dark:text-monetr-dark-content'>
            <span className='flex gap-2 items-center text-lg font-semibold'>
              <TollOutlined />
              Current:
            </span>
            &nbsp;
            <span className='text-lg font-semibold'>
              { balance?.getCurrentString() }
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
              { balance?.getExpensesString() }
            </span>
          </NavigationItem>
          <NavigationItem>
            <SavingsOutlined />
            Goals
            <span className='ml-auto text-sm bg-monetr-brand dark:text-dark-monetr-content-emphasis rounded-md py-0.5 px-1.5'>
              { balance?.getGoalsString() }
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

interface NavigationItemProps {
  children: ReactElement;
  active?: boolean;
}

function NavigationItem(props: NavigationItemProps): JSX.Element {
  const className = mergeTailwind({
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
