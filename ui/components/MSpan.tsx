import React from 'react';

import mergeTailwind from 'util/mergeTailwind';

export interface MSpanProps {
  color?: 'default' | 'muted' | 'subtle' | 'emphasis' | 'inherit';
  children: string | React.ReactNode | JSX.Element;
  ellipsis?: boolean;
  className?: string;
  size?: 'inherit' | 'sm' | 'md' | 'lg' | 'xl' | '2xl';
  weight?: 'normal' | 'medium' | 'semibold' | 'bold';
  ['data-testid']?: string;
}

const MSpanPropsDefaults: Omit<MSpanProps, 'children'> = {
  color: 'default',
  size: 'inherit',
};

export default function MSpan(props: MSpanProps): JSX.Element {
  props = {
    ...MSpanPropsDefaults,
    ...props,
  };

  const classNames = mergeTailwind(
    'flex gap-2 items-center',
    {
      'default': [
        'dark:text-dark-monetr-content',
        'text-monetr-content',
      ],
      'muted': [
        'dark:text-dark-monetr-content-muted',
        'text-monetr-content-muted',
      ],
      'subtle': [
        'dark:text-dark-monetr-content-subtle',
        'text-monetr-content-subtle',
      ],
      'emphasis': [
        'dark:text-dark-monetr-content-emphasis',
        'text-monetr-content-emphasis',
      ],
      'inherit': [
        'text-inherit',
      ],
    }[props.color],
    {
      'block text-ellipsis min-w-0 truncate w-full': props.ellipsis,
    },
    {
      'inherit': 'text-size-inherit',
      'sm': 'text-sm',
      'md': 'text-base',
      'lg': 'text-lg',
      'xl': 'text-xl',
      '2xl': 'text-2xl',
    }[props.size],
    {
      'normal': 'font-normal',
      'medium': 'font-medium',
      'semibold': 'font-semibold',
      'bold': 'font-bold',
    }[props.weight],
    props.className,
  );

  return (
    <span className={ classNames } data-testid={ props['data-testid'] }>
      {props.children}
    </span>
  );
}
