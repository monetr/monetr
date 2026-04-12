import type React from 'react';
import styles from './Img.module.scss';

export type ImgProps = React.ComponentPropsWithoutRef<'img'>;

export function Img({ className, ...props }: ImgProps) {
  return <img className={className || styles.img} {...props} />;
}
