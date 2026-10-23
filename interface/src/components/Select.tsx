import React, { Fragment, useCallback, useEffect, useMemo, useRef, useState } from 'react';
import {
  type UseComboboxSelectedItemChange,
  type UseComboboxState,
  type UseComboboxStateChangeOptions,
  useCombobox,
} from 'downshift';
import { ArrowDown, ArrowUp, LoaderCircle, PanelBottomClose, PanelBottomOpen } from 'lucide-react';

import { Drawer, DrawerContent, DrawerTrigger, DrawerWrapper } from '@monetr/interface/components/Drawer';
import ErrorText from '@monetr/interface/components/ErrorText';
import Label, { type LabelDecorator } from '@monetr/interface/components/Label';
import SelectActionOption from '@monetr/interface/components/SelectActionOption';
import { Skeleton } from '@monetr/interface/components/Skeleton';
import useIsMobile from '@monetr/interface/hooks/useIsMobile';
import mergeClasses from '@monetr/interface/util/mergeClasses';

import errorTextStyles from './ErrorText.module.scss';
import inputStyles from './FormTextField.module.scss';
import selectStyles from './Select.module.scss';

export interface SelectOption<V> {
  label: string;
  value: V;
  // group puts a small header above this option when it is different from the option before it
  group?: string;
  description?: string;
}

// SelectAction is a row in the menu that does something instead of picking a value, like creating a new expense. It
// never becomes the selected value.
export interface SelectAction {
  label: string;
  description?: string;
  group?: string;
  icon: React.FC<{ className?: string }>;
  onSelect: () => void;
}

export type SelectItem<V> = SelectOption<V> | SelectAction;

export function getIsSelectAction<V>(item: SelectItem<V>): item is SelectAction {
  return 'onSelect' in item;
}

// Downshift checks the event for this flag to skip its own handling, but it doesn't export a type for it.
// See: https://github.com/downshift-js/downshift#customizing-handlers
export type DownshiftEvent = React.SyntheticEvent<Element, Event & { preventDownshiftDefault?: boolean }>;

export interface SelectOptionComponentProps<V = unknown> extends SelectOption<V> {
  selected: boolean;
}

export interface SelectProps<V = unknown> {
  id?: string;
  className?: string;
  name?: string;
  placeholder?: string;
  label?: string;
  labelDecorator?: LabelDecorator;
  error?: string;
  description?: string;
  required?: boolean;
  disabled?: boolean;
  isLoading?: boolean;
  value?: SelectOption<V>;
  // icon is shown in front of the selected value
  icon?: React.ReactNode;
  options: Array<SelectOption<V>>;
  // actions are shown above the options, see SelectAction
  actions?: Array<SelectAction>;
  optionComponent?: React.FC<SelectOptionComponentProps<V>>;
  onChange: (newValue: SelectOption<V>) => void;
  // filterImpl is a function that takes the current filter input and returns a function that evaluates whether or not
  // the current option satisfies that filter input. This can be provided to allow for custom search implementations on
  // the options in a select. If no implementation is specified then the default behavior will be to see if the label
  // contains the provided text.
  filterImpl?: (filterText: string) => (option: SelectItem<V>) => boolean;
}

export interface SelectPropsLoading<V> extends SelectProps<V> {
  isLoading: true;
}

export function defaultFilterImplementation<V = unknown>(filterText: string): (option: SelectItem<V>) => boolean {
  return (option: SelectItem<V>) => {
    const text = filterText.toLocaleLowerCase();
    // Match on the group too, that way typing something like "suggested" finds whatever is under that header
    return option.label.toLocaleLowerCase().includes(text) || Boolean(option.group?.toLocaleLowerCase().includes(text));
  };
}

export function DefaultSelectOptionComponent<V = unknown>(props: SelectOptionComponentProps<V>): React.ReactNode {
  return props.label;
}

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
  const LabelDecorator = props.labelDecorator ?? (() => null);
  return (
    <div className={mergeClasses(errorTextStyles.errorTextPadding, props.className)}>
      <Label disabled={props.disabled} htmlFor={props.id} label={props.label} required={props.required}>
        <LabelDecorator disabled={props.disabled} name={props.name} />
      </Label>
      <div className={mergeClasses(inputStyles.input, selectStyles.selectLoading)} data-error={props.error}>
        <Skeleton className={selectStyles.loadingSkeleton} />
        <SelectIndicator disabled={props.disabled} isLoading={props.isLoading} open={false} />
      </div>
      <ErrorText error={props.error} />
    </div>
  );
}

