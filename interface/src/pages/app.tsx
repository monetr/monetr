/* eslint-disable max-len */
import { Fragment } from 'react';

import BankSidebar from '@monetr/interface/components/Layout/BankSidebar';
import BudgetingSidebar from '@monetr/interface/components/Layout/BudgetingSidebar';
import Expenses from '@monetr/interface/pages/expenses';
import Transactions from '@monetr/interface/pages/transactions';

interface MonetrWrapperProps {
  children: React.ReactNode;
}

export default function MonetrWrapper(props: MonetrWrapperProps): JSX.Element {
  return (
    <div className='w-full h-full bg-background flex'>
      <BankSidebar />
      <div className='w-full h-full flex min-w-0'>{props.children}</div>
    </div>
  );
}

interface BankViewProps {
  children: React.ReactNode;
}

export function BankView(props: BankViewProps): JSX.Element {
  return (
    <Fragment>
      <BudgetingSidebar />
      <div className='w-full h-full min-w-0 flex flex-col'>{props.children}</div>
    </Fragment>
  );
}

export function TransactionsView(): JSX.Element {
  return <Transactions />;
}

export function ExpensesView(): JSX.Element {
  return <Expenses />;
}
