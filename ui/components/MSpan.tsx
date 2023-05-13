import clsx from "clsx";
import React from "react";

export interface MSpanProps {
  variant?: 'normal' | 'light';
  children: string | React.ReactNode | JSX.Element;
}

const MSpanPropsDefaults: Omit<MSpanProps, 'children'> = {
  variant: 'normal',
}

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
  );

  return (
    <span className={ classNames }>
      { props.children }
    </span>
  )

}
