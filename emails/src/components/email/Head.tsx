import type React from 'react';

export type HeadProps = React.ComponentPropsWithoutRef<'head'>;

export function Head({ children, ...props }: HeadProps) {
  return (
    <head {...props}>
      <meta httpEquiv='Content-Type' content='text/html; charset=UTF-8' />
      <meta name='x-apple-disable-message-reformatting' />
      {children}
    </head>
  );
}
