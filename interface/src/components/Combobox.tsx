import * as React from 'react';
import { Fragment } from 'react';

import { Button } from '@monetr/interface/components/Button';
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '@monetr/interface/components/Command';
import { Popover, PopoverContent, PopoverTrigger } from '@monetr/interface/components/Popover';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import { cva, VariantProps } from 'class-variance-authority';
import { Check, ChevronsUpDown } from 'lucide-react';
 
const comboboxVariants = cva(
  [
    'w-full justify-between',
  ],
  {
    variants: {
      variant: {
        text: [
          'ring-0 group',
          // DROPDOWN IS CLOSED
          // When it's closed, only show the border when someone hovers over.
          'data-[state="closed"]:enabled:hover:ring-1',
          'data-[state="closed"]:enabled:hover:ring-dark-monetr-border-string',
          'data-[state="closed"]:focus:ring-0',
          // When the dropdown is closed, don't show any icons unless they are hovering.
          '[&_svg]:data-[state="closed"]:hover:block [&_svg]:data-[state="closed"]:hidden',


          // DROPDOWN IS OPEN
          // When its open, show the border all the time with the primary color.
          'data-[state="open"]:ring-dark-monetr-brand data-[state="open"]:ring-2',
          // When the dropdown is open then show icons
          '[&_svg]:data-[state="open"]:block',

          '',
        ],
      },
    },
    defaultVariants: {
      variant: 'text',
    },
  },
);

export interface ComboboxOption<V = string> {
  value: V;
  label: string;
  disabled?: boolean;
}

export interface ComboboxItemProps<V = string, O = ComboboxOption<V>> {
  currentValue: V;
  option: O;
}

export interface ComboboxProps<V = string> extends 
  VariantProps<typeof comboboxVariants> {
  disabled?: boolean;
  options: Array<ComboboxOption<V>>;
  value?: V;
  emptyString?: string;
  placeholder?: string;
  searchPlaceholder?: string;
  showSearch?: boolean;
  onSelect?: (value: V) => void;
  components?: {
    Item?: React.ComponentType<ComboboxItemProps<V>>;
  }
}

export function ComboboxItem(props: ComboboxItemProps): JSX.Element {
  return (
    <Fragment>
      <Check
        className={ mergeTailwind(
          'mr-2 h-4 w-4',
          props.currentValue == props.option.value ? 'opacity-100' : 'opacity-0'
        ) }
      />
      { props.option.label }
    </Fragment>
  );
}
 
export function Combobox(props: ComboboxProps) {
  const { Item } = {
    Item: ComboboxItem,
    ...props.components,
  };
  const [open, setOpen] = React.useState(false);
 
  return (
    <Popover open={ open } onOpenChange={ setOpen }>
      <PopoverTrigger asChild>
        <Button
          size='md'
          variant='outlined'
          role='combobox'
          aria-expanded={ open }
          className={ comboboxVariants({ variant: 'text' }) }
        >
          {props.value
            ? props.options.find(option => option.value === props.value)?.label
            : props.placeholder }
          <ChevronsUpDown className='ml-2 h-4 w-4 shrink-0 opacity-50' />
        </Button>
      </PopoverTrigger>
      <PopoverContent className='p-0'>
        <Command>
          { ((props.options.length > 1 && props.showSearch !== false) || props.showSearch) && (
            <CommandInput placeholder={ props.searchPlaceholder } />
          ) }
          <CommandList>
            <CommandEmpty>{ props.emptyString }</CommandEmpty>
            <CommandGroup>
              { props.options.map(option => (
                <CommandItem
                  key={ option.value }
                  value={ option.value.toString() }
                  onSelect={ currentValue => {
                    props.onSelect && props.onSelect(currentValue);
                    setOpen(false);
                  } }
                >
                  <Item currentValue={ props.value } option={ option } />
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
