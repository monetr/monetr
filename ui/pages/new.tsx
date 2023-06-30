/* eslint-disable max-len */
import React from 'react';
import { AccountBalance, AccountBalanceOutlined, HomeOutlined, KeyboardArrowDown, KeyboardArrowRight, Logout, MoreVert, PriceCheckOutlined, SavingsOutlined, ShoppingCartOutlined, TodayOutlined } from '@mui/icons-material';
import { Avatar, Divider, List, ListItem, ListSubheader } from '@mui/material';

import { Logo } from 'assets';
import clsx from 'clsx';
import { ReactElement } from 'components/types';
import { useInstitution } from 'hooks/institutions';
import { useIconSearch } from 'hooks/useIconSearch';


export default function NewMonetr(): JSX.Element {
  return (
    <div className='w-full h-full bg-zinc-900 flex'>
      <BankSidebar />
      <div className='w-full h-full flex'>
        <BudgetingSideBar />
        <div className='w-full h-full flex-grow overflow-y-scroll px-2'>
          <List dense disablePadding className='w-full'>
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
          </List>
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
  const className = clsx({
    'bg-zinc-700': props.active,
    'text-zinc-50': props.active,
    'text-zinc-400': !props.active,
  }, [
    'align-middle',
    'flex',
    'font-medium',
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
  return (
    <div className='hidden md:visible w-16 h-full bg-zinc-900 md:flex items-center py-4 gap-4 md:flex-col flex-none border-r-zinc-800 border border-transparent'>
      <div className='h-10 w-10'>
        <img src={ Logo } className="w-full" />
      </div>
      <Divider className='border-zinc-600 w-1/2' />
      <div className='h-full flex items-center gap-2 flex-col overflow-y-auto'>
        <LinkItem instituionId='ins_15' />
        <LinkItem instituionId='ins_116794' />
        <LinkItem instituionId='ins_127990' />
        <LinkItem instituionId='ins_3' />
      </div>
      <Logout className='text-zinc-400' />
    </div>
  );
}

function BudgetingSideBar(): JSX.Element {
  return (
    <div className='w-60 h-full bg-zinc-900 flex-none flex flex-col border-r-zinc-800 border border-transparent items-center'>
      <div className='w-full hover:bg-zinc-700/50 text-zinc-50 border-b-zinc-900 border-transparent border-[1px] h-16 flex items-center p-2'>
        <span className='text-zinc-50 font-semibold text-ellipsis whitespace-nowrap overflow-hidden shadow-'>
          Navy Federal Credit Union
        </span>
        <MoreVert />
      </div>
      <Divider className='border-zinc-600 w-1/2' />
      <div className='h-full flex flex-col gap-4 px-2 py-4 w-full items-center'>
        <div className='w-full'>
          <span className='hover:bg-zinc-700 hover:text-zinc-50 text-zinc-400 text-lg flex items-center font-semibold gap-2 p-1 align-middle rounded-md'>
            <AccountBalanceOutlined />
            Checking
            <span className='ml-auto text-xs bg-purple-500 text-zinc-50 rounded-sm py-0.5 px-1'>
              4567
            </span>
            <KeyboardArrowDown />
          </span>
        </div>
        <Divider className='border-zinc-600 w-1/2' />
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
          </NavigationItem>
          <NavigationItem>
            <SavingsOutlined />
            Goals
          </NavigationItem>
          <NavigationItem>
            <TodayOutlined />
            Funding Schedules
          </NavigationItem>
        </div>
      </div>
    </div>
  );
}

interface LinkItemProps {
  instituionId: string;
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

  return (
    <div className='rounded-full w-10 h-10 bg-zinc-800 drop-shadow-md flex justify-center items-center'>
      <InstitutionLogo />
    </div>
  );
}

interface TransactionDateHeaderProps {
  date: string;
}

function TransactionDateHeader(props: TransactionDateHeaderProps): JSX.Element {
  return (
    <ListSubheader className='bg-inherit backdrop-filter backdrop-blur-sm'>
      <span className='text-zinc-300 font-semibold text-base w-full h-full bg-inherit z-10'>
        {props.date}
      </span>
    </ListSubheader>
  );
}

interface TransactionItemProps {
  name: string;
  from?: string;
  category: string;
  amount: string;
}

function TransactionItem(props: TransactionItemProps): JSX.Element {
  const SpentFrom = () => {
    if (props.from) {
      return (
        <span className='text-zinc-50 font-bold text-base text-ellipsis whitespace-nowrap overflow-hidden'>
          {props.from || 'Free-To-Use'}
        </span>
      );
    }

    return (
      <span className='text-zinc-50 font-medium text-base text-ellipsis whitespace-nowrap overflow-hidden'>
        Free-To-Use
      </span>
    );
  };

  return (
    <ListItem className='w-full flex rounded-lg hover:bg-zinc-600 gap-4 group'>
      <div className='w-5/12 flex flex-row gap-4 items-center flex-shrink'>
        <TransactionIcon name={ props.name } />
        <div className='flex flex-col overflow-hidden'>
          <span className='text-zinc-50 font-semibold text-base w-full overflow-hidden text-ellipsis whitespace-nowrap'>
            {props.name}
          </span>
          <span className='text-zinc-200 font-medium text-sm w-full overflow-hidden text-ellipsis whitespace-nowrap'>
            {props.category}
          </span>
        </div>
      </div>
      <div className='w-5/12 overflow-hidden flex-shrink'>
        <span className='text-zinc-50 font-medium text-base text-ellipsis whitespace-nowrap overflow-hidden'>
          Spent from
        </span>
        &nbsp;
        <SpentFrom />
      </div>
      <span className='flex-none w-2/12 text-end text-red-500 font-semibold'>
        {props.amount}
      </span>
      <KeyboardArrowRight className='text-zinc-600 group-hover:text-zinc-50' />
    </ListItem>
  );
}

interface TransactionIconProps {
  name: string;
}

function TransactionIcon(props: TransactionIconProps): JSX.Element {

  // Try to retrieve the icon. If the icon cannot be retrieved or icons are not currently enabled in the application
  // config then this will simply return null.
  const icon = useIconSearch(props.name);
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
      <Avatar className='bg-white flex items-center justify-center h-10 w-10'>
        <div style={ styles } />
      </Avatar>
    );
  }

  // If we have no icon to work with then create an avatar with the first character of the transaction name.
  const letter = props.name.toUpperCase().charAt(0);
  return (
    <Avatar className='bg-zinc-800 h-10 w-10'>
      {letter}
    </Avatar>
  );
}
