import React from 'react';
import { useFormikContext } from 'formik';

import clsx from 'clsx';

type InputProps = React.DetailedHTMLProps<React.InputHTMLAttributes<HTMLInputElement>, HTMLInputElement>;
export interface MTextFieldProps extends InputProps {
  label?: string;
  error?: string;
  uppercase?: boolean;
  labelDecorator?: () => JSX.Element;
}

const MTextFieldPropsDefaults: Omit<MTextFieldProps, 'InputProps'> = {
  label: null,
  labelDecorator: () => null,
  disabled: false,
  uppercase: false,
};

export default function MTextField(props: MTextFieldProps = MTextFieldPropsDefaults): JSX.Element {
  const formikContext = useFormikContext();

  props = {
    ...MTextFieldPropsDefaults,
    ...props,
  };

  const labelClassNames = clsx(
    'mb-1',
    'block',
    'text-sm',
    'font-medium',
    'leading-6',
    {
      'text-gray-900': !props.disabled,
      'text-gray-400': props.disabled,
    },
  );

  const { labelDecorator, ...otherProps } = props;
  function Label() {
    if (!props.label) return null;
    const LabelDecorator = labelDecorator || MTextFieldPropsDefaults.labelDecorator;

    return (
      <div className="flex items-center justify-between">
        <label
          htmlFor={ props.id }
          className={ labelClassNames }
        >
          { props.label }
        </label>

        <LabelDecorator />
      </div>
    );
  }

  function Error() {
    if (!props.error) return null;

    return (
      <p className="text-sm font-medium text-red-500 mt-2">
        { props.error }
      </p>
    );
  }

  const classNames = clsx(
    {
      'ring-gray-300': !props.disabled && !props.error,
      'ring-red-300': !props.disabled && !!props.error,
      'ring-gray-200': props.disabled,
      'uppercase': props.uppercase,
    },
    'block',
    'border-0',
    'focus:ring-2',
    'focus:ring-inset',
    'focus:ring-purple-400',
    'placeholder:text-gray-400',
    'px-3',
    'py-1.5',
    'ring-1',
    'ring-inset',
    'rounded-lg',
    'shadow-sm',
    'sm:leading-6',
    'sm:text-sm',
    'text-gray-900',
    'w-full',
  );

  return (
    <div className={ props.className }>
      <Label />
      <div>
        <input
          value={ formikContext?.values[props.name] }
          onChange={ formikContext?.handleChange }
          onBlur={ formikContext?.handleBlur }
          disabled={ formikContext?.isSubmitting || props.disabled }
          { ...otherProps }
          className={ classNames }
        />
      </div>
      <Error />
    </div>
  );
}
