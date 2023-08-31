/* eslint-disable max-len */
import React, { Fragment } from 'react';

import BankSidebar from 'components/Layout/BankSidebar';
import BudgetingSidebar from 'components/Layout/BudgetingSidebar';
import Expenses from 'pages/expenses';
import Transactions from 'pages/transactions';

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
      <BudgetingSidebar />
      <div className='w-full h-full min-w-0 flex flex-col'>
        { props.children}
      </div>
    </Fragment>
  );
}

export function TransactionsView(): JSX.Element {
  return (
    <Transactions />
  );
}

export function ExpensesView(): JSX.Element {
  return (
    <Expenses />
  );
}

