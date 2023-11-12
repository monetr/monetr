import React, { useCallback } from 'react';
import { InputAttributes, NumberFormatValues, NumericFormat, NumericFormatProps } from 'react-number-format';
import { useFormikContext } from 'formik';

import MLabel, { MLabelDecorator, MLabelDecoratorProps } from './MLabel';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

type NumericField =  Omit<
  NumericFormatProps<InputAttributes>,
  'decimalScale' | 'fixedDecimalScale' | 'prefix' | 'type' | 'onChange' | 'onValueChange'
>

export interface MAmountFieldProps extends NumericField {
  label?: string;
  error?: string;
  labelDecorator?: MLabelDecorator;
}

const MAmountFieldPropsDefaults: MAmountFieldProps = {
  label: null,
  labelDecorator: ((_: MLabelDecoratorProps) => null),
  disabled: false,
};

export default function MAmountField(props: MAmountFieldProps = MAmountFieldPropsDefaults): JSX.Element {
  const formikContext = useFormikContext();
  const getFormikError = () => {
    if (!formikContext?.touched[props?.name]) return null;

    return formikContext?.errors[props?.name];
  };

  props = {
    ...MAmountFieldPropsDefaults,
    ...props,
    disabled: props?.disabled || formikContext?.isSubmitting,
    error: props?.error || getFormikError(),
  };

  const { labelDecorator, ...otherProps } = props;
  const LabelDecorator = labelDecorator || MAmountFieldPropsDefaults.labelDecorator;

  function Error() {
    if (!props.error) return null;

    return (
      <p className='text-xs font-medium text-red-500 mt-0.5'>
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

  // NumericFormat has a weird callback so we aren't using the typical onChange. Instead we are using onValueChange to
  // receive updates from the component and yeet them back up to formik.
  const onChange = useCallback((values: NumberFormatValues) => {
    if (formikContext) {
      formikContext.setFieldValue(props.name, values.floatValue);
    }
  }, [props.name, formikContext]);;

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
        <NumericFormat
          /* These top properties might be overwritten by the ...otherProps below, this is intended. */
          disabled={ formikContext?.isSubmitting }
          onBlur={ formikContext?.handleBlur }
          thousandSeparator=','
          thousandsGroupStyle='thousand'
          value={ value }
          { ...otherProps }
          /* Properties below this point cannot be overwritten by the caller! */
          className={ classNames }
          decimalScale={ 2 }
          fixedDecimalScale
          onValueChange={ onChange }
          prefix='$ '
        />
      </div>
      <Error />
    </div>
  );
}
