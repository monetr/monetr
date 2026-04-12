import type React from 'react';

export type BodyProps = React.HtmlHTMLAttributes<HTMLBodyElement>;

// Yahoo and AOL strip styles from <body> by converting it to a <div>. Wrapping in a presentation table with styles on
// the <td> survives this.
export default function Body({ children, style, className, ...props }: BodyProps) {
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
