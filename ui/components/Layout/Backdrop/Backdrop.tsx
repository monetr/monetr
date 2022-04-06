import { Backdrop1 } from 'assets';
import React from 'react';

interface BackdropProps {
  className?: string;
  children?: React.ReactNode;
}

export default function Backdrop(props: BackdropProps): JSX.Element {
  return (
    <div className={ props.className } style={ {
      backgroundImage: `url(${ Backdrop1 })`,
    } }>
      { props.children }
    </div>
  )
}
