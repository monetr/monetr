import type React from 'react';

export type TextProps = React.ComponentPropsWithoutRef<'p'>;

export function Text({ style, ...props }: TextProps) {
  return <p style={style} {...props} />;
}
