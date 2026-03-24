import type React from 'react';

export type ContainerProps = React.ComponentPropsWithoutRef<'table'> & {
  children?: React.ReactNode;
};

export function Container({ children, style, className, ...props }: ContainerProps) {
  return (
    <table
      align='center'
      width='100%'
      role='presentation'
      cellPadding='0'
      cellSpacing='0'
      border={0}
      style={{ maxWidth: '37.5em', borderCollapse: 'separate' }}
      {...props}
    >
      <tbody>
        <tr style={{ width: '100%' }}>
          <td className={className} style={style}>{children}</td>
        </tr>
      </tbody>
    </table>
  );
}
