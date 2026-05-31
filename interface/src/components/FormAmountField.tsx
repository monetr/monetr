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
import mergeClasses from '@monetr/interface/util/mergeClasses';

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
  label: undefined,
  labelDecorator: (_: LabelDecoratorProps) => null,
  disabled: false,
};

export default function MAmountField(props: MAmountFieldProps = MAmountFieldPropsDefaults): React.JSX.Element {
  const id = useId();
  const { data: localeInfo } = useLocaleCurrency(props.currency);
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
    ...MAmountFieldPropsDefaults,
    currency: localeInfo?.currency ?? 'USD',
    ...props,
    disabled: props?.disabled || formikContext?.isSubmitting,
    error: props?.error || getFormikError(),
  };

  // localeInfo comes from a query so it can be undefined before it has loaded. Fall back to sensible defaults so the
  // formatting helpers below always receive real strings.
  const locale = localeInfo?.locale ?? 'en_US';
  const currency = props.currency ?? 'USD';
  const currencyInfo = intlNumberFormat(locale, currency);

  const { labelDecorator, ...otherProps } = props;
  const LabelDecorator = labelDecorator || (() => null);

  // If we are working with a date picker, then take the current value and transform it for the actual input.
  const value = props.name ? formikContext?.values[props.name] : undefined;

  // NumericFormat has a weird callback so we aren't using the typical onChange. Instead we are using onValueChange to
  // receive updates from the component and yeet them back up to formik.
  const onChange = useCallback(
    (values: NumberFormatValues) => {
      if (formikContext && props.name) {
        formikContext.setFieldValue(props.name, values.floatValue);
      }
    },
    [props.name, formikContext],
  );

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
        <NumericFormat
          /* These top properties might be overwritten by the ...otherProps below, this is intended. */
          disabled={formikContext?.isSubmitting}
          onBlur={formikContext?.handleBlur}
          value={value}
          {...otherProps}
          /* Properties below this point cannot be overwritten by the caller! */
          className={inputStyles.input}
          data-error={Boolean(props.error)}
          decimalScale={currencyInfo.maximumFractionDigits}
          decimalSeparator={getDecimalSeparator(locale)}
          fixedDecimalScale
          onValueChange={onChange}
          placeholder={intlNumberFormatter(locale, currency)('0')}
          prefix={getCurrencySymbolPrefixed(locale, currency)}
          renderText={intlNumberFormatter(locale, currency)}
          thousandSeparator={getNumberGroupSeparator(locale)}
        />
      </div>
      <ErrorText error={props.error} />
    </div>
  );
}
