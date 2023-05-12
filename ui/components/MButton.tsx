import { ButtonBase, ButtonBaseProps } from '@mui/material';
import clsx from 'clsx';
import React from 'react';

export interface MButtonProps extends ButtonBaseProps {
  theme?: 'primary' | 'secondary' | 'cancel';
  kind?: 'solid' | 'text';
}

const MButtonPropsDefaults: MButtonProps = {
  disabled: false,
  theme: 'secondary',
  kind: 'solid',
};

export default function MButton(props: MButtonProps = MButtonPropsDefaults): JSX.Element {
  const theme = {
    'primary': {
      'solid': {
        'bg-purple-400': !props.disabled,
        'bg-purple-200': props.disabled,
        'hover:bg-purple-500': !props.disabled,
        'focus-visible:outline-purple-600': !props.disabled,
        'text-white': true,
      },
      'text': {
        'focus-visible:outline-purple-600': !props.disabled,
        'text-purple-400': !props.disabled,
        'text-purple-200': props.disabled,
      },
    },
    'secondary': {
      'solid': {
        'bg-white': !props.disabled,
        'hover:bg-gray-100': !props.disabled,
        'focus-visible:outline-purple-200': !props.disabled,
        'ring-1': true,
        'ring-gray-300': !props.disabled,
        'ring-gray-200': props.disabled,
        'ring-inset': true,
        'text-gray-900': !props.disabled,
        'text-gray-400': props.disabled,
      },
      'text': {
        'focus-visible:outline-purple-200': !props.disabled,
        'text-gray-900': !props.disabled,
        'text-gray-400': props.disabled,
      }
    },
    'cancel': {
      'solid': {
        'bg-red-500': !props.disabled,
        'bg-red-200': props.disabled,
        'hover:bg-red-600': !props.disabled,
        'focus-visible:outline-red-600': !props.disabled,
        'text-white': true,
      },
      'text': {
        'text-red-500': !props.disabled,
        'text-red-200': props.disabled,
        'focus-visible:outline-red-600': !props.disabled,
      },
    }
  }[props.theme || MButtonPropsDefaults.theme][props.kind || MButtonPropsDefaults.kind];
  const classNames = clsx(
    theme,
    { 'shadow-sm': props.kind === 'solid' },
    'focus-visible:outline',
    'focus-visible:outline-2',
    'focus-visible:outline-offset-2',
    'focus:outline-none',
    'font-semibold',
    'px-3',
    'py-1.5',
    'rounded-lg',
    'text-sm',
    'w-full',
  );

  return <ButtonBase
    { ...props }
    className={ classNames }
  />;
}
