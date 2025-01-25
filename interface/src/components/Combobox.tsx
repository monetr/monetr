import * as React from 'react';
import { Fragment } from 'react';
import { Check, ChevronsUpDown } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '@monetr/interface/components/Command';
import { Drawer, DrawerContent, DrawerTrigger, DrawerWrapper } from '@monetr/interface/components/Drawer';
import { Popover, PopoverContent, PopoverTrigger } from '@monetr/interface/components/Popover';
import useIsMobile from '@monetr/interface/hooks/useIsMobile';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import { cva, VariantProps } from 'class-variance-authority';
 
export const comboboxVariants = cva(
  [
    'justify-between truncate',
  ],
  {
    variants: {
      variant: {
        outlined: [
          'ring-1 enabled:ring-dark-monetr-border-string',
        ],
        text: [
          'ring-0',
          // Override the background color for combobox.
          'enabled:hover:bg-transparent',

          // DROPDOWN IS CLOSED
          // When it's closed, only show the border when someone hovers over.
          'data-[state="closed"]:enabled:hover:ring-1',
          'data-[state="closed"]:enabled:hover:ring-dark-monetr-border-string',
          'data-[state="closed"]:focus:ring-0',
          // When the dropdown is closed, don't show any icons unless they are hovering.
          // '[&_svg]:data-[state="closed"]:hover:opacity-50 [&_svg]:data-[state="closed"]:opacity-0',

          // DROPDOWN IS OPEN
          // When its open, show the border all the time with the primary color.
          'data-[state="open"]:ring-dark-monetr-brand data-[state="open"]:ring-2',
          // When the dropdown is open then show icons
          // '[&_svg]:data-[state="open"]:block',
        ],
      },
      size: {
        default: '',
        select: '',
        md: '',
      },
    },
    defaultVariants: {
      variant: 'outlined',
      size: 'md',
    },
  },
);

export interface ComboboxOption<V extends string> {
  value: V;
  label: string;
  disabled?: boolean;
}

export interface ComboboxItemProps<V extends string, O extends ComboboxOption<V>> {
  currentValue: V | undefined;
  option: O;
}

export interface ComboboxProps<V extends string, O extends ComboboxOption<V>> extends 
  VariantProps<typeof comboboxVariants> {
  className?: string;
  disabled?: boolean;
  options: Array<O>;
  value?: V;
  emptyString?: string;
  placeholder?: string;
  searchPlaceholder?: string;
  showSearch?: boolean;
  onSelect?: (value: V) => void;
  components?: {
    Item?: React.ComponentType<ComboboxItemProps<V, O>>;
  }
}

export function ComboboxItem<V extends string, O extends ComboboxOption<V>>(
  props: ComboboxItemProps<V, O>,
): JSX.Element {
  return (
    <Fragment>
      <Check
        className={ mergeTailwind(
          'mr-2 h-5 w-5',
          props.currentValue == props.option.value ? 'opacity-100' : 'opacity-0'
        ) }
      />
      { props.option.label }
    </Fragment>
  );
}
 
export function Combobox<V extends string, O extends ComboboxOption<V>>(props: ComboboxProps<V, O>) {
  const { Item } = {
    Item: ComboboxItem,
    ...props.components,
  };
  const [open, setOpen] = React.useState(false);
  const isMobile = useIsMobile();

  function Picker(): JSX.Element {
    return (
      <Command>
        { ((props.options.length > 1 && props.showSearch !== false) || props.showSearch) && (
          <CommandInput placeholder={ props.searchPlaceholder } />
        ) }
        <CommandList>
          <CommandEmpty>{ props.emptyString }</CommandEmpty>
          <CommandGroup className='pb-6 md:pb-1'>
            { props.options.map(option => (
              <CommandItem
                key={ option.value }
                value={ `${option.label} ${option.value}` /* makes search work properly :( */ }
                title={ option.label }
                onSelect={ () => {
                  props.onSelect && props.onSelect(option.value);
                  setOpen(false);
                } }
              >
                <Item currentValue={ props.value } option={ option } />
              </CommandItem>
            )) }
          </CommandGroup>
        </CommandList>
      </Command>
    );
  }

  if (isMobile) {
    return (
      <Drawer open={ open } onOpenChange={ setOpen }>
        <DrawerTrigger asChild>
          <Button
            size={ props.size }
            variant={ props.variant }
            role='combobox'
            aria-expanded={ open }
            disabled={ props.disabled }
            className={ mergeTailwind(
              comboboxVariants({ variant: props.variant, size: props.size }), 
              props.className,
            ) }
          >
            <div className='text-ellipsis text-nowrap min-w-0 overflow-hidden text-inherit'>
              { props.value
                ? props.options.find(option => option.value === props.value)?.label
                : props.placeholder }
            </div>
            <ChevronsUpDown className='h-3 w-3 shrink-0 opacity-50' />
          </Button>
        </DrawerTrigger>
        <DrawerContent>
          <DrawerWrapper>
            <Picker />
            <div className='h-6' />
          </DrawerWrapper>
        </DrawerContent>
      </Drawer>
    );
  }

  return (
    <Popover open={ open } onOpenChange={ setOpen }>
      <PopoverTrigger asChild>
        <Button
          size={ props.size }
          variant={ props.variant }
          role='combobox'
          aria-expanded={ open }
          disabled={ props.disabled }
          className={ mergeTailwind(
            comboboxVariants({ variant: props.variant, size: props.size }), 
            props.className,
          ) }
        >
          <div className='text-ellipsis text-nowrap min-w-0 overflow-hidden text-inherit'>
            { props.value
              ? props.options.find(option => option.value === props.value)?.label
              : props.placeholder }
          </div>
          <ChevronsUpDown className='h-3 w-3 shrink-0 opacity-50' />
        </Button>
      </PopoverTrigger>
      <PopoverContent>
        <Picker />
      </PopoverContent>
    </Popover>
  );
}
