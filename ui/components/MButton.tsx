import React from 'react';
import { ButtonBase, ButtonBaseProps } from '@mui/material';
import { useFormikContext } from 'formik';

import mergeTailwind from 'util/mergeTailwind';

export interface MButtonProps extends ButtonBaseProps {
  color?: 'primary' | 'secondary' | 'cancel';
  variant?: 'solid' | 'text' | 'outlined';
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
        'dark:bg-dark-monetr-brand': !disabled,
        'dark:hover:bg-dark-monetr-brand-subtle': !disabled,
        'bg-purple-400': !disabled,
        'bg-purple-200': disabled,
        'hover:bg-purple-500': !disabled,
        'focus-visible:outline-purple-600': !disabled,
        'text-white': true,
      },
      'text': {
        'dark:text-dark-monetr-brand-faint': !disabled,
        'focus-visible:outline-purple-600': !disabled,
        'text-purple-400': !disabled,
        'text-purple-200': disabled,
      },
    },
    'secondary': {
      'solid': {
        'bg-white': !disabled,
        'dark:bg-dark-monetr-background-subtle': !disabled,
        'dark:hover:bg-dark-monetr-background-emphasis': !disabled,
        'dark:ring-dark-monetr-border': !disabled,
        'dark:ring-dark-monetr-border-subtle': disabled,
        'dark:text-dark-monetr-content-emphasis': !disabled,
        'dark:text-dark-monetr-content-muted': disabled,
        'focus-visible:outline-purple-200': !disabled,
        'hover:bg-gray-100': !disabled,
        'ring-1': true,
        'ring-monetr-border-subtle': disabled,
        'ring-monetr-border': !disabled,
        'ring-inset': true,
        'text-gray-400': disabled,
        'text-gray-900': !disabled,
      },
      'text': {
        'dark:hover:bg-dark-monetr-background-emphasis': !disabled,
        'dark:text-dark-monetr-content-emphasis': !disabled,
        'dark:text-dark-monetr-content-muted': disabled,
        'focus-visible:outline-purple-200': !disabled,
        'text-gray-400': disabled,
        'text-gray-900': !disabled,
      },
      'outlined': {
        'dark:hover:ring-dark-monetr-brand': !disabled,
        'dark:focus:ring-dark-monetr-brand': !disabled,
        'dark:ring-dark-monetr-border-string': !disabled,
        'dark:text-gray-400': !disabled,
        'ring-1': !disabled,
        'ring-inset': !disabled,
        'focus:ring-2': !disabled,
        'focus:ring-inset': !disabled,
        'min-h-[38px]': true,
      },
    },
    'cancel': {
      'solid': {
        'dark:bg-red-600': !disabled,
        'dark:hover:bg-red-500': !disabled,
        'bg-dark-monetr-red': !disabled,
        'bg-red-200': disabled,
        'hover:bg-red-600': !disabled,
        'focus-visible:outline-red-600': !disabled,
        'text-white': true,
      },
      'text': {
        'text-dark-monetr-red': !disabled,
        'text-red-200': disabled,
        'focus-visible:outline-red-600': !disabled,
      },
    },
  }[theme][kind];
  const classNames = mergeTailwind(
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
    props.className,
  );

  return <ButtonBase
    { ...props }
    className={ classNames }
  />;
};

// MButton is a wrapper around MBaseButton but includes a formik hook in with some basic overrides.
export default function MFormButton(props: MButtonProps = MButtonPropsDefaults): JSX.Element {
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
