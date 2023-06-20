import React from 'react';
import { useFormikContext } from 'formik';

import clsx from 'clsx';

type InputProps = React.DetailedHTMLProps<React.InputHTMLAttributes<HTMLInputElement>, HTMLInputElement>;
export interface MTextFieldProps extends InputProps {
  label?: string;
  error?: string;
  uppercasetext?: boolean;
  labelDecorator?: () => JSX.Element;
}

const MTextFieldPropsDefaults: Omit<MTextFieldProps, 'InputProps'> = {
  label: null,
  labelDecorator: () => null,
  disabled: false,
  uppercasetext: undefined,
};

function LabelText(props: MTextFieldProps): JSX.Element {
  if (!props.label) return null;

  const labelClassNames = clsx(
    'mb-1',
    'block',
    'text-sm',
    'font-medium',
    'leading-6',
    {
      'text-gray-900': !props.disabled,
      'text-gray-500': props.disabled,
    },
  );

  return (
    <label
      htmlFor={ props.id }
      className={ labelClassNames }
    >
      {props.label}
    </label>
  );
}

function LabelRequired(props: MTextFieldProps): JSX.Element {
  if (!props.required) return null;
  return (
    <span className='text-red-500'>
      *
    </span>
  );
}

export default function MTextField(props: MTextFieldProps = MTextFieldPropsDefaults): JSX.Element {
  const formikContext = useFormikContext();
  const getFormikError = () => {
    if (!formikContext?.touched[props?.name]) return null;

    return formikContext?.errors[props?.name];
  };

  props = {
    ...MTextFieldPropsDefaults,
    ...props,
    disabled: props?.disabled || formikContext?.isSubmitting,
    error: props?.error || getFormikError(),
  };

  const { labelDecorator, ...otherProps } = props;
  const LabelDecorator = labelDecorator || MTextFieldPropsDefaults.labelDecorator;

  function Error() {
    if (!props.error) return null;

    return (
      <p className="text-xs font-medium text-red-500 mt-0.5">
        {props.error}
      </p>
    );
  }

  const classNames = clsx(
    {
      'ring-gray-300': !props.disabled && !props.error,
      'ring-red-300': !props.disabled && !!props.error,
      'ring-gray-200': props.disabled,
      'uppercase': props.uppercasetext,
    },
    {
      'focus:ring-purple-400': !props.error,
      'focus:ring-red-400': props.error,
    },
    {
      'text-gray-900': !props.disabled,
      'text-gray-500': props.disabled,
    },
    'block',
    'border-0',
    'focus:ring-2',
    'focus:ring-inset',
    'placeholder:text-gray-400',
    'px-3',
    'py-1.5',
    'ring-1',
    'ring-inset',
    'rounded-lg',
    'shadow-sm',
    'sm:leading-6',
    'sm:text-sm',
    'w-full',
  );

  const wrapperClassNames = clsx({
    // This will make it so the space below the input is the same when there is and isn't an error.
    'pb-[18px]': !props.error,
  }, props.className);

  return (
    <div className={ wrapperClassNames }>
      <div className="flex items-center justify-between">
        <div className='flex items-center gap-0.5'>
          <LabelText { ...props } />
          <LabelRequired { ...props } />
        </div>
        <LabelDecorator />
      </div>
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
