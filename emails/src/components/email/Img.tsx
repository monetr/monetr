import type React from 'react';

export type ImgProps = React.ComponentPropsWithoutRef<'img'>;

export function Img({ style, ...props }: ImgProps) {
  return (
    <img
      style={{
        display: 'block',
        outline: 'none',
        border: 'none',
        textDecoration: 'none',
        ...style,
      }}
      {...props}
    />
  );
}
