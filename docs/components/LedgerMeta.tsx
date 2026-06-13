import { usePage } from '@rspress/core/runtime';

import styles from './LedgerMeta.module.scss';

interface LedgerMetaProps {
  // Optional override for the date (e.g. "2026/06/12"). When omitted we use the
  // page's frontmatter date, falling back to the git "last updated" time.
  date?: string;
  // Optional override for the reading time (e.g. "4 min"). When omitted we use
  // the value rspress-plugin-reading-time computes onto pageData.readingTimeData.
  readingTime?: string;
}

// LedgerMeta renders the page-metadata ledger row. It is rendered automatically
// on documentation pages from the theme entry (so no per-page MDX edits) and is
// also registered as a global MDX component for manual use / overrides. With no
// props it sources everything from the current page.
export default function LedgerMeta({ date, readingTime }: LedgerMetaProps): React.JSX.Element {
  const { page } = usePage();

  const data = (page as unknown as { readingTimeData?: { minutes?: number } }).readingTimeData;
  const minutes = data?.minutes;
  const computedReadingTime =
    typeof minutes === 'number' ? `${Math.max(1, Math.ceil(minutes))} min` : undefined;

  const frontmatterDate = typeof page.frontmatter?.date === 'string' ? page.frontmatter.date : undefined;
  const updatedDate = page.lastUpdatedTime
    ? new Date(page.lastUpdatedTime).toISOString().slice(0, 10).replace(/-/g, '/')
    : undefined;

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
