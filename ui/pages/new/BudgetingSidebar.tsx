/* eslint-disable max-len */
import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { AccountBalanceWalletOutlined, LocalAtmOutlined, MoreVert, PriceCheckOutlined, SavingsOutlined, ShoppingCartOutlined, TodayOutlined, TollOutlined } from '@mui/icons-material';

import MBadge from 'components/MBadge';
import MDivider from 'components/MDivider';
import MSelectAccount from 'components/MSelectAccount';
import MSpan from 'components/MSpan';
import { ReactElement } from 'components/types';
import { useCurrentBalance } from 'hooks/balances';
import { useSelectedBankAccount } from 'hooks/bankAccounts';
import { useNextFundingDate } from 'hooks/fundingSchedules';
import { useLink } from 'hooks/links';
import mergeTailwind from 'util/mergeTailwind';

export interface BudgetingSidebarProps {
  className?: string;
}

export default function BudgetingSidebar(props: BudgetingSidebarProps): JSX.Element {
  const { data: bankAccount, isLoading, isError } = useSelectedBankAccount();
  const { data: link } = useLink(bankAccount?.linkId);
  const balance = useCurrentBalance();

  if (isLoading) {
    return null;
  }

  if (isError) {
    return null;
  }

  const className = mergeTailwind(
    'w-72 h-full flex-none flex flex-col dark:border-r-dark-monetr-border border border-transparent items-center',
    props.className,
  );

  function FreeToUse(): JSX.Element {
    switch (bankAccount?.accountSubType) {
      case 'checking':
      case 'savings':
        return (
          <div className='flex w-full justify-between'>
            <MSpan size='lg' weight='semibold' className='dark:text-dark-monetr-content-emphasis'>
              <AccountBalanceWalletOutlined />
              Free-To-Use:
            </MSpan>
            &nbsp;
            <MSpan size='lg' weight='semibold' className='dark:text-dark-monetr-content-emphasis'>
              {balance?.getFreeToUseString()}
            </MSpan>
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
          <div className='flex w-full justify-between'>
            <MSpan size='lg' weight='semibold' className='dark:text-dark-monetr-content-emphasis'>
              <LocalAtmOutlined />
              Available:
            </MSpan>
            &nbsp;
            <MSpan size='lg' weight='semibold' className='dark:text-dark-monetr-content-emphasis'>
              {balance?.getAvailableString()}
            </MSpan>
          </div>
        );
    }

    return null;
  }



  return (
    <div className={ className }>
      <div className='flex h-12 w-full items-center p-2 dark:text-dark-monetr-content-emphasis dark:hover:bg-dark-monetr-background-emphasis'>
        <span className='truncate text-xl font-semibold'>
          {link?.getName()}
        </span>
        <MoreVert className='ml-auto' />
      </div>
      <MDivider className='w-1/2' />
      <div className='flex h-full w-full flex-col items-center gap-4 px-2 py-4'>
        <MSelectAccount />
        <MDivider className='w-1/2' />

        <div className='flex w-full flex-col items-center gap-2 px-2'>
          <FreeToUse />
          <Available />
          <div className='flex w-full justify-between'>
            <MSpan size='lg' weight='semibold' className='dark:text-dark-monetr-content-emphasis'>
              <TollOutlined />
              Current:
            </MSpan>
            &nbsp;
            <MSpan size='lg' weight='semibold' className='dark:text-dark-monetr-content-emphasis'>
              {balance?.getCurrentString()}
            </MSpan>
          </div>
        </div>
        <MDivider className='w-1/2' />

        <div className='flex h-full w-full flex-col gap-2 overflow-y-auto'>
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
              {balance?.getExpensesString()}
            </MBadge>
          </NavigationItem>
          <NavigationItem to={ `/bank/${bankAccount?.bankAccountId}/goals` }>
            <SavingsOutlined />
            <MSpan ellipsis variant='inherit'>
              Goals
            </MSpan>
            <MBadge className='ml-auto'>
              {balance?.getGoalsString()}
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
      {next}
    </MBadge>
  );
}
