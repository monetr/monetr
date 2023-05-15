import React from 'react';
import { Link, LinkProps } from 'react-router-dom';

import { ReactElement, TextSize } from './types';

import clsx from 'clsx';

type BaseLinkProps = LinkProps & React.RefAttributes<HTMLAnchorElement>
export interface MLinkProps extends BaseLinkProps {
  children: ReactElement;
  size?: TextSize;
}

const MLinkPropsDefaults: Omit<MLinkProps, 'children' | 'to'> = {
  size: 'md',
};

export default function MLink(props: MLinkProps): JSX.Element {
  props = {
    ...MLinkPropsDefaults,
    ...props,
  };

  const classNames = clsx(
    'font-semibold',
    'text-purple-500',
    'hover:text-purple-600',
    `text-${props.size}`,
  );

  return (
    <Link
      { ...props }
      className={ classNames }
    >
      { props.children }
    </Link>
  );
}
