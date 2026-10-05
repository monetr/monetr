import { useCallback } from 'react';
import { useFormikContext } from 'formik';

import SwitchCard from '@monetr/interface/components/SwitchCard';
import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './FormSwitch.module.scss';

export interface FormSwitchProps {
  label: string;
  description: string;
  name: string;
  disabled?: boolean;
  /**
   * checked overrides what the switch shows, for when the field should read as off even though its value is still true,
   * like a toggle that only applies once another field is filled in
   */
  checked?: boolean;
  className?: string;
  'data-testid'?: string;
}

export default function FormSwitch(props: FormSwitchProps): React.JSX.Element {
  const formikContext = useFormikContext<Record<string, unknown>>();

  const onCheckedChange = useCallback(
    (checked: boolean) => formikContext?.setFieldValue(props.name, checked),
    [formikContext, props.name],
  );

  return (
    <div className={mergeClasses(styles.formSwitchRoot, props.className)}>
      <SwitchCard
        checked={props.checked ?? Boolean(formikContext?.values[props.name])}
        data-testid={props['data-testid']}
        description={props.description}
        disabled={Boolean(props.disabled) || formikContext?.isSubmitting}
        label={props.label}
        name={props.name}
        onBlur={formikContext?.handleBlur}
        onCheckedChange={onCheckedChange}
      />
    </div>
  );
}
