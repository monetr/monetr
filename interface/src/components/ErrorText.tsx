import styles from './ErrorText.module.css';

export interface ErrorTextProps {
  error?: string;
}

export default function ErrorText(props: ErrorTextProps): React.JSX.Element {
  if (!props.error) {
    return null;
  }
  return (
    <p aria-errormessage={props.error} className={styles.errorText}>
      {props.error}
    </p>
  );
}
