import type React from 'react';

type HeadingAs = 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6';

export type HeadingProps = React.ComponentPropsWithoutRef<'h1'> & {
  as?: HeadingAs;
};

export function Heading({ as: Tag = 'h1', style, ...props }: HeadingProps) {
  return <Tag style={style} {...props} />;
}
