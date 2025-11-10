import { useCallback } from 'react';
import { useFormikContext } from 'formik';

import { Checkbox, type CheckedState } from '@monetr/interface/components/Checkbox';
import Typography from '@monetr/interface/components/Typography';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

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
}

export default function FormCheckbox(props: FormCheckboxProps): React.JSX.Element {
  const formikContext = useFormikContext();

  const onCheckedChange = useCallback(
    (state: CheckedState) => formikContext?.setFieldValue(props.name, Boolean(state)),
    [formikContext, props.name],
  );

  props = {
    ...props,
    disabled: props?.disabled || formikContext?.isSubmitting,
    checked: props?.checked || formikContext.values[props.name],
  };

  return (
    <div className={mergeTailwind(formCheckboxStyles.formCheckboxRoot, props.className)}>
      <div className={formCheckboxStyles.formCheckboxWrapper}>
        <Checkbox
          id={props.id}
          name={props.name}
          disabled={props.disabled}
          checked={props.checked}
          onCheckedChange={onCheckedChange}
          onBlur={formikContext?.handleBlur}
        />
      </div>
      <div>
        {props.label && (
          <label htmlFor={props.id} className={labelStyles.labelText}>
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
