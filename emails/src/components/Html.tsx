import type React from 'react';

export type HtmlProps = React.ComponentPropsWithoutRef<'html'>;

export default function Html({ lang = 'en', dir = 'ltr', children, ...props }: HtmlProps) {
  return (
    <html dir={dir} lang={lang} {...props}>
      {children}
    </html>
  );
}
