import type React from 'react';
import { useId } from 'react';
import { useFormikContext } from 'formik';

import ErrorText from '@monetr/interface/components/ErrorText';
import Label, { type LabelDecorator, type LabelDecoratorProps } from '@monetr/interface/components/Label';
import { Skeleton } from '@monetr/interface/components/Skeleton';
import mergeClasses from '@monetr/interface/util/mergeClasses';

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
  label: undefined,
  labelDecorator: (_: LabelDecoratorProps) => null,
  disabled: false,
  uppercasetext: undefined,
};

export default function FormTextField(props: FormTextFieldProps = FormTextFieldPropsDefaults): React.JSX.Element {
  const id = useId();
  const formikContext = useFormikContext<Record<string, any>>();
  const getFormikError = (): string | undefined => {
    if (!props?.name || !formikContext?.touched[props.name]) {
      return undefined;
    }

    // These renderers are keyed by a flat field name, so the formik error for that field is a plain string.
    return formikContext?.errors[props.name] as string | undefined;
  };

  props = {
    id,
    ...FormTextFieldPropsDefaults,
    ...props,
    disabled: props?.disabled || formikContext?.isSubmitting,
    error: props?.error || getFormikError(),
  };

  const { labelDecorator, ...otherProps } = props;
  const LabelDecorator = labelDecorator || (() => null);

  // If we are working with a date picker, then take the current value and transform it for the actual input.
  const value = props.name ? formikContext?.values[props.name] : undefined;

  if (props.isLoading) {
    return (
      <div className={mergeClasses(errorTextStyles.errorTextPadding, props.className)}>
        <Label disabled htmlFor={props.id} label={props.label} required={props.required}>
          <LabelDecorator disabled name={props.name} />
        </Label>
        <div>
          <div aria-disabled='true' className={mergeClasses(inputStyles.input, selectStyles.selectLoading)}>
            <Skeleton className={selectStyles.loadingSkeleton} />
          </div>
        </div>
        <ErrorText error={props.error} />
      </div>
    );
  }

  return (
    <div className={mergeClasses(errorTextStyles.errorTextPadding, props.className)}>
      <Label disabled={props.disabled} htmlFor={props.id} label={props.label} required={props.required}>
        <LabelDecorator disabled={props.disabled} name={props.name} />
      </Label>
      <div>
        <input
          disabled={formikContext?.isSubmitting || props.disabled}
          onBlur={formikContext?.handleBlur}
          onChange={formikContext?.handleChange}
          tabIndex={0}
          value={value}
          {...otherProps}
          className={inputStyles.input}
          data-error={Boolean(props.error)}
        />
      </div>
      <ErrorText error={props.error} />
    </div>
  );
}
