/* eslint-disable max-len */
import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { AccountBalanceWalletOutlined, LocalAtmOutlined, PriceCheckOutlined, SavingsOutlined, ShoppingCartOutlined, TodayOutlined, TollOutlined } from '@mui/icons-material';
import { Infinity } from 'lucide-react';

import BudgetingSidebarTitle from './BudgetingSidebarTitle';
import PlaidBankStatusCard from '@monetr/interface/components/Layout/PlaidBankStatusCard';
import PlaidLastUpdatedCard from '@monetr/interface/components/Layout/PlaidLastUpdatedCard';
import SelectBankAccount from '@monetr/interface/components/Layout/SelectBankAccount';
import MBadge from '@monetr/interface/components/MBadge';
import MDivider from '@monetr/interface/components/MDivider';
import MSpan from '@monetr/interface/components/MSpan';
import { ReactElement } from '@monetr/interface/components/types';
import { useCurrentBalance } from '@monetr/interface/hooks/balances';
import { useSelectedBankAccount } from '@monetr/interface/hooks/bankAccounts';
import { useNextFundingDate } from '@monetr/interface/hooks/fundingSchedules';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { AmountType } from '@monetr/interface/util/amounts';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export interface BudgetingSidebarProps {
  className?: string;
}

export default function BudgetingSidebar(props: BudgetingSidebarProps): JSX.Element {
  const { data: locale } = useLocaleCurrency();
  const { data: bankAccount, isError } = useSelectedBankAccount();
  const balance = useCurrentBalance();


  const className = mergeTailwind(
    'w-72 h-full flex-none flex flex-col dark:border-r-dark-monetr-border border border-transparent items-center pb-6 lg:pb-4 overflow-auto',
    props.className,
  );

  // if (isLoading) {
  //   return null;
  // }

  if (isError) {
    return null;
  }

  function FreeToUse(): JSX.Element {
    switch (bankAccount?.accountSubType) {
      case 'checking':
      case 'savings':
        const valueClassName = mergeTailwind({
          'dark:text-dark-monetr-content-emphasis': balance?.free >= 0,
          'dark:text-dark-monetr-red': balance?.free < 0,
        });
        return (
          <div className='flex w-full justify-between'>
            <MSpan size='lg' weight='semibold' className='dark:text-dark-monetr-content-emphasis'>
              <AccountBalanceWalletOutlined />
              Free-To-Use:
            </MSpan>
            <MSpan size='lg' weight='semibold' className={ valueClassName }>
              { locale.formatAmount(balance?.free, AmountType.Stored) }
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
            <MSpan size='lg' weight='semibold' className='dark:text-dark-monetr-content-emphasis'>
              { locale.formatAmount(balance?.available, AmountType.Stored) }
            </MSpan>
          </div>
        );
    }

    return null;
  }

  function Limit(): JSX.Element {
    switch (bankAccount?.accountSubType) {
      case 'credit card':
        return (
          <div className='flex w-full justify-between'>
            <MSpan size='lg' weight='semibold' className='dark:text-dark-monetr-content-emphasis'>
              <Infinity />
              Limit:
            </MSpan>
            <MSpan size='lg' weight='semibold' className='dark:text-dark-monetr-content-emphasis'>
              { locale.formatAmount(balance?.limit, AmountType.Stored) }
            </MSpan>
          </div>
        );
    }

    return null;
  }

  return (
    <div className={ className }>
      <BudgetingSidebarTitle />
      <div className='flex h-full w-full flex-col items-center gap-4 px-2 pt-4'>
        <SelectBankAccount />
        <MDivider className='w-1/2' />

        <div className='flex w-full flex-col items-center gap-2 px-2'>
          <FreeToUse />
          <Available />
          <div className='flex w-full justify-between'>
            <MSpan size='lg' weight='semibold' className='dark:text-dark-monetr-content-emphasis'>
              <TollOutlined />
              Current:
            </MSpan>
            <MSpan size='lg' weight='semibold' className='dark:text-dark-monetr-content-emphasis'>
              { locale.formatAmount(balance?.current, AmountType.Stored) }
            </MSpan>
          </div>
          <Limit />
        </div>
        <MDivider className='w-1/2' />

        <div className='flex h-full w-full flex-col gap-2 pb-4'>
          <NavigationItem to={ `/bank/${bankAccount?.bankAccountId}/transactions` }>
            <ShoppingCartOutlined />
            <MSpan ellipsis color='inherit'>
              Transactions
            </MSpan>
          </NavigationItem>
          <NavigationItem to={ `/bank/${bankAccount?.bankAccountId}/expenses` }>
            <PriceCheckOutlined />
            <MSpan ellipsis color='inherit'>
              Expenses
            </MSpan>
            <MBadge className='ml-auto' size='sm'>
              { locale.formatAmount(balance?.expenses, AmountType.Stored) }
            </MBadge>
          </NavigationItem>
          <NavigationItem to={ `/bank/${bankAccount?.bankAccountId}/goals` }>
            <SavingsOutlined />
            <MSpan ellipsis color='inherit'>
              Goals
            </MSpan>
            <MBadge className='ml-auto' size='sm'>
              { locale.formatAmount(balance?.goals, AmountType.Stored) }
            </MBadge>
          </NavigationItem>
          <NavigationItem to={ `/bank/${bankAccount?.bankAccountId}/funding` }>
            <TodayOutlined />
            <MSpan ellipsis color='inherit'>
              Funding Schedules
            </MSpan>
            <NextFundingBadge />
          </NavigationItem>
        </div>
        <PlaidBankStatusCard />
        <PlaidLastUpdatedCard linkId={ bankAccount?.linkId } />
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
      { props.children }
    </Link>
  );
}


function NextFundingBadge(): JSX.Element {
  const next = useNextFundingDate();
  if (!next) return null;

  return (
    <MBadge className='ml-auto' size='sm'>
      { next }
    </MBadge>
  );
}
