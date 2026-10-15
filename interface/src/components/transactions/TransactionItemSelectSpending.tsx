import type React from 'react';
import { Fragment, useCallback, useEffect, useId, useMemo, useRef, useState } from 'react';
import {
  type UseComboboxSelectedItemChange,
  type UseComboboxState,
  type UseComboboxStateChangeOptions,
  useCombobox,
} from 'downshift';
import { Plus } from 'lucide-react';
import { rrulestr } from 'rrule';

import { SelectSpendingOptionComponent } from '@monetr/interface/components/MSelectSpending';
import {
  type DownshiftEvent,
  defaultFilterImplementation,
  getIsSelectAction,
  type SelectAction,
  SelectIndicator,
  type SelectItem,
  type SelectOption,
} from '@monetr/interface/components/Select';
import SelectActionOption from '@monetr/interface/components/SelectActionOption';
import { Skeleton } from '@monetr/interface/components/Skeleton';
import { useCurrentBalance } from '@monetr/interface/hooks/useCurrentBalance';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { usePatchTransaction } from '@monetr/interface/hooks/usePatchTransaction';
import { useRecurringTransaction } from '@monetr/interface/hooks/useRecurringTransaction';
import { useSpendings } from '@monetr/interface/hooks/useSpendings';
import { showNewExpenseModal } from '@monetr/interface/modals/NewExpenseModal';
import type { ID } from '@monetr/interface/models/ID';
import type Spending from '@monetr/interface/models/Spending';
import { FREE_TO_USE, FreeToUse } from '@monetr/interface/models/Spending';
import type Transaction from '@monetr/interface/models/Transaction';
import { AmountType } from '@monetr/interface/util/amounts';
import mergeClasses from '@monetr/interface/util/mergeClasses';

import inputStyles from '../FormTextField.module.scss';
import selectStyles from '../Select.module.scss';
import styles from './TransactionItemSelectSpending.module.scss';

// We only ever read a handful of fields off the spending options here, this lets us treat the real spending objects
// and the free-to-use pseudo spending interchangeably. Same shape MSelectSpending uses.
type SpendingOption = Pick<Spending | FreeToUse, 'spendingId' | 'spendingType' | 'currentAmount' | 'name'>;

export interface TransactionItemSelectSpendingProps {
  transaction: Transaction;
}

export default function TransactionItemSelectSpending(props: TransactionItemSelectSpendingProps): React.JSX.Element {
  const id = useId();
  const { data: spending, isLoading: spendingIsLoading } = useSpendings();
  const { data: balances, isLoading: balancesIsLoading } = useCurrentBalance();
  const { data: locale } = useLocaleCurrency();
  const patchTransaction = usePatchTransaction();
  // Every row in the list has one of these selects, so only look up the recurring transaction once the user goes for
  // this one. Otherwise every recurring row would make its own request when the page loads.
  const [engaged, setEngaged] = useState(false);
  const { data: recurring } = useRecurringTransaction(engaged ? props.transaction.transactionRecurringId : null);

  const options: Array<SelectOption<SpendingOption>> = useMemo(() => {
    const result: Array<SelectOption<SpendingOption>> = [
      // The "safe" balance can briefly be missing when switching bank accounts, this is a pseudo race condition. The
      // free-to-use option is derived from that balance so we only patch it in once the balance has loaded. The whole
      // select sits behind a loading guard below until then anyway so we never actually render without it.
      ...(balances ? [{ label: 'Free-To-Use', value: new FreeToUse(balances) }] : []),
      ...(spending ?? [])
        .sort((a, b) => (a.name.toLowerCase() > b.name.toLowerCase() ? 1 : -1))
        .map(item => ({
          label: item.name,
          value: item,
        })),
    ];

    // If the recurring charges are linked to an expense then move that expense to the top as a suggestion. Unless the
    // transaction is already spent from it, then there is nothing to suggest.
    let suggestedSpendingId: ID<Spending> | null = null;
    if (recurring && !recurring.deletedAt && recurring.spendingId !== props.transaction.spendingId) {
      suggestedSpendingId = recurring.spendingId;
    }
    const suggested = result.find(item => suggestedSpendingId && item.value.spendingId === suggestedSpendingId);
    if (!recurring || !suggested) {
      return result;
    }

    let description = `Linked to your ${props.transaction.getName()} charges`;
    if (recurring.autoMatched) {
      description = `monetr picked this since your recent ${props.transaction.getName()} charges were spent from it`;
    }

    return [
      {
        ...suggested,
        group: 'Suggested',
        description,
      },
      ...result.filter(item => item !== suggested),
    ];
  }, [balances, spending, recurring, props.transaction]);
  const actions: Array<SelectAction> = useMemo(() => {
    if (!recurring?.getCanCreateExpense()) {
      return [];
    }

    return [
      {
        group: `This charge repeats ${rrulestr(recurring.ruleset).toText()}`,
        label: `Create Expense From ${props.transaction.getName()}`,
        description: locale?.formatAmount(Math.abs(recurring.lastAmount), AmountType.Stored),
        icon: Plus,
        onSelect: () =>
          showNewExpenseModal({
            recurring,
            transaction: props.transaction,
          }),
      },
    ];
  }, [recurring, locale, props.transaction]);
  const value = useMemo(
    () => options.find(item => item.value.spendingId === (props.transaction.spendingId ?? FREE_TO_USE)),
    [options, props.transaction],
  );
  const onChange = useCallback(
    async (newValue: SelectOption<SpendingOption>) => {
      if (newValue.value.spendingId === props.transaction.spendingId) {
        return Promise.resolve();
      }

      // spendingId can be null when moving the transaction back to free-to-use, we need to send null to the server to
      // actually clear it.
      const newSpendingId = newValue.value.spendingId === FREE_TO_USE ? null : newValue.value.spendingId;

      return await patchTransaction({
        transactionId: props.transaction.transactionId,
        bankAccountId: props.transaction.bankAccountId,
        spendingId: newSpendingId,
      }).finally(() => {
        // Needs to be in a timeout for some reason. But basically re-focus the select after we have updated the
        // spending.
        setTimeout(() => {
          document.getElementById(id)?.focus();
        }, 50);
      });
    },
    [props, id, patchTransaction],
  );

  if (spendingIsLoading || balancesIsLoading) {
    return (
      <div className={styles.selectSpendingRoot}>
        <div className={styles.selectSpendingBlock}>
          <div className={mergeClasses(inputStyles.input, selectStyles.select, styles.selectSpendingWrapper)}>
            <Skeleton className={styles.selectSpendingSkeleton} />
          </div>
        </div>
      </div>
    );
  }

  return (
    <InnerSelect
      actions={actions}
      id={id}
      onChange={onChange}
      onEngage={() => setEngaged(true)}
      options={options}
      value={value}
    />
  );
}

