import React from 'react';

import { TextSize } from './types';

import clsx from 'clsx';

export interface MSpanProps {
  variant?: 'normal' | 'light';
  children: string | React.ReactNode | JSX.Element;
  size?: TextSize;
}

const MSpanPropsDefaults: Omit<MSpanProps, 'children'> = {
  variant: 'normal',
  size: 'md',
};

export default function MSpan(props: MSpanProps): JSX.Element {
  props = {
    ...MSpanPropsDefaults,
    ...props,
  };

  const classNames = clsx(
    {
      'text-gray-900': props.variant === 'normal',
      'text-gray-500': props.variant === 'light',
    },
    `text-${props.size}`,
  );

  return (
    <span className={ classNames }>
      { props.children }
    </span>
  );

}
