import { useContext } from 'react';
import { CalendarSync, PiggyBank, Receipt, Repeat, ShoppingCart } from 'lucide-react';
import { Link, useLocation } from 'wouter';

import Badge, { type BadgeProps } from '@monetr/interface/components/Badge';
import Divider from '@monetr/interface/components/Divider';
import { layoutVariants } from '@monetr/interface/components/Layout';
import BalanceAvailableAmount from '@monetr/interface/components/Layout/BalanceAvailableAmount';
import BalanceCurrentAmount from '@monetr/interface/components/Layout/BalanceCurrentAmount';
import BalanceFreeToUseAmount from '@monetr/interface/components/Layout/BalanceFreeToUseAmount';
import BalanceLimitAmount from '@monetr/interface/components/Layout/BalanceLimitAmount';
import LunchFlowLastUpdatedCard from '@monetr/interface/components/Layout/LunchFlowLastUpdatedCard';
import { MobileSidebarContext } from '@monetr/interface/components/Layout/MobileSidebarContextProvider';
import PlaidBankStatusCard from '@monetr/interface/components/Layout/PlaidBankStatusCard';
import PlaidLastUpdatedCard from '@monetr/interface/components/Layout/PlaidLastUpdatedCard';
import SelectBankAccount from '@monetr/interface/components/Layout/SelectBankAccount';
import { Tooltip, TooltipContent, TooltipTrigger } from '@monetr/interface/components/Tooltip';
import Typography from '@monetr/interface/components/Typography';
import { useCurrentBalance } from '@monetr/interface/hooks/useCurrentBalance';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useNextFundingDate } from '@monetr/interface/hooks/useNextFundingDate';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import { AmountType } from '@monetr/interface/util/amounts';
import mergeClasses from '@monetr/interface/util/mergeClasses';

import BudgetingSidebarTitle from './BudgetingSidebarTitle';

import styles from './BudgetingSidebar.module.scss';

export interface BudgetingSidebarProps {
  className?: string;
}

export default function BudgetingSidebar(props: BudgetingSidebarProps): React.JSX.Element | null {
  const { data: locale } = useLocaleCurrency();
  const { data: bankAccount, isError } = useSelectedBankAccount();
  const { data: balance } = useCurrentBalance();

  if (isError) {
    return null;
  }

  return (
    <div className={mergeClasses(styles.budgetSidebarRoot, props.className)}>
      <BudgetingSidebarTitle />
      <div className={styles.content}>
        <SelectBankAccount />
        <Divider className={layoutVariants({ width: '1/2' })} />

        <div className={styles.balances}>
          <BalanceFreeToUseAmount />
          <BalanceAvailableAmount />
          <BalanceCurrentAmount />
          <BalanceLimitAmount />
        </div>
        <Divider className={layoutVariants({ width: '1/2' })} />

        <div className={styles.navList}>
          <NavigationItem to={`/bank/${bankAccount?.bankAccountId}/transactions`}>
            <ShoppingCart />
            <Typography color='inherit' ellipsis size='lg' weight='medium'>
              Transactions
            </Typography>
          </NavigationItem>
          {/* recurring is a view of transactions, so it sits under them instead of being its own top level item */}
          <NavigationItem nested to={`/bank/${bankAccount?.bankAccountId}/recurring`}>
            <Repeat />
            <Typography color='inherit' ellipsis size='lg' weight='medium'>
              Recurring
            </Typography>
          </NavigationItem>
          <NavigationItem to={`/bank/${bankAccount?.bankAccountId}/expenses`}>
            <Receipt />
            <Typography color='inherit' ellipsis size='lg' weight='medium'>
              Expenses
            </Typography>
            <NavigationBadge tooltip='The total amount currently set aside for all of your expenses.'>
              {balance ? locale?.formatAmount(balance.expenses, AmountType.Stored) : ''}
            </NavigationBadge>
          </NavigationItem>
          <NavigationItem to={`/bank/${bankAccount?.bankAccountId}/goals`}>
            <PiggyBank />
            <Typography color='inherit' ellipsis size='lg' weight='medium'>
              Goals
            </Typography>
            <NavigationBadge tooltip='The total amount currently saved towards all of your goals.'>
              {balance ? locale?.formatAmount(balance.goals, AmountType.Stored) : ''}
            </NavigationBadge>
          </NavigationItem>
          <NavigationItem to={`/bank/${bankAccount?.bankAccountId}/funding`}>
            <CalendarSync />
            <Typography color='inherit' ellipsis size='lg' weight='medium'>
              Funding Schedules
            </Typography>
            <NextFundingBadge />
          </NavigationItem>
        </div>
        <PlaidBankStatusCard />
        <PlaidLastUpdatedCard linkId={bankAccount?.linkId} />
        <LunchFlowLastUpdatedCard linkId={bankAccount?.linkId} />
      </div>
    </div>
  );
}

interface NavigationItemProps {
  children: React.ReactNode;
  to: string;
  nested?: boolean;
}

function NavigationItem(props: NavigationItemProps): React.JSX.Element {
  const [pathname] = useLocation();
  const active = pathname.endsWith(props.to.replaceAll('.', ''));
  const { setIsOpen } = useContext(MobileSidebarContext);
  // Clicking the page we're already on just closes the mobile sidebar. Skipping the navigation avoids
  // scrolling back to the top of the page.
  const onClick = (event: React.MouseEvent) => {
    if (active) {
      event.preventDefault();
      setIsOpen(false);
    }
  };

  return (
    <Link
      className={mergeClasses(styles.navItem, props.nested && styles.navItemNested)}
      data-active={active}
      onClick={onClick}
      to={props.to}
    >
      {props.children}
    </Link>
  );
}

function NextFundingBadge(): React.JSX.Element | null {
  const next = useNextFundingDate();
  if (!next) {
    return null;
  }

  return (
    <NavigationBadge tooltip='The next date that money will be set aside for your expenses and goals.'>
      {next}
    </NavigationBadge>
  );
}

interface NavigationBadgeProps extends BadgeProps {
  tooltip: string;
}

function NavigationBadge({ tooltip, ...props }: NavigationBadgeProps): React.JSX.Element {
  return (
    <Tooltip delayDuration={100}>
      <TooltipTrigger asChild>
        <Badge className={styles.badgeRight} size='sm' {...props} />
      </TooltipTrigger>
      <TooltipContent side='right'>{tooltip}</TooltipContent>
    </Tooltip>
  );
}
