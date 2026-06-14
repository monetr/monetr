import { useCallback } from 'react';
import { useFormikContext } from 'formik';

import { Checkbox, type CheckedState } from '@monetr/interface/components/Checkbox';
import Typography from '@monetr/interface/components/Typography';
import mergeClasses from '@monetr/interface/util/mergeClasses';

import formCheckboxStyles from './FormCheckbox.module.scss';
import labelStyles from './Label.module.scss';

export interface FormCheckboxProps {
  id?: string;
  label?: React.ReactNode;
  description?: React.ReactNode;
  name: string;
  disabled?: boolean;
  checked?: boolean;
  className?: string;
  'data-testid'?: string;
}

export default function FormCheckbox(props: FormCheckboxProps): React.JSX.Element {
  const formikContext = useFormikContext<Record<string, any>>();

  const onCheckedChange = useCallback(
    (state: CheckedState) => formikContext?.setFieldValue(props.name, Boolean(state)),
    [formikContext, props.name],
  );

  props = {
    ...props,
    disabled: Boolean(props?.disabled) || formikContext?.isSubmitting,
    checked: Boolean(props?.checked) || formikContext.values[props.name],
  };

  return (
    <div className={mergeClasses(formCheckboxStyles.formCheckboxRoot, props.className)}>
      <div className={formCheckboxStyles.formCheckboxWrapper}>
        <Checkbox
          checked={props.checked}
          data-testid={props['data-testid']}
          disabled={props.disabled}
          id={props.id}
          name={props.name}
          onBlur={formikContext?.handleBlur}
          onCheckedChange={onCheckedChange}
        />
      </div>
      <div>
        {props.label && (
          <label aria-disabled={props.disabled} className={labelStyles.labelText} htmlFor={props.id}>
            {props.label}
          </label>
        )}
        {props.description && (
          <Typography component='p' size='sm'>
            {props.description}
          </Typography>
        )}
      </div>
    </div>
  );
}
