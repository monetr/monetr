import React from 'react';

import mergeTailwind from 'util/mergeTailwind';

export interface MSpanProps {
  variant?: 'normal' | 'light';
  children: string | React.ReactNode | JSX.Element;
  className?: string;
}

const MSpanPropsDefaults: Omit<MSpanProps, 'children'> = {
  variant: 'normal',
};

export default function MSpan(props: MSpanProps): JSX.Element {
  props = {
    ...MSpanPropsDefaults,
    ...props,
  };

  const classNames = mergeTailwind(
    {
      'dark:text-dark-monetr-content': props.variant === 'normal',
      'dark:text-dark-monetr-content-subtle': props.variant === 'light',
      'text-gray-900': props.variant === 'normal',
      'text-gray-500': props.variant === 'light',
    },
    'text-md',
    props.className,
  );

  return (
    <span className={ classNames }>
      { props.children }
    </span>
  );
}
