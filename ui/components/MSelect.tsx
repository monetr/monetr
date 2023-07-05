import React from 'react';
import Select, { Theme } from 'react-select';

import useTheme from 'hooks/useTheme';
import mergeTailwind from 'util/mergeTailwind';

interface MSelectProps<V> extends Omit<Omit<Parameters<Select>[0], 'theme'>, 'styles'> {
  label?: string;
  error?: string;
  required?: boolean;
  disabled?: boolean;
  value?: V | undefined;
}

function LabelText(props: MSelectProps<unknown>): JSX.Element {
  if (!props.label) return null;

  const labelClassNames = mergeTailwind(
    'mb-1',
    'block',
    'text-sm',
    'font-medium',
    'leading-6',
    {
      'text-gray-900': !props.disabled,
      'text-gray-500': props.disabled,
      'dark:text-dark-monetr-content-emphasis': !props.disabled,
    },
  );

  return (
    <label
      htmlFor={ props.id }
      className={ labelClassNames }
    >
      {props.label}
    </label>
  );
}

function LabelRequired(props: MSelectProps<unknown>): JSX.Element {
  if (!props.required) return null;
  return (
    <span className='text-red-500'>
      *
    </span>
  );
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

  return (
    <div className={ wrapperClassNames }>
      <div className="flex items-center justify-between monetr-">
        <div className='flex items-center gap-0.5'>
          <LabelText { ...props } />
          <LabelRequired { ...props } />
        </div>
      </div>
      <Select
        theme={ (baseTheme: Theme): Theme => ({
          ...baseTheme,
          borderRadius: 8,
          colors: {
            ...baseTheme.colors,
            neutral0: theme.tailwind.colors['dark-monetr']['background']['DEFAULT'],
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
        { ...props }
        styles={ {
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
