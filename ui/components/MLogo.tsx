import { Logo } from "assets";
import React from "react";

type ImgProps = React.DetailedHTMLProps<React.ImgHTMLAttributes<HTMLImageElement>, HTMLImageElement>;

export interface MLogoProps extends Omit<ImgProps, 'src'> {

}

export default function MLogo(props: MLogoProps): JSX.Element {
  return (
    <img
      { ...props }
      src={ Logo }
    />
  )
}
