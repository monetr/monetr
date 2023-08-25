import React from 'react';
import Select, { Theme } from 'react-select';

import MLabel, { MLabelDecorator, MLabelDecoratorProps } from './MLabel';

import useTheme from 'hooks/useTheme';
import mergeTailwind from 'util/mergeTailwind';

export interface MSelectProps<V> extends Omit<Parameters<Select>[0], 'theme'|'styles'|'isDisabled'> {
  label?: string;
  labelDecorator?: MLabelDecorator;
  error?: string;
  required?: boolean;
  disabled?: boolean;
  value?: V | undefined;
}

export default function MSelect<V>(props: MSelectProps<V>): JSX.Element {
  const theme = useTheme();

  function Error() {
    if (!props.error) return null;

    return (
      <p className="text-xs font-medium text-red-500 mt-0.5">
        {props.error}
      </p>
    );
  }

  const wrapperClassNames = mergeTailwind({
    // This will make it so the space below the input is the same when there is and isn't an error.
    'pb-[18px]': !props.error,
  }, props.className);

  const { labelDecorator, ...otherProps } = props;
  const LabelDecorator = labelDecorator ?? ((_: MLabelDecoratorProps) => null);

  return (
    <div className={ wrapperClassNames }>
      <MLabel
        label={ props.label }
        htmlFor={ props.id }
        required={ props.required }
        disabled={ props.disabled }
      >
        <LabelDecorator name={ props.name } disabled={ props.disabled } />
      </MLabel>
      <Select
        theme={ (baseTheme: Theme): Theme => ({
          ...baseTheme,
          borderRadius: 8,
          colors: {
            ...baseTheme.colors,
            neutral0: theme.tailwind.colors['dark-monetr']['background']['DEFAULT'],
            neutral5: theme.tailwind.colors['dark-monetr']['background']['subtle'],
            neutral10: theme.tailwind.colors['dark-monetr']['background']['emphasis'],
            neutral20: theme.tailwind.colors['dark-monetr']['border']['string'],
            neutral30: theme.tailwind.colors['dark-monetr']['content']['DEFAULT'],
            neutral60: theme.tailwind.colors['dark-monetr']['content']['emphasis'],
            neutral70: theme.tailwind.colors['dark-monetr']['content']['emphasis'],
            neutral80: theme.tailwind.colors['dark-monetr']['content']['emphasis'],
            neutral90: theme.tailwind.colors['dark-monetr']['content']['emphasis'],
            primary25: theme.tailwind.colors['dark-monetr']['background']['emphasis'],
            primary50: theme.tailwind.colors['dark-monetr']['brand']['faint'],
            primary: theme.tailwind.colors['dark-monetr']['brand']['DEFAULT'],
          },
        }) }
        { ...otherProps }
        isDisabled={ props.disabled }
        styles={ {
          placeholder: (base: object) => ({
            ...base,
            fontSize: '0.875rem', // Equivalent to text-sm and leading-6
            lineHeight: '1.5rem',
          }),
          valueContainer: (base: object) => ({
            ...base,
            fontSize: '0.875rem', // Equivalent to text-sm and leading-6
            lineHeight: '1.5rem',
          }),
          option: (base: object) => ({
            ...base,
            color: theme.tailwind.colors['dark-monetr']['content']['emphasized'],
          }),
          menuPortal: (base: object) => ({
            ...base,
            zIndex: 9999,
          }),
        } }
      />
      <Error />
    </div>
  );
}
