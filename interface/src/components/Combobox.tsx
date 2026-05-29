import * as React from 'react';
import { Fragment } from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { Check, ChevronsUpDown } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@monetr/interface/components/Command';
import { Drawer, DrawerContent, DrawerTrigger, DrawerWrapper } from '@monetr/interface/components/Drawer';
import { Popover, PopoverContent, PopoverTrigger } from '@monetr/interface/components/Popover';
import useIsMobile from '@monetr/interface/hooks/useIsMobile';
import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './Combobox.module.scss';

export const comboboxVariants = cva([styles.base], {
  variants: {
    variant: {
      outlined: styles.outlined,
      text: styles.text,
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
});

export interface ComboboxOption<V extends string> {
  value: V;
  label: string;
  disabled?: boolean;
}

export interface ComboboxItemProps<V extends string, O extends ComboboxOption<V>> {
  currentValue: V | undefined;
  option: O;
}

export interface ComboboxProps<V extends string, O extends ComboboxOption<V>>
  extends VariantProps<typeof comboboxVariants> {
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
  };
}

export function ComboboxItem<V extends string, O extends ComboboxOption<V>>(
  props: ComboboxItemProps<V, O>,
): JSX.Element {
  return (
    <Fragment>
      <Check
        className={mergeClasses(styles.checkIcon, {
          [styles.checkIconVisible]: props.currentValue === props.option.value,
          [styles.checkIconHidden]: props.currentValue !== props.option.value,
        })}
      />
      {props.option.label}
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
        {((props.options.length > 1 && props.showSearch !== false) || props.showSearch) && (
          <CommandInput placeholder={props.searchPlaceholder} />
        )}
        <CommandList>
          <CommandEmpty>{props.emptyString}</CommandEmpty>
          <CommandGroup className={styles.group}>
            {props.options.map(option => (
              <CommandItem
                key={option.value}
                onSelect={() => {
                  props.onSelect?.(option.value);
                  setOpen(false);
                }}
                title={option.label}
                value={`${option.label} ${option.value}` /* makes search work properly :( */}
              >
                <Item currentValue={props.value} option={option} />
              </CommandItem>
            ))}
          </CommandGroup>
        </CommandList>
      </Command>
    );
  }

  if (isMobile) {
    return (
      <Drawer onOpenChange={setOpen} open={open}>
        <DrawerTrigger asChild>
          <Button
            aria-expanded={open}
            className={mergeClasses(comboboxVariants({ variant: props.variant, size: props.size }), props.className)}
            disabled={props.disabled}
            role='combobox'
            size={props.size}
            variant={props.variant}
          >
            <div className={styles.triggerLabel}>
              {props.value ? props.options.find(option => option.value === props.value)?.label : props.placeholder}
            </div>
            <ChevronsUpDown className={styles.triggerIcon} />
          </Button>
        </DrawerTrigger>
        <DrawerContent>
          <DrawerWrapper>
            <Picker />
            <div className={styles.drawerSpacer} />
          </DrawerWrapper>
        </DrawerContent>
      </Drawer>
    );
  }

  return (
    <Popover onOpenChange={setOpen} open={open}>
      <PopoverTrigger asChild>
        <Button
          aria-expanded={open}
          className={mergeClasses(comboboxVariants({ variant: props.variant, size: props.size }), props.className)}
          disabled={props.disabled}
          role='combobox'
          size={props.size}
          variant={props.variant}
        >
          <div className={styles.triggerLabel}>
            {props.value ? props.options.find(option => option.value === props.value)?.label : props.placeholder}
          </div>
          <ChevronsUpDown className={styles.triggerIcon} />
        </Button>
      </PopoverTrigger>
      <PopoverContent>
        <Picker />
      </PopoverContent>
    </Popover>
  );
}