interface InnerSelectProps<T> {
  id: string;
  value?: SelectOption<T>;
  options: Array<SelectOption<T>>;
  actions: Array<SelectAction>;
  onChange: (newValue: SelectOption<T>) => void;
  // onEngage is called when the pointer comes over the select or it gets focus, before the menu actually opens
  onEngage: () => void;
}

function InnerSelect({
  id,
  value,
  options,
  actions,
  onChange,
  onEngage,
}: InnerSelectProps<SpendingOption>): React.JSX.Element {
  const inputWrapperRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const allItems: Array<SelectItem<SpendingOption>> = useMemo(() => [...actions, ...options], [actions, options]);
  const [items, setItems] = useState<Array<SelectItem<SpendingOption>>>(allItems);
  const {
    isOpen,
    getMenuProps,
    getInputProps,
    getItemProps,
    openMenu,
    closeMenu,
    selectedItem,
    setInputValue,
    highlightedIndex,
  } = useCombobox<SelectItem<SpendingOption>>({
    selectedItem: value,
    // By default the highest item should be "highlighted" unless the user moves the highlight themselves.
    defaultHighlightedIndex: 0,
    stateReducer(
      state: UseComboboxState<SelectItem<SpendingOption>>,
      { type, changes }: UseComboboxStateChangeOptions<SelectItem<SpendingOption>>,
    ): Partial<UseComboboxState<SelectItem<SpendingOption>>> {
      // defaultHighlightedIndex always wins when the menu opens, so the selected item would never get highlighted and
      // downshift would never scroll to it. So when we are opening and there is a selection, highlight that instead.
      // Match on the spending ID rather than the object because the options get rebuilt whenever balances change.
      if (changes.isOpen && !state.isOpen && changes.selectedItem && !getIsSelectAction(changes.selectedItem)) {
        const selectedSpendingId = changes.selectedItem.value.spendingId;
        const index = items.findIndex(item => !getIsSelectAction(item) && item.value.spendingId === selectedSpendingId);
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
        setItems(allItems.filter(defaultFilterImplementation<SpendingOption>(inputValue)));
      }
    },
    onSelectedItemChange(changes: UseComboboxSelectedItemChange<SelectItem<SpendingOption>>) {
      if (changes.selectedItem && !getIsSelectAction(changes.selectedItem)) {
        onChange(changes.selectedItem);
      }
    },
    items,
    itemToString(item: SelectItem<SpendingOption> | null) {
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
    openMenu();
  }, [openMenu]);

  useEffect(() => {
    if (!isOpen) {
      // Clear the filter when we are currently open and moving to a closed state, this makes it so that if we re-open
      // the menu it is not filtered. This also means that if the recurring transaction loads while the menu is already
      // open, the new rows wait until the next time it opens instead of shifting things around under the user.
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

  return (
    <div className={styles.selectSpendingRoot}>
      <div className={styles.selectSpendingBlock}>
        {/** biome-ignore lint/a11y/noStaticElementInteractions: Need to account for weird padding here */}
        <div
          className={mergeClasses(inputStyles.input, selectStyles.select, styles.selectSpendingWrapper)}
          onClick={onOpenClickHandler}
          onMouseEnter={onEngage}
          ref={inputWrapperRef}
        >
          <input
            {...getInputProps({
              id,
              className: styles.selectSpendingInput,
              onFocus: () => {
                onEngage();
                openMenu();
              },
              onKeyDown: selectHighlightedAction,
              ref: inputRef,
              spellCheck: false,
              'data-freetouse': value?.value.spendingId === FREE_TO_USE,
              autoComplete: 'off',
            })}
          />
          <SelectIndicator open={isOpen} />
        </div>
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
                    {!getIsSelectAction(item) && <SelectSpendingOptionComponent {...item} selected={selected} />}
                  </li>
                </Fragment>
              );
            })}
        </ul>
      </div>
    </div>
  );
}
