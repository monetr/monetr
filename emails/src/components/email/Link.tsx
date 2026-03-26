import type React from 'react';

export type LinkProps = React.ComponentPropsWithoutRef<'a'>;

export function Link({ target = '_blank', style, ...props }: LinkProps) {
  return <a style={style} target={target} {...props} />;
}
