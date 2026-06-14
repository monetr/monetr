import styles from './LedgerMeta.module.scss';

import { usePage } from '@rspress/core/runtime';

interface LedgerMetaProps {
  // Optional override for the date (e.g. "2026/06/12"). When omitted we use the
  // page's frontmatter date, falling back to the git "last updated" time.
  date?: string;
  // Optional override for the reading time (e.g. "4 min"). When omitted we use
  // the value rspress-plugin-reading-time computes onto pageData.readingTimeData.
  readingTime?: string;
}

// readingTimeLabel pulls the rounded minutes out of whatever
// rspress-plugin-reading-time hung off the page. The plugin sets
// pageData.readingTimeData via extendPageData but never widens rspress' page
// type, so the value reaches us as `unknown` through the page's index signature.
// Rather than cast our way through that, we narrow it. The plugin rounds up to
// whole minutes and so do we, with a floor of 1.
function readingTimeLabel(data: unknown): string | undefined {
  if (data && typeof data === 'object' && 'minutes' in data && typeof data.minutes === 'number') {
    return `${Math.max(1, Math.ceil(data.minutes))} min`;
  }

  return;
}

// formatLedgerDate turns a git commit timestamp (milliseconds) into our
// YYYY/MM/DD ledger format. We build it from the LOCAL date parts on purpose.
// toISOString renders in UTC, so an evening commit gets bumped to the next day
// and the row ends up looking a full day ahead of rspress' own "Last Updated"
// footer, which prints in local time. That mismatch is what made the date at
// the top of the page look newer than the updated-at down below.
function formatLedgerDate(timestamp: string | number): string {
  const date = new Date(timestamp);
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${date.getFullYear()}/${month}/${day}`;
}

// LedgerMeta renders the page-metadata ledger row. It is rendered automatically
// on documentation pages from the theme entry (so no per-page MDX edits) and is
// also registered as a global MDX component for manual use / overrides. With no
// props it sources everything from the current page.
export default function LedgerMeta({ date, readingTime }: LedgerMetaProps): React.JSX.Element {
  const { page } = usePage();

  const computedReadingTime = readingTimeLabel(page.readingTimeData);

  const frontmatterDate = typeof page.frontmatter?.date === 'string' ? page.frontmatter.date : undefined;
  const updatedDate = page.lastUpdatedTime ? formatLedgerDate(page.lastUpdatedTime) : undefined;

  const dateValue = date ?? frontmatterDate ?? updatedDate;
  const readingTimeValue = readingTime ?? computedReadingTime;

  return (
    <div className={styles.row} role='presentation'>
      {dateValue ? <span className={styles.date}>{dateValue}</span> : null}
      <span aria-hidden='true' className={styles.leader} />
      <span className={styles.label}>reading time</span>
      {readingTimeValue ? <span className={styles.value}>+ {readingTimeValue}</span> : null}
    </div>
  );
}
