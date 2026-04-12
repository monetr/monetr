import type React from 'react';

import styles from './Img.module.scss';

export type ImgProps = React.ComponentPropsWithoutRef<'img'>;

export default function Img({ className, ...props }: ImgProps) {
  // biome-ignore lint/a11y/useAltText: This is used for email templates, not real frontend
  return <img className={className || styles.img} {...props} />;
}
