/* eslint-disable max-len */
import React, { Fragment } from 'react';
import { AccountBalanceOutlined, AccountBalanceWalletOutlined, HomeOutlined, KeyboardArrowDown, KeyboardArrowRight, LocalAtmOutlined, MenuOutlined, MoreVert, PriceCheckOutlined, SavingsOutlined, ShoppingCartOutlined, TodayOutlined, TollOutlined } from '@mui/icons-material';
import { Avatar, Badge, styled } from '@mui/material';
import clsx from 'clsx';

import BankSidebar from './new/BankSidebar';
import ExpenseList from './new/ExpenseList';

import { ReactElement } from 'components/types';
import { useIconSearch } from 'hooks/useIconSearch';
import useTheme from 'hooks/useTheme';
import MDivider from 'components/MDivider';

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
    <Fragment>
      <div className='w-full h-12 flex items-center px-4 gap-4'>
        <MenuOutlined className='visible lg:hidden dark:text-dark-monetr-content-emphasis cursor-pointer' />
        <span className='text-2xl dark:text-dark-monetr-content-emphasis font-bold flex gap-2 items-center'>
          <ShoppingCartOutlined />
          Transactions
        </span>
      </div>
      <div className='w-full h-full overflow-y-auto min-w-0'>
        <ul className='w-full'>
          <li>
            <ul className='flex gap-2 flex-col'>
              <TransactionDateHeader date='1 July, 2023' />
              <TransactionItem name='Lunds & Byerlys' category='Food & Drink' amount='$248.14' pending />
            </ul>
          </li>
          <li>
            <ul className='flex gap-2 flex-col'>
              <TransactionDateHeader date='28 June, 2023' />
              <TransactionItem name='Starbucks Coffee' category='Food & Drink' amount='$10.24' from='Eating Out Budget' />
              <TransactionItem name='Arbys' category='Food & Drink' amount='$5.67' />
              <TransactionItem name='GitHub' category='Software' amount='$10.24' />
              <TransactionItem name='Target' category='Shops' amount='$10.24' />
              <TransactionItem name='Rocket Mortgage' category='Loan' amount='$1800.00' />
            </ul>
          </li>
          <li>
            <ul className='flex gap-2 flex-col'>
              <TransactionDateHeader date='25 June, 2023' />
              <TransactionItem name='Discord' category='Games & Entertainment' amount='$10.24' from='Discord' />
              <TransactionItem name='GitLab Inc' category='Service' amount='$10.24' from='Tools' />
              <TransactionItem name='Buildkite' category='Transfer' amount='$10.24' />
              <TransactionItem name='Sentry' category='Shops' amount='$10.24' />
              <TransactionItem name='Ngrok' category='Transfer' amount='$10.24' />
            </ul>
          </li>
          <li>
            <ul className='flex gap-2 flex-col'>
              <TransactionDateHeader date='21 June, 2023' />
              <TransactionItem name='GitHub' category='Service' amount='$10.24' />
              <TransactionItem name='Plaid' category='Service' amount='$2.40' />
              <TransactionItem name='Elliots Contribution' category='Payroll' amount='+ $250.00' />
              <TransactionItem name='FreshBooks' category='Accounting and Bookkeeping' amount='$17.00' />
            </ul>
          </li>
        </ul>
      </div>
    </Fragment>
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


interface TransactionDateHeaderProps {
  date: string;
}

function TransactionDateHeader(props: TransactionDateHeaderProps): JSX.Element {
  return (
    <li className='sticky top-0 z-10 h-10 flex items-center backdrop-blur-sm bg-gradient-to-t from-transparent dark:to-dark-monetr-background via-90%'>
      <span className='dark:text-dark-monetr-content-subtle font-semibold text-base z-10 px-3 md:px-4'>
        {props.date}
      </span>
    </li>
  );
}

interface TransactionItemProps {
  name: string;
  from?: string;
  pending?: boolean;
  category: string;
  amount: string;
}

