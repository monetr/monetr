import type React from 'react';

export type ImgProps = React.ComponentPropsWithoutRef<'img'>;

export function Img({ style, ...props }: ImgProps) {
  return (
    <img
      // Inline styles required for email client compatibility — prevents
      // unwanted borders, outlines, and link underlines that some clients
      // add to images by default. Must be inline since some clients strip
      // <style> tags.
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
