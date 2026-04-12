import type React from 'react';

export type ImgProps = React.ComponentPropsWithoutRef<'img'>;

export function Img({ style, ...props }: ImgProps) {
  return (
    <img
      // Must be inline -- some email clients add default borders and
      // underlines to images and strip <style> tags.
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
