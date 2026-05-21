import { CalendarSync, PiggyBank, Receipt, ShoppingCart } from 'lucide-react';
import { Link, useLocation } from 'wouter';

import Badge from '@monetr/interface/components/Badge';
import Divider from '@monetr/interface/components/Divider';
import { layoutVariants } from '@monetr/interface/components/Layout';
import BalanceAvailableAmount from '@monetr/interface/components/Layout/BalanceAvailableAmount';
import BalanceCurrentAmount from '@monetr/interface/components/Layout/BalanceCurrentAmount';
import BalanceFreeToUseAmount from '@monetr/interface/components/Layout/BalanceFreeToUseAmount';
import BalanceLimitAmount from '@monetr/interface/components/Layout/BalanceLimitAmount';
import PlaidBankStatusCard from '@monetr/interface/components/Layout/PlaidBankStatusCard';
import PlaidLastUpdatedCard from '@monetr/interface/components/Layout/PlaidLastUpdatedCard';
import SelectBankAccount from '@monetr/interface/components/Layout/SelectBankAccount';
import Typography from '@monetr/interface/components/Typography';
import { useCurrentBalance } from '@monetr/interface/hooks/useCurrentBalance';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useNextFundingDate } from '@monetr/interface/hooks/useNextFundingDate';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import { AmountType } from '@monetr/interface/util/amounts';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import BudgetingSidebarTitle from './BudgetingSidebarTitle';
import styles from './BudgetSidebar.module.scss';

export interface BudgetingSidebarProps {
  className?: string;
}

export default function BudgetingSidebar(props: BudgetingSidebarProps): JSX.Element {
  const { data: locale } = useLocaleCurrency();
  const { data: bankAccount, isError } = useSelectedBankAccount();
  const { data: balance } = useCurrentBalance();

  if (isError) {
    return null;
  }

  return (
    <div className={mergeTailwind(styles.budgetSidebarRoot, props.className)}>
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
          <NavigationItem to={`/bank/${bankAccount?.bankAccountId}/expenses`}>
            <Receipt />
            <Typography color='inherit' ellipsis size='lg' weight='medium'>
              Expenses
            </Typography>
            <Badge className={styles.badgeRight} size='sm'>
              {locale.formatAmount(balance?.expenses, AmountType.Stored)}
            </Badge>
          </NavigationItem>
          <NavigationItem to={`/bank/${bankAccount?.bankAccountId}/goals`}>
            <PiggyBank />
            <Typography color='inherit' ellipsis size='lg' weight='medium'>
              Goals
            </Typography>
            <Badge className={styles.badgeRight} size='sm'>
              {locale.formatAmount(balance?.goals, AmountType.Stored)}
            </Badge>
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
      </div>
    </div>
  );
}

interface NavigationItemProps {
  children: React.ReactNode;
  to: string;
}

function NavigationItem(props: NavigationItemProps): JSX.Element {
  const [pathname] = useLocation();
  const active = pathname.endsWith(props.to.replaceAll('.', ''));

  return (
    <Link className={styles.navItem} data-active={active} to={props.to}>
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
    <Badge className={styles.badgeRight} size='sm'>
      {next}
    </Badge>
  );
}
