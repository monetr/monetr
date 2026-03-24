import type React from 'react';

export type LinkProps = React.ComponentPropsWithoutRef<'a'>;

export function Link({ target = '_blank', style, ...props }: LinkProps) {
  return (
    <a
      target={target}
      style={{
        color: '#067df7',
        textDecoration: 'none',
        ...style,
      }}
      {...props}
    />
  );
}
