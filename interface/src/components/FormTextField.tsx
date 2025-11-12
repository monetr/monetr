import type React from 'react';
import { useId } from 'react';
import { useFormikContext } from 'formik';

import ErrorText from '@monetr/interface/components/ErrorText';
import Label, { type LabelDecorator, type LabelDecoratorProps } from '@monetr/interface/components/Label';
import { Skeleton } from '@monetr/interface/components/Skeleton';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import errorTextStyles from './ErrorText.module.scss';
import inputStyles from './FormTextField.module.scss';
import selectStyles from './Select.module.scss';

type InputProps = React.DetailedHTMLProps<React.InputHTMLAttributes<HTMLInputElement>, HTMLInputElement>;

export interface FormTextFieldProps extends InputProps {
  label?: string;
  error?: string;
  uppercasetext?: boolean;
  labelDecorator?: LabelDecorator;
  isLoading?: boolean;
}

const FormTextFieldPropsDefaults: Omit<FormTextFieldProps, 'InputProps'> = {
  label: null,
  labelDecorator: (_: LabelDecoratorProps) => null,
  disabled: false,
  uppercasetext: undefined,
};

export default function FormTextField(props: FormTextFieldProps = FormTextFieldPropsDefaults): JSX.Element {
  const id = useId();
  const formikContext = useFormikContext();
  const getFormikError = () => {
    if (!formikContext?.touched[props?.name]) {
      return null;
    }

    return formikContext?.errors[props?.name];
  };

  props = {
    id,
    ...FormTextFieldPropsDefaults,
    ...props,
    disabled: props?.disabled || formikContext?.isSubmitting,
    error: props?.error || getFormikError(),
  };

  const { labelDecorator, ...otherProps } = props;
  const LabelDecorator = labelDecorator || FormTextFieldPropsDefaults.labelDecorator;

  // If we are working with a date picker, then take the current value and transform it for the actual input.
  const value = formikContext?.values[props.name];

  if (props.isLoading) {
    return (
      <div className={mergeTailwind(errorTextStyles.errorTextPadding, props.className)}>
        <Label label={props.label} disabled htmlFor={props.id} required={props.required}>
          <LabelDecorator name={props.name} disabled />
        </Label>
        <div>
          <div aria-disabled='true' className={mergeTailwind(inputStyles.input, selectStyles.selectLoading)}>
            <Skeleton className='w-full h-5 mr-2' />
          </div>
        </div>
        <ErrorText error={props.error} />
      </div>
    );
  }

  return (
    <div className={mergeTailwind(errorTextStyles.errorTextPadding, props.className)}>
      <Label label={props.label} disabled={props.disabled} htmlFor={props.id} required={props.required}>
        <LabelDecorator name={props.name} disabled={props.disabled} />
      </Label>
      <div>
        <input
          value={value}
          onChange={formikContext?.handleChange}
          onBlur={formikContext?.handleBlur}
          disabled={formikContext?.isSubmitting || props.disabled}
          tabIndex={0}
          {...otherProps}
          className={inputStyles.input}
          data-error={Boolean(props.error)}
        />
      </div>
      <ErrorText error={props.error} />
    </div>
  );
}
