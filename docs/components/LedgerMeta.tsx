import { usePage } from '@rspress/core/runtime';

import styles from './LedgerMeta.module.scss';

interface LedgerMetaProps {
  // The publish / updated date, already formatted (e.g. "2026/06/12").
  date?: string;
  // Optional override for the reading time (e.g. "4 min"). When omitted we use
  // the value rspress-plugin-reading-time computes onto pageData.readingTimeData.
  readingTime?: string;
}

// LedgerMeta is registered globally in the theme entry, so MDX pages can drop a
// <LedgerMeta date="..." /> row without importing anything. The reading time is
// computed per page by rspress-plugin-reading-time unless overridden via prop.
export default function LedgerMeta({ date, readingTime }: LedgerMetaProps): React.JSX.Element {
  const { page } = usePage();
  const data = (page as unknown as { readingTimeData?: { minutes?: number } }).readingTimeData;
  const minutes = data?.minutes;
  const computed = typeof minutes === 'number' ? `${Math.max(1, Math.ceil(minutes))} min` : undefined;
  const value = readingTime ?? computed;

  return (
    <div className={styles.row} role='presentation'>
      {date ? <span className={styles.date}>{date}</span> : null}
      <span aria-hidden='true' className={styles.leader} />
      <span className={styles.label}>reading time</span>
      {value ? <span className={styles.value}>+ {value}</span> : null}
    </div>
  );
}
