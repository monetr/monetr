/* eslint-disable max-len */
import React from 'react';
import styled from '@emotion/styled';
import { useFormikContext } from 'formik';

import { ReactElement } from './types';
import useTheme from '@monetr/interface/hooks/useTheme';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export interface MCheckboxProps {
  id?: string;
  label?: ReactElement;
  description?: ReactElement;
  name?: string;
  disabled?: boolean;
  checked?: boolean;
  onChange?: {
    /** Classic React change handler, keyed by input name */
    (e: React.ChangeEvent<any>): void;
  }
  className?: string;
}

export default function MCheckbox(props: MCheckboxProps): JSX.Element {
  const formikContext = useFormikContext();
  const theme = useTheme();

  const borderColor = theme.mediaColorSchema === 'dark' ?
    theme.tailwind.colors['dark-monetr']['border']['subtle'] :
    theme.tailwind.colors['gray']['300'];

  const Checkbox = styled('input')(() => ({
    MozAppearance: 'none',
    WebkitAppearance: 'none',
    appearance: 'none',
    padding: '0',
    WebkitPrintColorAdjust: 'exact',
    colorAdjust: 'exact',
    display: 'inline-block',
    verticalAlign: 'middle',
    backgroundOrigin: 'border-box',
    WebkitUserSelect: 'none',
    MozUserSelect: 'none',
    userSelect: 'none',
    flexShrink: '0',
    height: '1rem',
    width: '1rem',
    backgroundColor: props.disabled ? theme.tailwind.colors['gray']['500'] : ['white'],
    borderColor: borderColor,
    borderWidth: '1px',
    backgroundSize: '100% 100%',
    cursor: props.disabled ? 'default' : 'pointer',
    borderRadius: '0.25rem',
    '&:checked': {
      backgroundImage:
        'url("data:image/svg+xml;charset=utf-8,%3Csvg xmlns=\'http://www.w3.org/2000/svg\' viewBox=\'0 0 16 16\'%3E%3Cpath' +
        ' fill-rule=\'evenodd\' clip-rule=\'evenodd\' d=\'M12 5c-.28 0-.53.11-.71.29L7 9.59l-2.29-2.3a1.003 ' +
        '1.003 0 00-1.42 1.42l3 3c.18.18.43.29.71.29s.53-.11.71-.29l5-5A1.003 1.003 0 0012 5z\' fill=\'%23fff\'/%3E%3C/svg%3E")',
      backgroundColor: props.disabled ? theme.tailwind.colors['purple']['300'] : theme.tailwind.colors['purple']['500'],
    },
  }));

  function Label(): JSX.Element {
    if (!props.label) return null;

    const labelClasses = mergeTailwind(
      'font-medium',
      {
        'dark:text-dark-monetr-content-emphasis': !props.disabled,
        'text-gray-900': !props.disabled,
        'text-gray-500': props.disabled,
        'cursor-pointer': !props.disabled,
      }
    );

    return (
      <label htmlFor={ props.id } className={ labelClasses }>
        {props.label}
      </label>
    );
  }

  function Description(): JSX.Element {
    if (!props.description) return null;

    return (
      <p className="text-gray-500 dark:text-dark-monetr-content">
        { props.description }
      </p>
    );
  }

  props = {
    ...props,
    disabled: props?.disabled || formikContext?.isSubmitting,
    onChange: props?.onChange || formikContext?.handleChange,
    checked: props?.checked || formikContext.values[props.name],
  };

  const className = mergeTailwind(
    'flex',
    'gap-x-3',
    'pb-3',
    props.className,
  );

  return (
    <div className={ className }>
      <div className="flex h-6 items-center">
        <Checkbox
          id={ props.id }
          name={ props.name }
          type="checkbox"
          disabled={ props.disabled }
          checked={ props.checked }
          onChange={ props.onChange }
          onBlur={ formikContext?.handleBlur }
        />
      </div>
      <div className="text-sm leading-6">
        <Label />
        <Description />
      </div>
    </div>
  );
}
