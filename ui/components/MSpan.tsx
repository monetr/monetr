import React from 'react';

import mergeTailwind from 'util/mergeTailwind';

export interface MSpanStyleProps {
  color?: 'default' | 'muted' | 'subtle' | 'emphasis' | 'inherit';
  ellipsis?: boolean;
  className?: string;
  size?: 'inherit' | 'xs' | 'sm' | 'md' | 'lg' | 'xl' | '2xl';
  weight?: 'normal' | 'medium' | 'semibold' | 'bold';
  component?: React.ElementType;
}

export interface MSpanProps extends MSpanStyleProps {
  children: string | React.ReactNode | JSX.Element;
  ['data-testid']?: string;
  onClick?: () => void;
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

  const classNames = MSpanDeriveClasses(props);
  const Element = props.component ?? 'span';

  return (
    <Element className={ classNames } data-testid={ props['data-testid'] } onClick={ props.onClick }>
      {props.children}
    </Element>
  );
}

/**
 * Generates a list of class names based on the props provided. This is used to style the MSpan component but can be
 * called anywhere if you want to have another component have consistent styling to that of the MSpan.
 */
export function MSpanDeriveClasses(props: MSpanStyleProps): string {
  props = {
    ...MSpanPropsDefaults,
    ...props,
  };

  return mergeTailwind(
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
      'block text-ellipsis min-w-0 truncate': props.ellipsis,
    },
    {
      'inherit': 'text-size-inherit',
      'xs': 'text-xs',
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
    {
      'code': 'dark:bg-dark-monetr-background-subtle px-1.5 rounded-lg',
    }[props.component?.toString()],
    props.className,
  );
}
