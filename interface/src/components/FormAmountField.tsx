import { useCallback, useId } from 'react';
import { useFormikContext } from 'formik';
import {
  type InputAttributes,
  type NumberFormatValues,
  NumericFormat,
  type NumericFormatProps,
} from 'react-number-format';

import ErrorText from '@monetr/interface/components/ErrorText';
import Label, { type LabelDecorator, type LabelDecoratorProps } from '@monetr/interface/components/Label';
import { Skeleton } from '@monetr/interface/components/Skeleton';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import {
  getCurrencySymbolPrefixed,
  getDecimalSeparator,
  getNumberGroupSeparator,
  intlNumberFormat,
  intlNumberFormatter,
} from '@monetr/interface/util/amounts';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import errorTextStyles from './ErrorText.module.scss';
import inputStyles from './FormTextField.module.scss';
import selectStyles from './Select.module.scss';

type NumericField = Omit<
  NumericFormatProps<InputAttributes>,
  'decimalScale' | 'fixedDecimalScale' | 'prefix' | 'type' | 'onChange' | 'onValueChange'
>;

export interface MAmountFieldProps extends NumericField {
  label?: string;
  error?: string;
  labelDecorator?: LabelDecorator;
  currency?: string;
  isLoading?: boolean;
}

const MAmountFieldPropsDefaults: MAmountFieldProps = {
  label: null,
  labelDecorator: (_: LabelDecoratorProps) => null,
  disabled: false,
};

export default function MAmountField(props: MAmountFieldProps = MAmountFieldPropsDefaults): JSX.Element {
  const id = useId();
  const { data: localeInfo } = useLocaleCurrency(props.currency);
  const formikContext = useFormikContext();
  const getFormikError = () => {
    if (!formikContext?.touched[props?.name]) {
      return null;
    }

    return formikContext?.errors[props?.name];
  };

  props = {
    id,
    ...MAmountFieldPropsDefaults,
    currency: localeInfo.currency ?? 'USD',
    ...props,
    disabled: props?.disabled || formikContext?.isSubmitting,
    error: props?.error || getFormikError(),
  };
  const currencyInfo = intlNumberFormat(localeInfo.locale, props.currency);

  const { labelDecorator, ...otherProps } = props;
  const LabelDecorator = labelDecorator || MAmountFieldPropsDefaults.labelDecorator;

  // If we are working with a date picker, then take the current value and transform it for the actual input.
  const value = formikContext?.values[props.name];

  // NumericFormat has a weird callback so we aren't using the typical onChange. Instead we are using onValueChange to
  // receive updates from the component and yeet them back up to formik.
  const onChange = useCallback(
    (values: NumberFormatValues) => {
      if (formikContext) {
        formikContext.setFieldValue(props.name, values.floatValue);
      }
    },
    [props.name, formikContext],
  );

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
        <NumericFormat
          /* These top properties might be overwritten by the ...otherProps below, this is intended. */
          disabled={formikContext?.isSubmitting}
          onBlur={formikContext?.handleBlur}
          value={value}
          {...otherProps}
          /* Properties below this point cannot be overwritten by the caller! */
          className={inputStyles.input}
          fixedDecimalScale
          decimalScale={currencyInfo.maximumFractionDigits}
          decimalSeparator={getDecimalSeparator(localeInfo.locale)}
          thousandSeparator={getNumberGroupSeparator(localeInfo.locale)}
          onValueChange={onChange}
          renderText={intlNumberFormatter(localeInfo.locale, props.currency)}
          placeholder={intlNumberFormatter(localeInfo.locale, props.currency)('0')}
          prefix={getCurrencySymbolPrefixed(localeInfo.locale, props.currency)}
          data-error={Boolean(props.error)}
        />
      </div>
      <ErrorText error={props.error} />
    </div>
  );
}
