import { Fragment } from 'react';
import { Link } from 'wouter';

import ArrowLink from '@monetr/interface/components/ArrowLink';
import Badge from '@monetr/interface/components/Badge';
import MerchantIcon from '@monetr/interface/components/MerchantIcon';
import { useFundingSchedule } from '@monetr/interface/hooks/useFundingSchedule';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import type Spending from '@monetr/interface/models/Spending';
import { AmountType } from '@monetr/interface/util/amounts';
import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './GoalItem.module.scss';

export interface GoalItemProps {
  spending: Spending;
}

export default function GoalItem({ spending }: GoalItemProps): JSX.Element | null {
  const { data: locale } = useLocaleCurrency();
  const { data: fundingSchedule } = useFundingSchedule(spending.fundingScheduleId);

  if (!locale) {
    return null;
  }

  const detailsPath = `/bank/${spending.bankAccountId}/goals/${spending.spendingId}/details`;

  // By default the contribution string should simply be the amount that will be added to this goal per funding schedule
  // it is associated with.
  let contributionString = `${locale.formatAmount(spending.nextContributionAmount, AmountType.Stored)} / ${fundingSchedule?.name}`;
  // But if the goal is no longer in progress (it is complete). Then indicate that.
  if (!spending.getGoalIsInProgress()) {
    contributionString = 'Complete';
  } else if (spending.isPaused) {
    // Or if the goal is just paused.
    contributionString = 'Paused';
  }

  return (
    <li className={styles.root}>
      <Link className={styles.mobileLink} to={detailsPath} />
      <div className={styles.inner}>
        <MerchantIcon name={spending.name} />
        <div className={styles.column}>
          <div className={styles.contentRow}>
            <div className={styles.nameSection}>
              <div className={styles.nameColumn}>
                <span className={styles.name}>
                  {spending.name}
                  <span className={styles.nameInlineDetail}>&nbsp;• {spending.getNextOccurrenceString()}</span>
                </span>
                <span className={styles.occurrenceDesktop}>{spending.getNextOccurrenceString()}</span>
                <span className={styles.contributionMobile}>{contributionString}</span>
              </div>
            </div>
            <div className={styles.contributionDesktopWrap}>
              <span className={styles.contributionDesktop}>{contributionString}</span>
            </div>
            <GoalAmount spending={spending} />
          </div>
          <GoalProgressBar spending={spending} />
        </div>
        <ArrowLink to={detailsPath} />
      </div>
    </li>
  );
}

interface GoalProps {
  spending: Spending;
}

function GoalAmount({ spending }: GoalProps): JSX.Element | null {
  const { data: locale } = useLocaleCurrency();
  const amountClass = mergeClasses(styles.amount, {
    [styles.amountComplete]: spending.targetAmount <= spending.currentAmount,
    [styles.amountInProgress]: spending.targetAmount !== spending.currentAmount,
  });

  if (!locale) {
    return null;
  }

  const currentAmountString = locale.formatAmount(spending.currentAmount, AmountType.Stored);
  const targetAmountString = locale.formatAmount(spending.targetAmount, AmountType.Stored);

  if (spending.getGoalIsInProgress()) {
    return (
      <Fragment>
        <div className={styles.amountMobile}>
          <div className={styles.amountColumn}>
            <span className={amountClass}>{currentAmountString}</span>
            <hr className={styles.amountDivider} />
            <span className={styles.targetAmount}>{targetAmountString}</span>
          </div>
        </div>
        <div className={styles.amountDesktop}>
          <div className={styles.amountColumn}>
            <div className={styles.amountRow}>
              <span className={amountClass}>{currentAmountString}</span>
              &nbsp;
              <span className={styles.ofLabel}>of</span>
              &nbsp;
              <span className={styles.targetAmount}>{targetAmountString}</span>
            </div>
          </div>
        </div>
      </Fragment>
    );
  }

  return (
    <div className={styles.badgeWrap}>
      <Badge className={styles.badge} weight='medium'>
        {locale.formatAmount(spending.currentAmount, AmountType.Stored)}
      </Badge>
    </div>
  );
}

function GoalProgressBar({ spending }: GoalProps): JSX.Element {
  const { usedAmount, currentAmount, targetAmount } = spending;
  const usedProgress = ((Math.min(usedAmount, targetAmount) / targetAmount) * 100).toFixed(0);
  const allocatedProgress = ((Math.min(currentAmount + usedAmount, targetAmount) / targetAmount) * 100).toFixed(0);
  return (
    <div className={styles.progressTrack}>
      <div className={styles.progressAllocated} style={{ width: `${allocatedProgress}%` }}></div>
      <div className={styles.progressUsed} style={{ width: `${usedProgress}%` }}></div>
    </div>
  );
}
