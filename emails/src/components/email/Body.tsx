import type React from 'react';

export type BodyProps = React.HtmlHTMLAttributes<HTMLBodyElement>;

/**
 * Email-compatible body wrapper. Wraps children in a presentation table
 * because Yahoo and AOL strip all styles from the <body> element when
 * converting it to a <div>. The user's styles and className are applied
 * to the inner <td> instead so they survive this stripping.
 */
export function Body({ children, style, className, ...props }: BodyProps) {
  return (
    <body {...props}>
      <table align='center' border={0} cellPadding='0' cellSpacing='0' role='presentation' width='100%'>
        <tbody>
          <tr>
            <td className={className} style={style}>
              {children}
            </td>
          </tr>
        </tbody>
      </table>
    </body>
  );
}
