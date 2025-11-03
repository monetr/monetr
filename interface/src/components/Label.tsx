import styles from './Label.module.scss';

export interface LabelDecoratorProps {
  name?: string;
  disabled?: boolean;
}

export type LabelDecorator = React.FC<LabelDecoratorProps>;

export interface LabelProps {
  htmlFor?: string;
  label?: string;
  required?: boolean;
  disabled?: boolean;
  children?: React.ReactNode;
}

export default function Label(props: LabelProps): React.JSX.Element {
  return (
    <div className={styles.labelContainer}>
      <div className={styles.labelWrapper}>
        {props.label && (
          <label htmlFor={props.htmlFor} className={styles.labelText} aria-disabled={props.disabled}>
            {props.label}
          </label>
        )}
        {props.required && <span className={styles.labelRequiredStar}>*</span>}
      </div>
      {props.children}
    </div>
  );
}
