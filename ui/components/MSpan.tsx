import React from 'react';

import mergeTailwind from 'util/mergeTailwind';

export interface MSpanProps {
  variant?: 'normal' | 'light' | 'inherit';
  children: string | React.ReactNode | JSX.Element;
  ellipsis?: boolean;
  className?: string;
  size?: 'inherit' | 'sm' | 'md' | 'lg' | 'xl';
  weight?: 'normal' | 'medium' | 'semibold' | 'bold';
}

const MSpanPropsDefaults: Omit<MSpanProps, 'children'> = {
  variant: 'normal',
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
      'light': [
        'dark:text-dark-monetr-content-subtle',
        'text-monetr-content-subtle',
      ],
      'normal': [
        'dark:text-dark-monetr-content',
        'text-monetr-content',
      ],
      'inherit': [
        'text-inherit',
      ],
    }[props.variant],
    {
      'text-ellipsis overflow-hidden whitespace-nowrap min-w-0': props.ellipsis,
    },
    {
      'inherit': 'text-size-inherit',
      'sm': 'text-sm',
      'md': 'text-base',
      'lg': 'text-lg',
      'xl': 'text-xl',
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
    <span className={ classNames }>
      { props.children }
    </span>
  );
}
