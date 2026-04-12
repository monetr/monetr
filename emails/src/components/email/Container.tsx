import type React from 'react';

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
      // Must be inline -- email clients that strip <style> tags need these
      // for table layout to work.
      style={{ maxWidth: '37.5em', borderCollapse: 'separate' }}
      width='100%'
      {...props}
    >
      <tbody>
        {/* Must be inline for table layout in email clients */}
        <tr style={{ width: '100%' }}>
          <td className={className} style={style}>
            {children}
          </td>
        </tr>
      </tbody>
    </table>
  );
}
