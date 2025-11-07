import { CalendarSync, Infinity as InfinityIcon, PiggyBank, Receipt, ShoppingCart } from 'lucide-react';
import { Link, useLocation } from 'react-router-dom';

import Badge from '@monetr/interface/components/Badge';
import BalanceAvailableAmount from '@monetr/interface/components/Layout/BalanceAvailableAmount';
import BalanceCurrentAmount from '@monetr/interface/components/Layout/BalanceCurrentAmount';
import BalanceFreeToUseAmount from '@monetr/interface/components/Layout/BalanceFreeToUseAmount';
import PlaidBankStatusCard from '@monetr/interface/components/Layout/PlaidBankStatusCard';
import PlaidLastUpdatedCard from '@monetr/interface/components/Layout/PlaidLastUpdatedCard';
import SelectBankAccount from '@monetr/interface/components/Layout/SelectBankAccount';
import MDivider from '@monetr/interface/components/MDivider';
import MSpan from '@monetr/interface/components/MSpan';
import { useCurrentBalance } from '@monetr/interface/hooks/useCurrentBalance';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useNextFundingDate } from '@monetr/interface/hooks/useNextFundingDate';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import { AmountType } from '@monetr/interface/util/amounts';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import BudgetingSidebarTitle from './BudgetingSidebarTitle';

export interface BudgetingSidebarProps {
  className?: string;
}

export default function BudgetingSidebar(props: BudgetingSidebarProps): JSX.Element {
  const { data: locale } = useLocaleCurrency();
  const { data: bankAccount, isError } = useSelectedBankAccount();
  const { data: balance } = useCurrentBalance();

  const className = mergeTailwind(
    'w-72 h-full flex-none flex flex-col dark:border-r-dark-monetr-border border border-transparent items-center pb-6 lg:pb-4 overflow-auto',
    props.className,
  );

  if (isError) {
    return null;
  }

  function Limit(): JSX.Element {
    switch (bankAccount?.accountSubType) {
      case 'credit card':
        return (
          <div className='flex w-full justify-between'>
            <MSpan size='lg' weight='semibold' className='dark:text-dark-monetr-content-emphasis'>
              <InfinityIcon />
              Limit:
            </MSpan>
            <MSpan size='lg' weight='semibold' className='dark:text-dark-monetr-content-emphasis'>
              {locale.formatAmount(balance?.limit, AmountType.Stored)}
            </MSpan>
          </div>
        );
    }

    return null;
  }

  return (
    <div className={className}>
      <BudgetingSidebarTitle />
      <div className='flex h-full w-full flex-col items-center gap-4 px-2 pt-4'>
        <SelectBankAccount />
        <MDivider className='w-1/2' />

        <div className='flex w-full flex-col items-center gap-2 px-2'>
          <BalanceFreeToUseAmount />
          <BalanceAvailableAmount />
          <BalanceCurrentAmount />
          <Limit />
        </div>
        <MDivider className='w-1/2' />

        <div className='flex h-full w-full flex-col gap-2 pb-4'>
          <NavigationItem to={`/bank/${bankAccount?.bankAccountId}/transactions`}>
            <ShoppingCart />
            <MSpan ellipsis color='inherit'>
              Transactions
            </MSpan>
          </NavigationItem>
          <NavigationItem to={`/bank/${bankAccount?.bankAccountId}/expenses`}>
            <Receipt />
            <MSpan ellipsis color='inherit'>
              Expenses
            </MSpan>
            <Badge className='ml-auto' size='sm'>
              {locale.formatAmount(balance?.expenses, AmountType.Stored)}
            </Badge>
          </NavigationItem>
          <NavigationItem to={`/bank/${bankAccount?.bankAccountId}/goals`}>
            <PiggyBank />
            <MSpan ellipsis color='inherit'>
              Goals
            </MSpan>
            <Badge className='ml-auto' size='sm'>
              {locale.formatAmount(balance?.goals, AmountType.Stored)}
            </Badge>
          </NavigationItem>
          <NavigationItem to={`/bank/${bankAccount?.bankAccountId}/funding`}>
            <CalendarSync />
            <MSpan ellipsis color='inherit'>
              Funding Schedules
            </MSpan>
            <NextFundingBadge />
          </NavigationItem>
        </div>
        <PlaidBankStatusCard />
        <PlaidLastUpdatedCard linkId={bankAccount?.linkId} />
      </div>
    </div>
  );
}

interface NavigationItemProps {
  children: React.ReactNode;
  to: string;
}

function NavigationItem(props: NavigationItemProps): JSX.Element {
  const location = useLocation();
  const active = location.pathname.endsWith(props.to.replaceAll('.', ''));

  const className = mergeTailwind(
    {
      'dark:bg-dark-monetr-background-emphasis': active,
      'dark:text-dark-monetr-content-emphasis': active,
      'dark:text-dark-monetr-content-subtle': !active,
      'font-semibold': active,
      'font-medium': !active,
    },
    [
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
    ],
  );

  return (
    <Link className={className} to={props.to}>
      {props.children}
    </Link>
  );
}

function NextFundingBadge(): JSX.Element {
  const next = useNextFundingDate();
  if (!next) {
    return null;
  }

  return (
    <Badge className='ml-auto' size='sm'>
      {next}
    </Badge>
  );
}
