import { useQuery } from '@tanstack/react-query';

import styles from './StatusBanner.module.scss';

import { Link } from '@rspress/core/theme-original';

interface GithubIssue {
  number: number;
  title: string;
  body: string | null;
  labels: Array<{ name: string }>;
}

// Upptime on monetr/status tracks incidents and scheduled maintenance as GitHub issues. Anything for monetr itself is
// labeled `monetr` while Plaid's are labeled `Plaid`, so filtering on the label leaves Plaid out of this.
export default function StatusBanner() {
  const { data } = useQuery<Array<GithubIssue>>({
    queryKey: ['https://api.github.com/repos/monetr/status/issues?state=open&labels=monetr'],
  });

  // If we get rate limited by GitHub then data will be an error object instead of an array
  if (!Array.isArray(data)) {
    return null;
  }

  const issue = data.find(isActive);
  if (!issue) {
    return null;
  }

  const maintenance = isMaintenance(issue);
  return (
    <div className={styles.root} role='status'>
      <span>
        {maintenance && 'monetr is currently undergoing maintenance, some things might not work right now.'}
        {!maintenance && 'monetr is currently experiencing issues, some things might not work right now.'}
      </span>
      <Link className={styles.link} href='https://status.monetr.app/' rel='noreferrer' target='_blank'>
        View Status
      </Link>
    </div>
  );
}

function isMaintenance(issue: GithubIssue): boolean {
  return issue.labels.some(label => label.name === 'maintenance');
}

function isActive(issue: GithubIssue): boolean {
  // Incidents stay open until they are resolved
  if (!isMaintenance(issue)) {
    return true;
  }

  // Maintenance issues can be opened ahead of time, so only show them once the window has started. Upptime keeps the
  // start time in an html comment at the top of the issue body and closes the issue itself once the window ends.
  const start = issue.body?.match(/^start: (.+)$/m)?.[1];
  if (!start) {
    return true;
  }

  return Date.parse(start) <= Date.now();
}
