import React, { Fragment, useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { cva } from 'class-variance-authority';
import { ArrowDown, ArrowUp, LoaderCircle, PanelBottomClose, PanelBottomOpen } from 'lucide-react';

import { Drawer, DrawerContent, DrawerTrigger, DrawerWrapper } from '@monetr/interface/components/Drawer';
import MLabel, { type MLabelDecorator } from '@monetr/interface/components/MLabel';
import { Skeleton } from '@monetr/interface/components/Skeleton';
import useIsMobile from '@monetr/interface/hooks/useIsMobile';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import { type UseComboboxSelectedItemChange, useCombobox } from 'downshift';

export interface SelectOption<V> {
  label: string;
  value: V;
}

export interface SelectOptionComponentProps<V = unknown> extends SelectOption<V> {
  selected: boolean;
}

export interface SelectProps<V = unknown> {
  id?: string;
  className?: string;
  name?: string;
  placeholder?: string;
  label?: string;
  labelDecorator?: MLabelDecorator;
  error?: string;
  required?: boolean;
  disabled?: boolean;
  isLoading?: boolean;
  value?: SelectOption<V>;
  options: Array<SelectOption<V>>;
  optionComponent?: React.FC<SelectOptionComponentProps<V>>;
  onChange: (newValue: SelectOption<V>) => void;
  // filterImpl is a function that takes the current filter input and returns a function that evaluates whether or not
  // the current option satisfies that filter input. This can be provided to allow for custom search implementations on
  // the options in a select. If no implementation is specified then the default behavior will be to see if the label
  // contains the provided text.
  filterImpl?: (filterText: string) => (option: SelectOption<V>) => boolean;
}

export interface SelectPropsLoading<V> extends SelectProps<V> {
  isLoading: true;
}

export function defaultFilterImplementation<V = unknown>(filterText: string): (option: SelectOption<V>) => boolean {
  return (option: SelectOption<V>) => {
    return option.label.toLocaleLowerCase().includes(filterText.toLocaleLowerCase());
  };
}

export function DefaultSelectOptionComponent<V = unknown>(props: SelectOptionComponentProps<V>): React.JSX.Element {
  return <Fragment>{props.label}</Fragment>;
}

const SelectClasses = cva(
  [
    'group',
    'block',
    'border-0',
    'focus-within:ring-2 focus-within:ring-inset',
    'placeholder:text-content-placeholder',
    'px-3 py-1.5',
    'ring-1 ring-inset',
    'rounded-lg',
    'shadow-sm',
    'sm:leading-6',
    'text-sm',
    'w-full',
    'dark:caret-zinc-50',
    'min-h-[38px]',
    // Disabled styles
    'disabled:dark:bg-background-subtle',
    'disabled:dark:ring-background-emphasis',
    'disabled:ring-gray-200',
    'disabled:text-content-disabled',
    'aria-disabled:dark:bg-background-subtle',
    'aria-disabled:dark:ring-background-emphasis',
    'aria-disabled:ring-gray-200',
    'aria-disabled:text-content-disabled',
    // Enabled styles
    'dark:bg-transparent',
    'dark:text-content',
    'text-gray-900',
    // Default ring when we are not disabled or focused
    'dark:ring-dark-monetr-border-string',
  ],
  {
    variants: {
      error: {
        true: [
          // When we are in an error state then focusing the text box should have a red ring
          'focus-within:ring-red-400',
          // If we are not disabled and we are in an error status then we should have a lighter red ring
          'dark:ring-red-500',
        ],
        // When we hover we should lighten the ring slightly
        false: [
          // When we are not focused or in an error state this is the hover ring state.
          'dark:hover:ring-zinc-400',
          // However if we are focused and not in an error state then the ring should be the primary color.
          'dark:focus-within:ring-dark-monetr-brand dark:hover:focus-within:ring-dark-monetr-brand',
        ],
      },
    },
    defaultVariants: {
      error: false,
    },
  },
);

export default function Select<V>(props: SelectProps<V>): React.JSX.Element {
  const isMobile = useIsMobile();
  if (props.isLoading) {
    return <SelectLoading<V> {...props} isLoading />;
  }

  if (isMobile) {
    return <SelectDrawer<V> {...props} />;
  }

  return <SelectCombobox<V> {...props} />;
}

export function SelectLoading<V>(props: SelectPropsLoading<V>): React.JSX.Element {
  const classNames = SelectClasses({
    error: Boolean(props.error),
  });

  const wrapperClassNames = mergeTailwind(
    {
      // This will make it so the space below the input is the same when there is and isn't an error.
      'pb-[18px]': !props.error,
    },
    props.className,
  );
  const LabelDecorator = props.labelDecorator || (() => null);

  return (
    <div className={wrapperClassNames}>
      <MLabel label={props.label} disabled={props.disabled} htmlFor={props.id} required={props.required}>
        <LabelDecorator name={props.name} disabled={props.disabled} />
      </MLabel>
      <div className={mergeTailwind(classNames, 'flex cursor-progress gap-1 items-center')}>
        <Skeleton className='w-full h-5 mr-2' />
        <SelectIndicator disabled={props.disabled} isLoading={props.isLoading} open={false} />
      </div>
      {Boolean(props.error) && <p className='text-xs font-medium text-red-500 mt-0.5'>{props.error}</p>}
    </div>
  );
}

export function SelectCombobox<V>(props: SelectProps<V>): React.JSX.Element {
  const inputWrapperRef = useRef<HTMLDivElement>(null);
  const [items, setItems] = useState<Array<SelectOption<V>>>(props.options);
  const filterImplementation = useMemo(() => {
    if (props.filterImpl) {
      return props.filterImpl;
    }

    return defaultFilterImplementation<V>;
  }, [props.filterImpl]);

  const { isOpen, getMenuProps, getInputProps, getItemProps, openMenu, selectedItem } = useCombobox({
    selectedItem: props.value,
    // By default the highest item should be "highlighted" unless the user moves the highlight themselves.
    defaultHighlightedIndex: 0,
    // But the initially highlighted item should be the one they have selected otherwise fallback to the first item.
    initialHighlightedIndex: props.value ? props.options.indexOf(props.value) : 0,
    onInputValueChange({ inputValue }) {
      setItems(props.options.filter(filterImplementation(inputValue)));
    },
    onSelectedItemChange(changes: UseComboboxSelectedItemChange<SelectOption<V>>) {
      if (changes.selectedItem) {
        props?.onChange(changes.selectedItem);
      }
    },
    items,
    itemToString(item: SelectOption<V>) {
      return item ? item.label : '';
    },
  });

  const onOpenClickHandler = useCallback(() => {
    if (props.disabled) {
      return;
    }

    openMenu();
  }, [props, openMenu]);

  useEffect(() => {
    if (!isOpen) {
      // Clear the filter when we are currently open and moving to a closed state, this makes it so that if we re-open
      // the menu it is not filtered.
      setItems(props.options);
    }
  }, [props.options, isOpen]);

  const renderStyles = useMemo(() => {
    // Controls the height of the menu that is rendered, makes sure that we dont render past the bottom of the page.
    if (isOpen) {
      const distanceFromTop = inputWrapperRef.current.offsetTop;
      const heightOfWindow = window.innerHeight;
      const heightOfWrapper = inputWrapperRef.current.offsetHeight;
      const widthOfWrapper = inputWrapperRef.current.offsetWidth;
      const bottomPadding = 8;
      const spaceBelow = heightOfWindow - distanceFromTop - heightOfWrapper - bottomPadding;
      const maxHeight = 400;
      return {
        maxHeight: `${spaceBelow < maxHeight ? spaceBelow : maxHeight}px`,
        width: `${widthOfWrapper}px`,
      };
    }

    return {};
  }, [isOpen]);

  const classNames = SelectClasses({
    error: Boolean(props.error),
  });

  const wrapperClassNames = mergeTailwind(
    {
      // This will make it so the space below the input is the same when there is and isn't an error.
      'pb-[18px]': !props.error,
    },
    props.className,
  );
  const LabelDecorator = props.labelDecorator || (() => null);

  return (
    <div className={wrapperClassNames}>
      <MLabel label={props.label} disabled={props.disabled} htmlFor={props.id} required={props.required}>
        <LabelDecorator name={props.name} disabled={props.disabled} />
      </MLabel>
      {/** biome-ignore lint/a11y/noStaticElementInteractions: Need to account for weird padding here */}
      <div
        onClick={onOpenClickHandler}
        ref={inputWrapperRef}
        className={mergeTailwind(classNames, 'flex cursor-text gap-1 items-center')}
        aria-disabled={props.disabled}
      >
        <input
          {...getInputProps({
            disabled: props.disabled,
            'aria-disabled': props.disabled,
            placeholder: props.placeholder,
            className: mergeTailwind('flex-1 bg-transparent disabled:text-gray-500'),
            onFocus: openMenu,
            spellCheck: false,
          })}
        />
        <SelectIndicator disabled={props.disabled} isLoading={props.isLoading} open={isOpen} />
      </div>
      {Boolean(props.error) && <p className='text-xs font-medium text-red-500 mt-0.5'>{props.error}</p>}
      <ul
        className={mergeTailwind(
          'absolute dark:bg-dark-monetr-background-focused rounded-lg p-1 overflow-y-auto space-y-0.5',
          {
            hidden: !(isOpen && items.length),
          },
        )}
        {...getMenuProps()}
        style={renderStyles}
      >
        {isOpen &&
          items.map((item, index) => (
            <li
              key={item.label}
              className={mergeTailwind(
                [
                  'group w-full rounded-lg px-2 py-1.5',
                  'hover:bg-zinc-600 aria-selected:bg-zinc-600',
                  'cursor-pointer disabled:cursor-not-allowed',
                ],
                {
                  // The _ACTUAL_ selected state will be slightly darker than the hover state.
                  'bg-zinc-700': selectedItem?.value === item.value,
                },
              )}
              {...getItemProps({
                item,
                index,
              })}
            >
              {React.createElement<SelectOptionComponentProps<V>>(
                props.optionComponent ?? DefaultSelectOptionComponent,
                {
                  ...item,
                  selected: selectedItem?.value === item.value,
                },
              )}
            </li>
          ))}
      </ul>
    </div>
  );
}

export function SelectDrawer<V>(props: SelectProps<V>): React.JSX.Element {
  const [open, setOpen] = useState<boolean>(false);
  const onChange = useCallback(
    (option: SelectOption<V>) => {
      if (props.onChange) {
        props.onChange(option);
        setOpen(false);
      }
    },
    [props],
  );
  const classNames = SelectClasses({
    error: Boolean(props.error),
  });

  const wrapperClassNames = mergeTailwind(
    {
      // This will make it so the space below the input is the same when there is and isn't an error.
      'pb-[18px]': !props.error,
    },
    props.className,
  );
  const LabelDecorator = props.labelDecorator || (() => null);

  return (
    <div className={wrapperClassNames}>
      <MLabel label={props.label} disabled={props.disabled} htmlFor={props.id} required={props.required}>
        <LabelDecorator name={props.name} disabled={props.disabled} />
      </MLabel>
      <Drawer open={open} onOpenChange={setOpen}>
        <DrawerTrigger asChild>
          <button
            type='button'
            className={mergeTailwind(classNames, 'flex cursor-text gap-1 items-center text-start')}
            aria-disabled={props.disabled}
          >
            <span
              aria-disabled={props.disabled}
              className={mergeTailwind('flex-1 bg-transparent disabled:text-gray-500', {
                // If we don't have a value then use the placeholder text style.
                'text-content-placeholder': !props.value?.label,
                'text-content-disabled': props.disabled,
              })}
            >
              {props.value?.label ?? props.placeholder}
            </span>
            <SelectIndicator disabled={props.disabled} isLoading={props.isLoading} open={open} />
          </button>
        </DrawerTrigger>
        <DrawerContent>
          <DrawerWrapper>
            <ul className={mergeTailwind('space-y-0.5 pl-2 pr-2')}>
              {open &&
                props.options.map(item => (
                  <li
                    key={item.label}
                    className={mergeTailwind(
                      [
                        'group w-full rounded-lg px-2 py-1.5',
                        'hover:bg-zinc-600 aria-selected:bg-zinc-600',
                        'active:bg-zinc-600 aria-selected:bg-zinc-600',
                        'cursor-pointer disabled:cursor-not-allowed',
                      ],
                      {
                        // The _ACTUAL_ selected state will be slightly darker than the hover state.
                        'bg-zinc-700': props.value === item,
                      },
                    )}
                    onClick={() => onChange(item)}
                  >
                    {React.createElement<SelectOptionComponentProps<V>>(
                      props.optionComponent ?? DefaultSelectOptionComponent,
                      {
                        ...item,
                        selected: props.value === item,
                      },
                    )}
                  </li>
                ))}
            </ul>
          </DrawerWrapper>
        </DrawerContent>
      </Drawer>
      {Boolean(props.error) && <p className='text-xs font-medium text-red-500 mt-0.5'>{props.error}</p>}
    </div>
  );
}

interface SelectIndicator {
  isLoading?: boolean;
  disabled?: boolean;
  open?: boolean;
}

const SelectIndicatorClasses = cva(
  [
    'size-5',
    // Put the dropdown icon last
    'order-last',
  ],
  {
    variants: {
      isLoading: {
        true: 'animate-spin',
        false: '',
      },
      disabled: {
        true: 'text-gray-500',
        // Should match the placeholder text color when not focused.
        false: [
          'text-gray-400',
          // When the textbox is focused it should match the text color
          // But only show hover and focus states when we are not disabled.
          'group-hover:text-zinc-200 group-focus-within:text-zinc-200',
        ],
      },
    },
    defaultVariants: {
      isLoading: false,
      disabled: false,
    },
  },
);

function SelectIndicator({ isLoading, disabled, open }: SelectIndicator): React.JSX.Element {
  const isMobile = useIsMobile();
  const className = SelectIndicatorClasses({ isLoading, disabled });
  if (isLoading) {
    return <LoaderCircle className={className} />;
  }

  if (isMobile) {
    return open ? <PanelBottomClose className={className} /> : <PanelBottomOpen className={className} />;
  }

  return open ? <ArrowDown className={className} /> : <ArrowUp className={className} />;
}