export function SelectCombobox<V>(props: SelectProps<V>): React.JSX.Element {
  const inputWrapperRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const allItems: Array<SelectItem<V>> = useMemo(
    () => [...(props.actions ?? []), ...props.options],
    [props.actions, props.options],
  );
  const [items, setItems] = useState<Array<SelectItem<V>>>(allItems);
  const filterImplementation = useMemo(() => {
    if (props.filterImpl) {
      return props.filterImpl;
    }

    return defaultFilterImplementation<V>;
  }, [props.filterImpl]);

  const {
    isOpen,
    getMenuProps,
    getInputProps,
    getItemProps,
    openMenu,
    closeMenu,
    selectedItem,
    selectItem,
    setInputValue,
    highlightedIndex,
  } = useCombobox<SelectItem<V>>({
    selectedItem: props.value,
    // By default the highest item should be "highlighted" unless the user moves the highlight themselves.
    defaultHighlightedIndex: 0,
    stateReducer(
      state: UseComboboxState<SelectItem<V>>,
      { type, changes }: UseComboboxStateChangeOptions<SelectItem<V>>,
    ): Partial<UseComboboxState<SelectItem<V>>> {
      // defaultHighlightedIndex always wins when the menu opens, so the selected item would never get highlighted and
      // downshift would never scroll to it. So when we are opening and there is a selection, highlight that instead.
      // Match on the value rather than the object because a lot of callers rebuild their options every render.
      if (changes.isOpen && !state.isOpen && changes.selectedItem && !getIsSelectAction(changes.selectedItem)) {
        const selectedValue = changes.selectedItem.value;
        const index = items.findIndex(item => !getIsSelectAction(item) && item.value === selectedValue);
        if (index >= 0) {
          return { ...changes, highlightedIndex: index };
        }
      }

      // Downshift selects whatever is highlighted when the input loses focus. If that is an action then tabbing away
      // would run it, so keep the current selection and its text instead. Actions only run when they are picked.
      if (
        type === useCombobox.stateChangeTypes.InputBlur &&
        changes.selectedItem &&
        getIsSelectAction(changes.selectedItem)
      ) {
        return {
          ...changes,
          selectedItem: state.selectedItem,
          inputValue: state.selectedItem?.label ?? '',
        };
      }

      // Clicking the input opens the menu and then the click bubbles up to the wrapper which calls openMenu again, same
      // with focus. openMenu recalculates the highlight even when we are already open, which would throw away the one
      // we just picked above. So if we are already open just leave the highlight alone.
      if (type === useCombobox.stateChangeTypes.FunctionOpenMenu && state.isOpen) {
        return { ...changes, highlightedIndex: state.highlightedIndex };
      }

      return changes;
    },
    onInputValueChange({ inputValue, isOpen }) {
      // Only filter items if we are open!
      if (isOpen) {
        setItems(allItems.filter(filterImplementation(inputValue)));
      }
    },
    onSelectedItemChange(changes: UseComboboxSelectedItemChange<SelectItem<V>>) {
      if (changes.selectedItem && !getIsSelectAction(changes.selectedItem)) {
        props?.onChange(changes.selectedItem);
      }
    },
    items,
    itemToString(item: SelectItem<V> | null) {
      return item ? item.label : '';
    },
  });

  // Actions never become the selected value, this just gets typescript to agree with that
  const selectedValue = selectedItem && !getIsSelectAction(selectedItem) ? selectedItem.value : undefined;

  // Picking an action skips downshift's own handling, otherwise the action would become the selected value and its
  // label would end up in the input.
  function selectAction(event: DownshiftEvent, action: SelectAction) {
    event.nativeEvent.preventDownshiftDefault = true;
    closeMenu();
    // If they typed to find the action then put the current value back in the input. This has to happen after the menu
    // is closed or it would filter the menu again.
    setInputValue(selectedItem?.label ?? '');
    action.onSelect();
  }

  function selectHighlightedAction(event: React.KeyboardEvent<HTMLInputElement>) {
    const item = items[highlightedIndex];
    if (event.key === 'Enter' && isOpen && item && getIsSelectAction(item)) {
      selectAction(event, item);
    }
  }

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
      setItems(allItems);
    }
  }, [allItems, isOpen]);

  // When the menu opens highlight all of the text in the input, that way if the user just starts typing it replaces the
  // current value and starts filtering instead of appending to it. This has to happen after the click is done or the
  // browser will move the cursor and undo it.
  useEffect(() => {
    if (isOpen) {
      inputRef.current?.select();
    }
  }, [isOpen]);

  // This effect gives the parent component some basic control over the state of the value of this combobox. Mainly this
  // makes it so that if the parent component (specifically select frequency) changes its value to null. Then we need to
  // clear the combobox value here as well so that way we reflect it. This is because it is a managed component in that
  // sense.
  useEffect(() => {
    if (!props.value) {
      selectItem(null);
    }
  }, [props.value, selectItem]);

  const renderStyles = useMemo(() => {
    // Controls the height of the menu that is rendered, makes sure that we dont render past the bottom of the page.
    if (isOpen && inputWrapperRef.current) {
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

  const LabelDecorator = props.labelDecorator ?? (() => null);

  return (
    <div className={mergeClasses(errorTextStyles.errorTextPadding, props.className)}>
      <Label disabled={props.disabled} htmlFor={props.id} label={props.label} required={props.required}>
        <LabelDecorator disabled={props.disabled} name={props.name} />
      </Label>
      {/** biome-ignore lint/a11y/noStaticElementInteractions: Need to account for weird padding here */}
      <div
        aria-disabled={props.disabled}
        className={mergeClasses(inputStyles.input, selectStyles.select)}
        data-error={Boolean(props.error)}
        onClick={onOpenClickHandler}
        ref={inputWrapperRef}
      >
        {props.icon && <span className={selectStyles.icon}>{props.icon}</span>}
        <input
          {...getInputProps({
            disabled: props.disabled,
            'aria-disabled': props.disabled,
            placeholder: props.placeholder,
            className: selectStyles.input,
            onFocus: openMenu,
            onKeyDown: selectHighlightedAction,
            ref: inputRef,
            spellCheck: false,
            'data-1p-ingore': true,
          })}
        />
        <SelectIndicator disabled={props.disabled} isLoading={props.isLoading} open={isOpen} />
      </div>
      {props.required && !props.value && !props.disabled && (
        <SelectRequiredInput onFocus={() => inputRef.current?.focus()} />
      )}
      <ErrorText description={props.description} error={props.error} />
      <ul
        className={selectStyles.unorderedList}
        data-hidden={!(isOpen && items.length)}
        {...getMenuProps()}
        style={renderStyles}
      >
        {isOpen &&
          items.map((item, index) => {
            const previous = items[index - 1];
            const selected = !getIsSelectAction(item) && selectedValue === item.value;
            return (
              <Fragment key={`${item.group ?? ''}${item.label}`}>
                {item.group && item.group !== previous?.group && (
                  <li className={selectStyles.groupHeader} role='presentation'>
                    {item.group}
                  </li>
                )}
                {!item.group && previous?.group && <li className={selectStyles.separator} role='presentation' />}
                <li
                  className={selectStyles.option}
                  data-selected={selected}
                  {...getItemProps({
                    item,
                    index,
                    onClick: (event: React.MouseEvent) => {
                      if (getIsSelectAction(item)) {
                        selectAction(event, item);
                      }
                    },
                  })}
                >
                  {getIsSelectAction(item) && <SelectActionOption action={item} />}
                  {!getIsSelectAction(item) &&
                    React.createElement<SelectOptionComponentProps<V>>(
                      props.optionComponent ?? DefaultSelectOptionComponent,
                      {
                        ...item,
                        selected,
                      },
                    )}
                </li>
              </Fragment>
            );
          })}
      </ul>
    </div>
  );
}

export function SelectDrawer<V>(props: SelectProps<V>): React.JSX.Element {
  const buttonRef = useRef<HTMLButtonElement>(null);
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

  const items: Array<SelectItem<V>> = [...(props.actions ?? []), ...props.options];

  // Actions usually open something else, like a modal. When the drawer finishes closing it gives focus back to its
  // button, which steals it from whatever the action opened. So remember that an action closed it and skip that.
  // onAnimationEnd would have been nicer but vaul only calls it when the drawer closes itself, not when we close it.
  const closedByActionRef = useRef<boolean>(false);

  function selectAction(action: SelectAction) {
    closedByActionRef.current = true;
    setOpen(false);
    action.onSelect();
  }

  function skipReturningFocus(event: Event) {
    if (closedByActionRef.current) {
      event.preventDefault();
      closedByActionRef.current = false;
    }
  }

  const LabelDecorator = props.labelDecorator ?? (() => null);

  return (
    <div className={mergeClasses(errorTextStyles.errorTextPadding, props.className)}>
      <Label disabled={props.disabled} htmlFor={props.id} label={props.label} required={props.required}>
        <LabelDecorator disabled={props.disabled} name={props.name} />
      </Label>
      <Drawer onOpenChange={setOpen} open={open}>
        <DrawerTrigger asChild>
          <button
            aria-disabled={props.disabled}
            className={mergeClasses(inputStyles.input, selectStyles.select)}
            disabled={props.disabled}
            ref={buttonRef}
            type='button'
          >
            {props.icon && <span className={selectStyles.icon}>{props.icon}</span>}
            <span
              aria-disabled={props.disabled}
              className={selectStyles.selectText}
              data-placeholder={!props.value?.label}
            >
              {props.value?.label ?? props.placeholder}
            </span>
            <SelectIndicator disabled={props.disabled} isLoading={props.isLoading} open={open} />
          </button>
        </DrawerTrigger>
        <DrawerContent onCloseAutoFocus={skipReturningFocus}>
          <DrawerWrapper>
            <ul className={selectStyles.unorderedListMobile}>
              {open &&
                items.map((item, index) => {
                  const previous = items[index - 1];
                  return (
                    <Fragment key={`${item.group ?? ''}${item.label}`}>
                      {item.group && item.group !== previous?.group && (
                        <li className={selectStyles.groupHeader} role='presentation'>
                          {item.group}
                        </li>
                      )}
                      {!item.group && previous?.group && <li className={selectStyles.separator} role='presentation' />}
                      {getIsSelectAction(item) && (
                        <li className={selectStyles.optionTouch} onClick={() => selectAction(item)}>
                          <SelectActionOption action={item} />
                        </li>
                      )}
                      {!getIsSelectAction(item) && (
                        <li
                          className={selectStyles.optionTouch}
                          data-selected={props.value === item}
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
                      )}
                    </Fragment>
                  );
                })}
            </ul>
          </DrawerWrapper>
        </DrawerContent>
      </Drawer>
      {props.required && !props.value && !props.disabled && (
        <SelectRequiredInput onFocus={() => buttonRef.current?.focus()} />
      )}
      <ErrorText description={props.description} error={props.error} />
    </div>
  );
}

interface SelectRequiredInputProps {
  onFocus: () => void;
}

// SelectRequiredInput is a hidden input that is only rendered while a required select has nothing picked. There isn't a
// real form input behind our select, so without this the browser will happily submit the form with it empty.
// react-select did the same thing for us and we lost it when we moved off of it. When the browser focuses this to show
// the "please fill out this field" bubble, we hand focus over to the actual select.
function SelectRequiredInput(props: SelectRequiredInputProps): React.JSX.Element {
  return (
    <input
      aria-hidden='true'
      className={selectStyles.requiredInput}
      onChange={() => {}}
      onFocus={props.onFocus}
      required
      tabIndex={-1}
      value=''
    />
  );
}

interface SelectIndicator {
  isLoading?: boolean;
  disabled?: boolean;
  open?: boolean;
}

export function SelectIndicator({ isLoading, disabled, open }: SelectIndicator): React.JSX.Element {
  const isMobile = useIsMobile();
  // Hover/focus colours come from the `.select` descendant selectors.
  const className = selectStyles.indicator;
  if (isLoading) {
    return <LoaderCircle aria-disabled={disabled} className={className} data-loading='true' />;
  }

  if (isMobile) {
    return open ? (
      <PanelBottomClose aria-disabled={disabled} className={className} />
    ) : (
      <PanelBottomOpen aria-disabled={disabled} className={className} />
    );
  }

  return open ? (
    <ArrowDown aria-disabled={disabled} className={className} />
  ) : (
    <ArrowUp aria-disabled={disabled} className={className} />
  );
}
