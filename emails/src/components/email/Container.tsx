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
      // Inline styles on the table wrapper are structural — they control the
      // table layout engine and must be present as inline attributes for email
      // clients that strip <style> tags.
      style={{ maxWidth: '37.5em', borderCollapse: 'separate' }}
      width='100%'
      {...props}
    >
      <tbody>
        {/* width must be inline for table layout in email clients */}
        <tr style={{ width: '100%' }}>
          <td className={className} style={style}>
            {children}
          </td>
        </tr>
      </tbody>
    </table>
  );
}
