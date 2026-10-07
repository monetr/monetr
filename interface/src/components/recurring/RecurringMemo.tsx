import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './RecurringMemo.module.scss';

export interface RecurringMemoProps {
  memo: string;
  className?: string;
}

// RecurringMemo shows the raw text a charge came in with from the bank. It's a code element so it picks up the
// monospace font and reads like the bank statement.
export default function RecurringMemo(props: RecurringMemoProps): React.JSX.Element {
  return (
    <code className={mergeClasses(styles.recurringMemo, props.className)} title={props.memo}>
      {props.memo}
    </code>
  );
}
