import type React from 'react';

export type ButtonProps = React.ComponentPropsWithoutRef<'a'>;

export function Button({ children, style, target = '_blank', ...props }: ButtonProps) {
  return (
    <a
      target={target}
      style={{
        display: 'inline-block',
        textDecoration: 'none',
        ...style,
      }}
      {...props}
    >
      {children}
    </a>
  );
}
