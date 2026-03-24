import type React from 'react';

export type SectionProps = React.ComponentPropsWithoutRef<'table'> & {
  children?: React.ReactNode;
};

export function Section({ children, style, className, ...props }: SectionProps) {
  return (
    <table
      align='center'
      width='100%'
      role='presentation'
      cellPadding='0'
      cellSpacing='0'
      border={0}
      {...props}
    >
      <tbody>
        <tr>
          <td className={className} style={style}>{children}</td>
        </tr>
      </tbody>
    </table>
  );
}
