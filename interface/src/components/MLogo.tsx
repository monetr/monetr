import React from 'react';

import Logo from '@monetr/interface/assets/Logo';

type ImgProps = React.DetailedHTMLProps<React.ImgHTMLAttributes<HTMLImageElement>, HTMLImageElement>;

export interface MLogoProps extends Omit<ImgProps, 'src'> {}

export default function MLogo(props: MLogoProps): JSX.Element {
  return <Logo {...props} />;
}
