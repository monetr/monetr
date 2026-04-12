import type React from 'react';
import { Section } from './Section';
import styles from './Button.module.scss';

export type ButtonProps = React.ComponentPropsWithoutRef<'a'> & {
  children: React.ReactNode;
};

export function Button({ children, target = '_blank', ...props }: ButtonProps) {
  return (
    <Section className={styles.section}>
      <a className={styles.button} target={target} {...props}>
        <p className={styles.text}>{children}</p>
      </a>
    </Section>
  );
}
