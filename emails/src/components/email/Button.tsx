import type React from 'react';

export type ButtonProps = React.ComponentPropsWithoutRef<'a'>;

export function Button({ children, style, target = '_blank', ...props }: ButtonProps) {
  return (
    <a
      target={target}
      // Inline styles required for email client compatibility — some clients
      // strip <style> tags entirely, so these must survive as inline attributes.
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
