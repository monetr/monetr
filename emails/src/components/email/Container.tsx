import type React from 'react';
import styles from './Container.module.scss';

export type ContainerProps = React.ComponentPropsWithoutRef<'table'> & {
  children?: React.ReactNode;
};

export function Container({ children, style, className, ...props }: ContainerProps) {
  return (
    <table
      align='center'
      border={0}
      cellPadding='0'
      cellSpacing='0'
      role='presentation'
      className={styles.table}
      width='100%'
      {...props}
    >
      <tbody>
        <tr className={styles.row}>
          <td className={className} style={style}>
            {children}
          </td>
        </tr>
      </tbody>
    </table>
  );
}
