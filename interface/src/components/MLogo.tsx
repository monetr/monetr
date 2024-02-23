import React from 'react';

import { Logo } from '@monetr/interface/assets';

type ImgProps = React.DetailedHTMLProps<React.ImgHTMLAttributes<HTMLImageElement>, HTMLImageElement>;

export interface MLogoProps extends Omit<ImgProps, 'src'> {

}

export default function MLogo(props: MLogoProps): JSX.Element {
  let logo = Logo;
  if (typeof logo === 'object') {
    logo = logo?.src;
  }

  return (
    <img
      { ...props }
      src={ logo }
    />
  );
}
