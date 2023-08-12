/* eslint-disable max-len */
import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { AccountBalanceWalletOutlined, LocalAtmOutlined, MoreVert, PriceCheckOutlined, SavingsOutlined, ShoppingCartOutlined, TodayOutlined, TollOutlined } from '@mui/icons-material';

import MDivider from 'components/MDivider';
import MSelectAccount from 'components/MSelectAccount';
import { ReactElement } from 'components/types';
import { useCurrentBalance } from 'hooks/balances';
import { useSelectedBankAccount } from 'hooks/bankAccounts';
import { useNextFundingDate } from 'hooks/fundingSchedules';
import { useLink } from 'hooks/links';
import mergeTailwind from 'util/mergeTailwind';
import MBadge from 'components/MBadge';
import MSpan from 'components/MSpan';

export interface BudgetingSidebarProps {
  className?: string;
}

export default function BudgetingSidebar(props: BudgetingSidebarProps): JSX.Element {
  const { result: bankAccount, isLoading, isError } = useSelectedBankAccount();
  const link = useLink(bankAccount?.linkId);
  const balance = useCurrentBalance();

  if (isLoading)  {
    return null;
  }

  if (isError) {
    return null;
  }

  const className = mergeTailwind(
    // 'hidden lg:w-72 h-full flex-none lg:flex flex-col dark:border-r-dark-monetr-border border border-transparent items-center',
    'w-72 h-full flex-none flex flex-col dark:border-r-dark-monetr-border border border-transparent items-center',
    props.className,
  );

  function FreeToUse(): JSX.Element {
    switch (bankAccount?.accountSubType) {
      case 'checking':
      case 'savings':
        return (
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
        );
    }

    return null;
  }

  function Available(): JSX.Element {
    switch (bankAccount?.accountSubType) {
      case 'checking':
      case 'savings':
        return (
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
        );
    }

    return null;
  }


  return (
    <div className={ className }>
      <div className='w-full dark:hover:bg-dark-monetr-background-emphasis dark:text-dark-monetr-content-emphasis h-12 flex items-center p-2'>
        <span className='font-semibold text-ellipsis whitespace-nowrap overflow-hidden text-xl'>
          { link?.getName() }
        </span>
        <MoreVert className='ml-auto' />
      </div>
      <MDivider className='w-1/2' />
      <div className='h-full flex flex-col gap-4 px-2 py-4 w-full items-center'>
        <MSelectAccount />
        <MDivider className='w-1/2' />

        <div className='w-full flex items-center flex-col gap-2 px-2'>
          <FreeToUse />
          <Available />
          <div className='w-full flex justify-between br-red-100'>
            <MSpan size='lg' weight='semibold'>
              <TollOutlined />
              Current:
            </MSpan>
            &nbsp;
            <MSpan size='lg' weight='semibold'>
              { balance?.getCurrentString() }
            </MSpan>
          </div>
        </div>
        <MDivider className='w-1/2' />

        <div className='h-full w-full flex flex-col gap-2 overflow-y-auto'>
          <NavigationItem to={ `/bank/${bankAccount?.bankAccountId}/transactions` }>
            <ShoppingCartOutlined />
            <MSpan ellipsis variant='inherit'>
              Transactions
            </MSpan>
          </NavigationItem>
          <NavigationItem to={ `/bank/${bankAccount?.bankAccountId}/expenses` }>
            <PriceCheckOutlined />
            <MSpan ellipsis variant='inherit'>
              Expenses
            </MSpan>
            <MBadge className='ml-auto'>
              { balance?.getExpensesString() }
            </MBadge>
          </NavigationItem>
          <NavigationItem  to={ `/bank/${bankAccount?.bankAccountId}/goals` }>
            <SavingsOutlined />
            <MSpan ellipsis variant='inherit'>
              Goals
            </MSpan>
            <MBadge className='ml-auto'>
              { balance?.getGoalsString() }
            </MBadge>
          </NavigationItem>
          <NavigationItem to={ `/bank/${bankAccount?.bankAccountId}/funding` }>
            <TodayOutlined />
            <MSpan ellipsis variant='inherit'>
              Funding Schedules
            </MSpan>
            <NextFundingBadge />
          </NavigationItem>
        </div>
      </div>
    </div>
  );
}

interface NavigationItemProps {
  children: ReactElement;
  to: string;
}

function NavigationItem(props: NavigationItemProps): JSX.Element {
  const location = useLocation();
  const active = location.pathname.endsWith(props.to.replaceAll('.', ''));

  const className = mergeTailwind({
    'dark:bg-dark-monetr-background-emphasis': active,
    'dark:text-dark-monetr-content-emphasis': active,
    'dark:text-dark-monetr-content-subtle': !active,
    'font-semibold': active,
    'font-medium': !active,
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
    <Link className={ className } to={ props.to }>
      {props.children}
    </Link>
  );
}


function NextFundingBadge(): JSX.Element {
  const next = useNextFundingDate();
  if (!next) return null;

  return (
    <MBadge className='ml-auto'>
      { next }
    </MBadge>
  );
}
