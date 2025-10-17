/* eslint-disable max-len */
import React from 'react';

import logoData from './logo.svg';

interface LogoProps {
  className?: string;
}

export default function Logo(props: LogoProps): JSX.Element {
  return <img className={props.className} src={logoData} alt='monetr' />;
}