function TransactionItem(props: TransactionItemProps): JSX.Element {
  const SpentFrom = () => {
    if (props.from) {
      return (
        <span className='dark:text-dark-monetr-content-emphasis font-bold text-sm md:text-base text-ellipsis whitespace-nowrap overflow-hidden min-w-0'>
          {props.from || 'Free-To-Use'}
        </span>
      );
    }

    return (
      <span className='dark:text-dark-monetr-content font-medium text-sm md:text-base text-ellipsis whitespace-nowrap overflow-hidden min-w-0'>
        Free-To-Use
      </span>
    );
  };

  return (
    <li className='w-full px-1 md:px-2'>
      <div className='flex rounded-lg hover:bg-zinc-600 gap-1 md:gap-4 group px-2 py-1 h-full cursor-pointer md:cursor-auto'>
        <div className='w-full md:w-1/2 flex flex-row gap-4 items-center flex-1 min-w-0'>
          <TransactionIcon name={ props.name } pending={ props.pending } />
          <div className='flex flex-col overflow-hidden min-w-0'>
            <span className='text-zinc-50 font-semibold text-base w-full overflow-hidden text-ellipsis whitespace-nowrap min-w-0'>
              {props.name}
            </span>
            <span className='hidden md:block dark:text-dark-monetr-content font-medium text-sm w-full overflow-hidden text-ellipsis whitespace-nowrap min-w-0'>
              {props.category}
            </span>
            <span className='flex md:hidden text-sm w-full overflow-hidden text-ellipsis whitespace-nowrap min-w-0'>
              <span className='flex-none dark:text-dark-monetr-content font-medium text-sm text-ellipsis whitespace-nowrap overflow-hidden min-w-0'>
                Spent from
              </span>
              &nbsp;
              <SpentFrom />
            </span>
          </div>
        </div>
        <div className='hidden md:flex w-1/2 overflow-hidden flex-1 min-w-0 items-center'>
          <span className='flex-none dark:text-dark-monetr-content font-medium text-base text-ellipsis whitespace-nowrap overflow-hidden min-w-0'>
            Spent from
          </span>
          &nbsp;
          <SpentFrom />
        </div>
        <div className='flex md:min-w-[8em] shrink-0 justify-end gap-2 items-center'>
          <span className='text-end text-red-500 font-semibold'>
            {props.amount}
          </span>
          <KeyboardArrowRight className='dark:text-dark-monetr-content-subtle dark:group-hover:text-dark-monetr-content-emphasis flex-none md:cursor-pointer' />
        </div>
      </div>
    </li>
  );
}

interface TransactionIconProps {
  name: string;
  pending?: boolean;
}

function TransactionIcon(props: TransactionIconProps): JSX.Element {
  const windTheme = useTheme();
  // Try to retrieve the icon. If the icon cannot be retrieved or icons are not currently enabled in the application
  // config then this will simply return null.
  const icon = useIconSearch(props.name);
  const IconContent = () => {
    if (icon?.svg) {
      // It is possible for colors to be missing for a given icon. When this happens just fall back to a black color.
      const colorStyles = icon?.colors?.length > 0 ?
        { backgroundColor: `#${icon.colors[0]}` } :
        { backgroundColor: '#000000' };

      const styles = {
        WebkitMaskImage: `url(data:image/svg+xml;base64,${icon.svg})`,
        WebkitMaskRepeat: 'no-repeat',
        height: '30px',
        width: '30px',
        ...colorStyles,
      };

      return (
        <div className='dark:bg-white flex items-center justify-center h-10 w-10 rounded-full'>
          <div style={ styles } />
        </div>
      );
    }

    // If we have no icon to work with then create an avatar with the first character of the transaction name.
    const letter = props.name.toUpperCase().charAt(0);
    return (
      <Avatar className='dark:bg-dark-monetr-background-subtle dark:text-dark-monetr-content h-10 w-10'>
        {letter}
      </Avatar>
    );
  };
  const StyledBadge = styled(Badge)(() => ({
    '& .MuiBadge-badge': {
      backgroundColor: windTheme.tailwind.colors['blue']['500'],
      color:  windTheme.tailwind.colors['blue']['500'],
      boxShadow: `0 0 0 2px ${windTheme.tailwind.colors['dark-monetr']['background']['DEFAULT']}`,
      '&::after': {
        position: 'absolute',
        top: 0,
        left: 0,
        width: '100%',
        height: '100%',
        borderRadius: '50%',
        animation: 'ripple 1.2s infinite ease-in-out',
        border: '1px solid currentColor',
        content: '""',
      },
    },
    '@keyframes ripple': {
      '0%': {
        transform: 'scale(.8)',
        opacity: 1,
      },
      '100%': {
        transform: 'scale(2.4)',
        opacity: 0,
      },
    },
  }));

  if (props.pending) {
    return (
      <StyledBadge
        overlap='circular'
        anchorOrigin={ { vertical: 'bottom', horizontal: 'right' } }
        variant='dot'
      >
        <IconContent />
      </StyledBadge>
    );
  }

  return (
    <IconContent />
  );
}
