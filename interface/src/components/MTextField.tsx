import React from 'react';
import { useFormikContext } from 'formik';

import MLabel, { MLabelDecorator, MLabelDecoratorProps } from './MLabel';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

type InputProps = React.DetailedHTMLProps<React.InputHTMLAttributes<HTMLInputElement>, HTMLInputElement>;
export interface MTextFieldProps extends InputProps {
  label?: string;
  error?: string;
  uppercasetext?: boolean;
  labelDecorator?: MLabelDecorator;
}

const MTextFieldPropsDefaults: Omit<MTextFieldProps, 'InputProps'> = {
  label: null,
  labelDecorator: ((_: MLabelDecoratorProps) => null),
  disabled: false,
  uppercasetext: undefined,
};

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

  const classNames = mergeTailwind(
    {
      'dark:focus:ring-dark-monetr-brand': !props.disabled && !props.error,
      'dark:hover:ring-zinc-400': !props.disabled && !props.error,
      'dark:ring-dark-monetr-border-string': !props.disabled && !props.error,
      'dark:ring-red-500': !props.disabled && !!props.error,
      'ring-gray-300': !props.disabled && !props.error,
      'ring-red-300': !props.disabled && !!props.error,
      'uppercase': props.uppercasetext,
    },
    {
      'focus:ring-purple-400': !props.error,
      'focus:ring-red-400': props.error,
    },
    {
      'dark:bg-dark-monetr-background': !props.disabled,
      'dark:text-zinc-200': !props.disabled,
      'text-gray-900': !props.disabled,
    },
    {
      'dark:bg-dark-monetr-background-subtle': props.disabled,
      'dark:ring-dark-monetr-background-emphasis': props.disabled,
      'ring-gray-200': props.disabled,
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
    'text-sm',
    'w-full',
    'dark:caret-zinc-50',
    'min-h-[38px]',
  );

  const wrapperClassNames = mergeTailwind({
    // This will make it so the space below the input is the same when there is and isn't an error.
    'pb-[18px]': !props.error,
  }, props.className);

  // If we are working with a date picker, then take the current value and transform it for the actual input.
  const value = formikContext?.values[props.name];

  return (
    <div className={ wrapperClassNames }>
      <MLabel
        label={ props.label }
        disabled={ props.disabled }
        htmlFor={ props.id }
        required={ props.required }
      >
        <LabelDecorator name={ props.name } disabled={ props.disabled } />
      </MLabel>
      <div>
        <input
          value={ value }
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
