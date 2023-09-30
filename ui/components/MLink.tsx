import React from 'react';
import { Link, LinkProps } from 'react-router-dom';

import { ReactElement, TextSize } from './types';

import mergeTailwind from 'util/mergeTailwind';

type BaseLinkProps = LinkProps & React.RefAttributes<HTMLAnchorElement>
export interface MLinkProps extends BaseLinkProps {
  children: ReactElement;
  color?: 'primary' | 'secondary';
  size?: TextSize;
}

const MLinkPropsDefaults: Omit<MLinkProps, 'children' | 'to'> = {
  size: 'md',
  color: 'primary',
};

export default function MLink(props: MLinkProps): JSX.Element {
  props = {
    ...MLinkPropsDefaults,
    ...props,
  };

  const colors = {
    'primary': [
      'dark:text-dark-monetr-brand-faint',
      'dark:hover:text-dark-monetr-brand-bright',
      'text-purple-500',
      'hover:text-purple-600',
      'rounded',
      'focus:ring-1',
      'focus:ring-purple-500',
    ],
    'secondary': [
      'text-gray-400',
      'hover:text-gray-500',
    ],
  };

  const classNames = mergeTailwind(
    'font-semibold',
    ...colors[props.color],
    `text-${props.size}`,
    props.className,
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
