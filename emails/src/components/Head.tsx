import type React from 'react';

export type HeadProps = React.ComponentPropsWithoutRef<'head'>;

export default function Head({ children, ...props }: HeadProps) {
  return (
    <head {...props}>
      <meta content='text/html; charset=UTF-8' httpEquiv='Content-Type' />
      <meta name='x-apple-disable-message-reformatting' />
      {children}
    </head>
  );
}
