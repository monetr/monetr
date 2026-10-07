import { Fragment } from 'react';

import styles from './ErrorText.module.scss';

export interface ErrorTextProps {
  error?: string;
  // The description always shows and takes up its own space. The error still goes in the spot reserved under the field
  // so it doesn't move anything around when it shows up.
  description?: string;
}

export default function ErrorText(props: ErrorTextProps): React.ReactNode {
  return (
    <Fragment>
      {props.description && <p className={styles.descriptionText}>{props.description}</p>}
      {props.error && (
        <p aria-errormessage={props.error} className={styles.errorText}>
          {props.error}
        </p>
      )}
    </Fragment>
  );
}
