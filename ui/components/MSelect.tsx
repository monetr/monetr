import React, { ReactNode } from 'react';
import Select, { ActionMeta, GroupBase, OnChangeValue, OptionsOrGroups, Theme } from 'react-select';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { CheckCircleOutline } from '@mui/icons-material';
import { SwipeableDrawer } from '@mui/material';

import MLabel, { MLabelDecorator, MLabelDecoratorProps } from './MLabel';
import MSpan from './MSpan';

import useTheme from 'hooks/useTheme';
import mergeTailwind from 'util/mergeTailwind';
import { ExtractProps } from 'util/typescriptEvils';

export interface Value<T> {
  label: string;
  value: T;
}

export interface MSelectProps<V extends Value<any>> extends Omit<Parameters<Select>[0], 'theme'|'styles'|'isDisabled'> {
  label?: string;
  labelDecorator?: MLabelDecorator;
  error?: string;
  required?: boolean;
  disabled?: boolean;
  value?: V | undefined;
}

export default function MSelect<V extends Value<any> = Value<any>>(props: MSelectProps<V>): JSX.Element {
  const theme = useTheme();

  function Error() {
    if (!props.error) return null;

    return (
      <p className="text-xs font-medium text-red-500 mt-0.5">
        {props.error}
      </p>
    );
  }

  const { labelDecorator, className, ...otherProps } = props;
  const wrapperClassNames = mergeTailwind({
    // This will make it so the space below the input is the same when there is and isn't an error.
    'pb-[18px]': !props.error,
  }, className);

  const LabelDecorator = labelDecorator ?? ((_: MLabelDecoratorProps) => null);

  const classNames = mergeTailwind(
    {
      'dark:focus:ring-dark-monetr-brand': !props.disabled && !props.error,
      'dark:hover:ring-zinc-400': !props.disabled && !props.error,
      'dark:ring-dark-monetr-border-string': !props.disabled && !props.error,
      'dark:ring-red-500': !props.disabled && !!props.error,
      'ring-gray-300': !props.disabled && !props.error,
      'ring-red-300': !props.disabled && !!props.error,
    },
    {
      'focus:ring-purple-400': !props.error,
      'focus:ring-red-400': props.error,
    },
    {
      'dark:bg-dark-monetr-background': !props.disabled,
      'dark:text-zinc-200': !props.disabled,
      'text-gray-900': !props.disabled,
    },
    { // If there is not a value, the change the text of the button to be 400 for the placeholder.
      'dark:text-gray-400': !Boolean(props.value),
    },
    {
      'dark:bg-dark-monetr-background-subtle': props.disabled,
      'dark:ring-dark-monetr-background-emphasis': props.disabled,
      'ring-gray-200': props.disabled,
      'text-gray-500': props.disabled,
    },
    'block',
    'md:hidden',
    'border-0',
    'dark:caret-zinc-50',
    'focus:ring-2',
    'focus:ring-inset',
    'min-h-[38px]',
    'px-3',
    'py-1.5',
    'ring-1',
    'ring-inset',
    'rounded-lg',
    'shadow-sm',
    'sm:leading-6',
    'text-left',
    'text-sm',
    'w-full',
    'relative',
  );

  function ValueContainer(): JSX.Element {
    if (props.value?.label) {
      return (
        <span className="truncate">{ props?.value?.label }</span>
      );
    }

    return (
      <span className="truncate">{ props?.placeholder }</span>
    );
  }

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
      <button
        type='button'
        disabled={ props.disabled }
        className={ classNames }
        onClick={ () => showSelectModal({
          title: props.placeholder,
          options: props.options,
          value: props.value,
          onChange: props.onChange,
        }) }
        role='none'
      >
        <ValueContainer />
      </button>
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
        className='hidden md:block'
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

interface SelectModalProps {
  title: ReactNode;
  options: OptionsOrGroups<unknown, GroupBase<unknown>>;
  value: unknown;
  onChange: (newValue: OnChangeValue<any, false>, meta: ActionMeta<any>) => unknown;
}

function SelectModal(
  props: SelectModalProps,
): JSX.Element {
  const modal = useModal();

  const iOS = typeof navigator !== 'undefined' && /iPad|iPhone|iPod/.test(navigator.userAgent);

  const options: OptionsOrGroups<Value<unknown>, GroupBase<Value<unknown>>> =
    props.options as OptionsOrGroups<Value<unknown>, GroupBase<Value<unknown>>>;

  return (
    <SwipeableDrawer
      disableBackdropTransition={ !iOS } disableDiscovery={ iOS }
      anchor="bottom"
      open={ modal.visible }
      onClose={ modal.hide }
      onOpen={ () => modal.show() }
      className='backdrop-blur-sm backdrop-brightness-50'
    >
      <div className='h-full flex flex-col gap-4 bg-dark-monetr-background pb-8'>
        <MSpan weight='bold' size='xl' className='p-2'>
          { props.title }
        </MSpan>

        <ul className='w-full flex flex-col gap-2'>
          { options.map(item => (
            <li
              key={ item['value'] }
              className='w-full flex items-center active:bg-dark-monetr-background-subtle py-2'
              onClick={ () => {
                props.onChange(item, undefined);
                modal.hide();
              }  }
            >
              { props.value?.['value'] === item['value'] && (
                <CheckCircleOutline className='mx-2 w-6' />
              )}
              { props.value?.['value'] !== item['value'] && (
                <div className='mx-2 w-6' />
              )}
              <MSpan size='lg' weight='medium'>
                { item.label }
              </MSpan>
            </li>
          ))}
        </ul>
      </div>
    </SwipeableDrawer>
  );
}

const selectModal = NiceModal.create<SelectModalProps>(SelectModal);

function showSelectModal(props: SelectModalProps): Promise<Value<unknown>> {
  return NiceModal.show<Value<unknown>, ExtractProps<typeof selectModal>, {}>(selectModal, props);
}
