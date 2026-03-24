import type React from 'react';

export type HtmlProps = React.ComponentPropsWithoutRef<'html'>;

export function Html({ lang = 'en', dir = 'ltr', children, ...props }: HtmlProps) {
  return (
    <html lang={lang} dir={dir} {...props}>
      {children}
    </html>
  );
}
