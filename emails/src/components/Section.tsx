import type React from 'react';

export type SectionProps = React.ComponentPropsWithoutRef<'table'> & {
  children?: React.ReactNode;
};

export default function Section({ children, style, className, ...props }: SectionProps) {
  return (
    <table align='center' border={0} cellPadding='0' cellSpacing='0' role='presentation' width='100%' {...props}>
      <tbody>
        <tr>
          <td className={className} style={style}>
            {children}
          </td>
        </tr>
      </tbody>
    </table>
  );
}
