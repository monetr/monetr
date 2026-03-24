import type React from 'react';

export type TextProps = React.ComponentPropsWithoutRef<'p'>;

export function Text({ style, ...props }: TextProps) {
  return (
    <p
      style={{
        fontSize: '14px',
        lineHeight: '24px',
        margin: '16px 0',
        ...style,
      }}
      {...props}
    />
  );
}
