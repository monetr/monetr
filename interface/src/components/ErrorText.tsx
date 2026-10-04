import { Fragment } from 'react';

import styles from './ErrorText.module.scss';

export interface ErrorTextProps {
  error?: string;
  /**
   * description is helper text for the field, it always shows and takes up its own space. the error still goes in the
   * spot reserved under the field so it doesnt move anything around when it shows up
   */
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
