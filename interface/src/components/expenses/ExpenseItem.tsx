import { format, isThisYear } from 'date-fns';
import { ChevronRight } from 'lucide-react';
import { Link } from 'react-router-dom';
import { rrulestr } from 'rrule';

import Badge from '@monetr/interface/components/Badge';
import MerchantIcon from '@monetr/interface/components/MerchantIcon';
import Typography from '@monetr/interface/components/Typography';
import { useFundingSchedule } from '@monetr/interface/hooks/useFundingSchedule';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import type Spending from '@monetr/interface/models/Spending';
import { AmountType } from '@monetr/interface/util/amounts';
import capitalize from '@monetr/interface/util/capitalize';

import styles from './ExpenseItem.module.scss';

export interface ExpenseItemProps {
  spending: Spending;
}

function getAmountStatus(spending: Spending): string {
  if (spending.targetAmount === spending.currentAmount) {
    return 'funded';
  }
  if (spending.targetAmount < spending.currentAmount) {
    return 'overfunded';
  }
  if (spending.isBehind) {
    return 'behind';
  }
  return 'ahead';
}

export default function ExpenseItem({ spending }: ExpenseItemProps): JSX.Element | null {
  const { data: locale } = useLocaleCurrency();
  const { data: fundingSchedule } = useFundingSchedule(spending.fundingScheduleId);

  const { ruleset, nextRecurrence } = spending;
  if (!ruleset || !nextRecurrence) {
    return null;
  }

  const rule = rrulestr(ruleset);

  const amountStatus = getAmountStatus(spending);
  const detailsPath = `/bank/${spending.bankAccountId}/expenses/${spending.spendingId}/details`;

  const dateString = isThisYear(nextRecurrence)
    ? format(nextRecurrence, 'MMM do')
    : format(nextRecurrence, 'MMM do, yyyy');

  return (
    <li className={styles.root}>
      <Link className={styles.mobileLink} to={detailsPath} />
      <div className={styles.inner}>
        <div className={styles.leftSection}>
          <MerchantIcon name={spending.name} />
          <div className={styles.nameColumn}>
            <div className={styles.nameRow}>
              <Typography color='emphasis' ellipsis weight='semibold'>
                {spending.name}
              </Typography>
              <Badge className={styles.dateBadge} size='xs'>
                {dateString}
              </Badge>
            </div>
            {/* This block only shows on desktop screens */}
            <span className={styles.ruleText}>{capitalize(rule.toText())}</span>
            {/* This block only shows on mobile screens */}
            <span className={styles.mobileContribution}>
              {locale.formatAmount(spending.nextContributionAmount, AmountType.Stored)} / {fundingSchedule?.name}
            </span>
          </div>
        </div>

        {/* This block only shows on desktop screens */}
        <div className={styles.desktopContribution}>
          <span className={styles.contributionValue}>
            {locale.formatAmount(spending.nextContributionAmount, AmountType.Stored)} / {fundingSchedule?.name}
          </span>
        </div>

        {/* This block only shows on mobile screens */}
        <div className={styles.mobileAmountSection}>
          <div className={styles.amountColumn}>
            <span className={styles.currentAmount} data-status={amountStatus}>
              {locale.formatAmount(spending.currentAmount, AmountType.Stored)}
            </span>
            <hr className={styles.divider} />
            <span className={styles.targetAmount}>{locale.formatAmount(spending.targetAmount, AmountType.Stored)}</span>
          </div>
          <ChevronRight className={styles.chevron} />
        </div>

        {/* This block only shows on desktops or larger screens */}
        <div className={styles.desktopAmountSection}>
          <div className={styles.amountColumn}>
            <div className={styles.desktopAmountRow}>
              <span className={styles.currentAmount} data-status={amountStatus}>
                {locale.formatAmount(spending.currentAmount, AmountType.Stored)}
              </span>
              &nbsp;
              <span className={styles.ofLabel}>of</span>
              &nbsp;
              <span className={styles.targetAmount}>
                {locale.formatAmount(spending.targetAmount, AmountType.Stored)}
              </span>
            </div>
          </div>
          <Link className={styles.arrowLink} tabIndex={-1} to={detailsPath}>
            <ChevronRight />
          </Link>
        </div>
      </div>
    </li>
  );
}
