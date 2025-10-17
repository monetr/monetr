/* eslint-disable max-len */
import React from 'react';
import { useFormikContext } from 'formik';

import type { ReactElement } from './types';
import { Checkbox } from '@monetr/interface/components/Checkbox';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export interface MCheckboxProps {
  id?: string;
  label?: ReactElement;
  description?: ReactElement;
  name?: string;
  disabled?: boolean;
  checked?: boolean;
  className?: string;
}

export default function MCheckbox(props: MCheckboxProps): JSX.Element {
  const formikContext = useFormikContext();

  function Label(): JSX.Element {
    if (!props.label) return null;

    const labelClasses = mergeTailwind('font-medium', {
      'dark:text-dark-monetr-content-emphasis': !props.disabled,
      'text-gray-900': !props.disabled,
      'text-gray-500': props.disabled,
      'cursor-pointer': !props.disabled,
    });

    return (
      <label htmlFor={props.id} className={labelClasses}>
        {props.label}
      </label>
    );
  }

  function Description(): JSX.Element {
    if (!props.description) return null;

    return <p className='text-gray-500 dark:text-dark-monetr-content'>{props.description}</p>;
  }

  props = {
    ...props,
    disabled: props?.disabled || formikContext?.isSubmitting,
    checked: props?.checked || formikContext.values[props.name],
  };

  const className = mergeTailwind('flex', 'gap-x-3', 'pb-3', props.className);

  return (
    <div className={className}>
      <div className='flex h-6 items-center'>
        <Checkbox
          id={props.id}
          name={props.name}
          disabled={props.disabled}
          checked={props.checked}
          onCheckedChange={state => formikContext?.setFieldValue(props.name, Boolean(state))}
          onBlur={formikContext?.handleBlur}
        />
      </div>
      <div className='text-sm leading-6'>
        <Label />
        <Description />
      </div>
    </div>
  );
}
