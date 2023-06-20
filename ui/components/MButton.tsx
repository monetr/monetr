import React from 'react';
import { ButtonBase, ButtonBaseProps } from '@mui/material';
import { useFormikContext } from 'formik';

import clsx from 'clsx';

export interface MButtonProps extends ButtonBaseProps {
  color?: 'primary' | 'secondary' | 'cancel';
  variant?: 'solid' | 'text';
  submitting?: boolean;
}

const MButtonPropsDefaults: MButtonProps = {
  disabled: false,
  color: 'secondary',
  variant: 'solid',
};

// MBaseButton is the button implementation without the formik hook in and overrides.
// If you need to use a monetr button without formik then this should be used instead.
export function MBaseButton(props: MButtonProps = MButtonPropsDefaults): JSX.Element {
  const { disabled, color: theme, variant: kind }: MButtonProps = {
    ...MButtonPropsDefaults,
    ...props,
  };
  const themeClasses = {
    'primary': {
      'solid': {
        'bg-purple-400': !disabled,
        'bg-purple-200': disabled,
        'hover:bg-purple-500': !disabled,
        'focus-visible:outline-purple-600': !disabled,
        'text-white': true,
      },
      'text': {
        'focus-visible:outline-purple-600': !disabled,
        'text-purple-400': !disabled,
        'text-purple-200': disabled,
      },
    },
    'secondary': {
      'solid': {
        'bg-white': !disabled,
        'hover:bg-gray-100': !disabled,
        'focus-visible:outline-purple-200': !disabled,
        'ring-1': true,
        'ring-gray-300': !disabled,
        'ring-gray-200': disabled,
        'ring-inset': true,
        'text-gray-900': !disabled,
        'text-gray-400': disabled,
      },
      'text': {
        'focus-visible:outline-purple-200': !disabled,
        'text-gray-900': !disabled,
        'text-gray-400': disabled,
      },
    },
    'cancel': {
      'solid': {
        'bg-red-500': !disabled,
        'bg-red-200': disabled,
        'hover:bg-red-600': !disabled,
        'focus-visible:outline-red-600': !disabled,
        'text-white': true,
      },
      'text': {
        'text-red-500': !disabled,
        'text-red-200': disabled,
        'focus-visible:outline-red-600': !disabled,
      },
    },
  }[theme][kind];
  const classNames = clsx(
    themeClasses,
    { 'shadow-sm': kind === 'solid' },
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
};

// MButton is a wrapper around MBaseButton but includes a formik hook in with some basic overrides.
export default function MButton(props: MButtonProps = MButtonPropsDefaults): JSX.Element {
  const formikContext = useFormikContext();
  props = {
    ...MButtonPropsDefaults,
    ...props,
    disabled: formikContext?.isSubmitting || props?.disabled,
    onSubmit: props?.onSubmit || (props.type === 'submit' ? formikContext?.submitForm : undefined),
  };
  return (
    <MBaseButton { ...props } />
  );
}
