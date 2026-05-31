import { format } from 'date-fns';
import { ChevronRight } from 'lucide-react';
import { rrulestr } from 'rrule';
import { Link, useLocation } from 'wouter';

import { Avatar, AvatarFallback } from '@monetr/interface/components/Avatar';
import Typography from '@monetr/interface/components/Typography';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useNextFundingForecast } from '@monetr/interface/hooks/useNextFundingForecast';
import type FundingSchedule from '@monetr/interface/models/FundingSchedule';
import { AmountType } from '@monetr/interface/util/amounts';
import capitalize from '@monetr/interface/util/capitalize';

import styles from './FundingItem.module.scss';

export interface FundingItemProps {
  funding: FundingSchedule;
}

export default function FundingItem(props: FundingItemProps): JSX.Element | null {
  const { data: localeCurrency } = useLocaleCurrency();
  const { data: locale } = useLocale();
  const [, navigate] = useLocation();
  const { funding } = props;
  const contributionForecast = useNextFundingForecast(funding.fundingScheduleId);
  const rule = rrulestr(funding.ruleset);
  const letter = funding.name.toUpperCase().charAt(0) || '?';

  if (!locale || !localeCurrency) {
    return null;
  }

  const ruleDescription = capitalize(rule.toText());

  const next = funding.nextRecurrence;

  const dateFormatString = locale.formatLong.date({
    width: 'long',
  });
  // const dateFormatString = isThisYear(next) ? 'EEEE LLLL do' : 'EEEE LLLL do, yyyy';
  // TODO look into format distance.
  const nextOccurrenceString = format(next, dateFormatString, {
    locale,
  });

  function openDetails() {
    navigate(`/bank/${funding.bankAccountId}/funding/${funding.fundingScheduleId}/details`);
  }

  return (
    <li className={styles.root}>
      <Link
        className={styles.mobileLink}
        to={`/bank/${funding.bankAccountId}/funding/${funding.fundingScheduleId}/details`}
      />
      <div className={styles.inner}>
        <Avatar>
          <AvatarFallback>{letter}</AvatarFallback>
        </Avatar>
        <div className={styles.detailsColumn}>
          <Typography color='emphasis' ellipsis size='inherit' weight='semibold'>
            {funding.name}
          </Typography>
          <Typography ellipsis size='sm' weight='medium'>
            {ruleDescription}
          </Typography>
          <Typography ellipsis size='sm' weight='medium'>
            {nextOccurrenceString}
          </Typography>
        </div>
        <div className={styles.amountSection}>
          <div className={styles.amountColumn}>
            <div className={styles.amountRow}>
              <span className={styles.estimatedLabel}>Estimated Contribution</span>
              <span className={styles.estimatedLabelShort}>Est.</span>
              &nbsp;
              <span className={styles.estimatedValue}>
                {contributionForecast?.data
                  ? localeCurrency.formatAmount(contributionForecast.data, AmountType.Stored)
                  : '...'}
              </span>
            </div>
          </div>
          <ChevronRight className={styles.chevron} onClick={openDetails} />
        </div>
      </div>
    </li>
  );
}
