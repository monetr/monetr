/* eslint-disable max-len */
import React, { Fragment, useState } from 'react';
import { AccountBalance, AccountBalanceOutlined, AccountBalanceWalletOutlined, HomeOutlined, KeyboardArrowDown, KeyboardArrowRight, LocalAtmOutlined, Logout, MenuOutlined, MoreVert, PriceCheckOutlined, SavingsOutlined, Settings, ShoppingCartOutlined, TodayOutlined, TollOutlined } from '@mui/icons-material';
import { Avatar, Badge, styled } from '@mui/material';
import clsx from 'clsx';

import ExpenseList from './new/ExpenseList';

import { Logo } from 'assets';
import { ReactElement } from 'components/types';
import { useInstitution } from 'hooks/institutions';
import { useIconSearch } from 'hooks/useIconSearch';
import useTheme from 'hooks/useTheme';
import { rrulestr } from 'rrule';

interface MonetrWrapperProps {
  children: ReactElement;
}

export default function MonetrWrapper(props: MonetrWrapperProps): JSX.Element {
  return (
    <div className='w-full h-full bg-zinc-900 flex'>
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
        <MenuOutlined className='visible lg:hidden text-zinc-50 cursor-pointer' />
        <span className='text-2xl text-zinc-50 font-bold flex gap-2 items-center'>
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
    'text-zinc-50': props.active,
    'font-semibold': props.active,
    'text-zinc-400': !props.active,
    'font-medium': !props.active,
  }, [
    'align-middle',
    'cursor-pointer',
    'flex',
    'text-lg',
    'gap-2',
    'hover:bg-zinc-700',
    'hover:text-zinc-50',
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

function BankSidebar(): JSX.Element {
  // Important things to note. The width is 16. The width of the icons is 12.
  // This leaves a padding of 2 on each side, which isn't even needed with items-center? Not sure which
  // would be better.
  // py-2 pushes the icons down the same distance they are from the side.
  // gap-2 makes sure they are evenly spaced.
  // TODO: Need to show an active state on the icon somehow. This might need more padding.

  const [activeOne, setActiveOne] = useState('ins_15');

  return (
    <div className='hidden lg:visible w-16 h-full bg-zinc-900 lg:flex items-center lg:py-4 gap-4 lg:flex-col border-r-zinc-800 flex-none border border-transparent'>
      <div className='h-10 w-10'>
        <img src={ Logo } className="w-full" />
      </div>
      <hr className='w-1/2 border-0 border-b-[thin] border-zinc-600' />
      <div className='h-full w-full flex items-center flex-col overflow-y-auto'>
        <LinkItem instituionId='ins_15' active={ activeOne === 'ins_15' } onClick={ () => setActiveOne('ins_15') } />
        <LinkItem instituionId='ins_116794' active={ activeOne === 'ins_116794' } onClick={ () => setActiveOne('ins_116794') }  />
        <LinkItem instituionId='ins_127990' active={ activeOne === 'ins_127990' } onClick={ () => setActiveOne('ins_127990') }  />
        <LinkItem instituionId='ins_3' active={ activeOne === 'ins_3' }  onClick={ () => setActiveOne('ins_3') } />
      </div>
      <Settings className='hover:text-zinc-50 text-zinc-400 cursor-pointer' />
      <Logout className='hover:text-zinc-50 text-zinc-400 cursor-pointer' />
    </div>
  );
}

function BudgetingSideBar(): JSX.Element {
  return (
    <div className='hidden lg:w-72 h-full bg-zinc-900 flex-none lg:flex flex-col border-r-zinc-800 border border-transparent items-center'>
      <div className='w-full hover:bg-zinc-700/50 text-zinc-50 border-b-zinc-900 border-transparent border-[1px] h-12 flex items-center p-2'>
        <span className='text-zinc-50 font-semibold text-ellipsis whitespace-nowrap overflow-hidden shadow-'>
          Navy Federal Credit Union
        </span>
        <MoreVert className='ml-auto' />
      </div>
      <hr className='w-1/2 border-0 border-b-[thin] border-zinc-600' />
      <div className='h-full flex flex-col gap-4 px-2 py-4 w-full items-center'>
        <div className='w-full'>
          <span className='cursor-pointer hover:bg-zinc-700 hover:text-zinc-50 text-zinc-400 text-lg flex items-center font-semibold gap-2 p-1 align-middle rounded-md'>
            <AccountBalanceOutlined />
            Checking
            <span className='ml-auto text-xs bg-purple-500 text-zinc-50 rounded-sm py-0.5 px-1'>
              4567
            </span>
            <KeyboardArrowDown />
          </span>
        </div>
        <hr className='w-1/2 border-0 border-b-[thin] border-zinc-600' />

        <div className='w-full flex items-center flex-col gap-2 px-2'>
          <div className='w-full flex justify-between'>
            <span className='flex gap-2 items-center text-zinc-200 text-lg font-semibold'>
              <AccountBalanceWalletOutlined />
              Free-To-Use:
            </span>
            &nbsp;
            <span className='text-zinc-200 text-lg font-semibold'>
               $154.65
            </span>
          </div>
          <div className='w-full flex justify-between'>
            <span className='flex gap-2 items-center text-zinc-200 text-lg font-semibold'>
              <LocalAtmOutlined />
              Available:
            </span>
            &nbsp;
            <span className='text-zinc-200 text-lg font-semibold'>
              $4,241.30
            </span>
          </div>
          <div className='w-full flex justify-between'>
            <span className='flex gap-2 items-center text-zinc-200 text-lg font-semibold'>
              <TollOutlined />
              Current:
            </span>
            &nbsp;
            <span className='text-zinc-200 text-lg font-semibold'>
              $4,241.30
            </span>
          </div>
        </div>
        <hr className='w-1/2 border-0 border-b-[thin] border-zinc-600' />

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
            <span className='ml-auto text-sm bg-purple-500 text-zinc-50 rounded-md py-0.5 px-1.5'>
              $1,554.43
            </span>
          </NavigationItem>
          <NavigationItem>
            <SavingsOutlined />
            Goals
            <span className='ml-auto text-sm bg-purple-500 text-zinc-50 rounded-md py-0.5 px-1.5'>
              $2,549.43
            </span>
          </NavigationItem>
          <NavigationItem>
            <TodayOutlined />
            Funding Schedules
            <span className='ml-auto text-sm bg-purple-500 text-zinc-50 rounded-md py-0.5 px-1.5'>
              7/15
            </span>
          </NavigationItem>
        </div>
      </div>
    </div>
  );
}

interface LinkItemProps {
  instituionId: string;
  active?: boolean;
  onClick: () => void;
}

function LinkItem(props: LinkItemProps): JSX.Element {
  const { result: institution } = useInstitution(props.instituionId);

  const InstitutionLogo = () => {
    if (!institution?.logo) return <AccountBalance color='info' />;

    return (
      <img
        src={ `data:image/png;base64,${institution.logo}` }
      />
    );
  };

  const classes = clsx(
    'absolute',
    'bg-zinc-300',
    'right-0',
    'rounded-l-xl',
    'transition-transform',
    'w-1.5',
    {
      'h-8': props.active,
      'scale-y-100': props.active,
    },
    {
      'h-4': !props.active,
      'group-hover:scale-y-100': !props.active,
      'group-hover:scale-x-100': !props.active,
      'scale-x-0': !props.active,
      'scale-y-50': !props.active,
    },
  );

  return (
    <div className='w-full h-12 flex items-center justify-center relative group' onClick={ props.onClick }>
      <div className={ classes } />
      <div className='cursor-pointer absolute rounded-full w-10 h-10 bg-zinc-800 drop-shadow-md flex justify-center items-center'>
        <InstitutionLogo />
      </div>
    </div>
  );
}

interface TransactionDateHeaderProps {
  date: string;
}

function TransactionDateHeader(props: TransactionDateHeaderProps): JSX.Element {
  return (
    <li className='sticky top-0 z-10 h-10 flex items-center backdrop-blur-sm bg-gradient-to-t from-transparent to-zinc-900 via-90%'>
      <span className='text-zinc-300 font-semibold text-base bg-inherit z-10 px-3 md:px-4'>
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
        <span className='text-zinc-50 font-bold text-sm md:text-base text-ellipsis whitespace-nowrap overflow-hidden min-w-0'>
          {props.from || 'Free-To-Use'}
        </span>
      );
    }

    return (
      <span className='text-zinc-50/75 font-medium text-sm md:text-base text-ellipsis whitespace-nowrap overflow-hidden min-w-0'>
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
            <span className='hidden md:block text-zinc-200 font-medium text-sm w-full overflow-hidden text-ellipsis whitespace-nowrap min-w-0'>
              {props.category}
            </span>
            <span className='flex md:hidden text-sm w-full overflow-hidden text-ellipsis whitespace-nowrap min-w-0'>
              <span className='flex-none text-zinc-50/75 font-medium text-sm text-ellipsis whitespace-nowrap overflow-hidden min-w-0'>
                Spent from
              </span>
              &nbsp;
              <SpentFrom />
            </span>
          </div>
        </div>
        <div className='hidden md:flex w-1/2 overflow-hidden flex-1 min-w-0 items-center'>
          <span className='flex-none text-zinc-50/75 font-medium text-base text-ellipsis whitespace-nowrap overflow-hidden min-w-0'>
            Spent from
          </span>
          &nbsp;
          <SpentFrom />
        </div>
        <div className='flex md:min-w-[8em] shrink-0 justify-end gap-2 items-center'>
          <span className='text-end text-red-500 font-semibold'>
            {props.amount}
          </span>
          <KeyboardArrowRight className='text-zinc-600 group-hover:text-zinc-50 flex-none md:cursor-pointer' />
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
        <div className='bg-white flex items-center justify-center h-10 w-10 rounded-full'>
          <div style={ styles } />
        </div>
      );
    }

    // If we have no icon to work with then create an avatar with the first character of the transaction name.
    const letter = props.name.toUpperCase().charAt(0);
    return (
      <Avatar className='bg-zinc-800 h-10 w-10'>
        {letter}
      </Avatar>
    );
  };
  const StyledBadge = styled(Badge)(() => ({
    '& .MuiBadge-badge': {
      backgroundColor: windTheme.tailwind.colors['blue']['500'],
      color:  windTheme.tailwind.colors['blue']['500'],
      boxShadow: `0 0 0 2px ${windTheme.tailwind.colors['zinc']['900']}`,
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
