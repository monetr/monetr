import type React from 'react';
import { useCallback, useEffect, useId, useMemo, useRef, useState } from 'react';
import { type UseComboboxSelectedItemChange, useCombobox } from 'downshift';

import { SelectSpendingOptionComponent } from '@monetr/interface/components/MSelectSpending';
import { defaultFilterImplementation, SelectIndicator, type SelectOption } from '@monetr/interface/components/Select';
import { Skeleton } from '@monetr/interface/components/Skeleton';
import { useCurrentBalance } from '@monetr/interface/hooks/useCurrentBalance';
import { useSpendings } from '@monetr/interface/hooks/useSpendings';
import { useUpdateTransaction } from '@monetr/interface/hooks/useUpdateTransaction';
import type Spending from '@monetr/interface/models/Spending';
import { FREE_TO_USE, FreeToUse } from '@monetr/interface/models/Spending';
import Transaction from '@monetr/interface/models/Transaction';
import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './TransactionItemSelectSpending.module.scss';
import inputStyles from '../FormTextField.module.scss';
import selectStyles from '../Select.module.scss';

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
  const updateTransaction = useUpdateTransaction();

  const options: Array<SelectOption<SpendingOption>> = useMemo(
    () => [
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
    ],
    [balances, spending],
  );
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

      const updatedTransaction = new Transaction({
        ...props.transaction,
        spendingId: newSpendingId,
      });

      return await updateTransaction(updatedTransaction).finally(() => {
        // Needs to be in a timeout for some reason. But basically re-focus the select after we have updated the
        // spending.
        setTimeout(() => {
          document.getElementById(id)?.focus();
        }, 50);
      });
    },
    [props, id, updateTransaction],
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

  return <InnerSelect id={id} onChange={onChange} options={options} value={value} />;
}

interface InnerSelectProps<T> {
  id: string;
  value?: SelectOption<T>;
  options: Array<SelectOption<T>>;
  onChange: (newValue: SelectOption<T>) => void;
}

function InnerSelect({ id, value, options, onChange }: InnerSelectProps<SpendingOption>): React.JSX.Element {
  const inputWrapperRef = useRef<HTMLDivElement>(null);
  const [items, setItems] = useState<Array<SelectOption<SpendingOption>>>(options);
  const { isOpen, getMenuProps, getInputProps, getItemProps, openMenu, selectedItem } = useCombobox({
    selectedItem: value,
    // By default the highest item should be "highlighted" unless the user moves the highlight themselves.
    defaultHighlightedIndex: 0,
    // But the initially highlighted item should be the one they have selected otherwise fallback to the first item.
    initialHighlightedIndex: value ? options.indexOf(value) : 0,
    onInputValueChange({ inputValue, isOpen }) {
      // Only filter items if we are open!
      if (isOpen) {
        setItems(options.filter(defaultFilterImplementation<SpendingOption>(inputValue)));
      }
    },
    onSelectedItemChange(changes: UseComboboxSelectedItemChange<SelectOption<SpendingOption>>) {
      if (changes.selectedItem) {
        onChange(changes.selectedItem);
      }
    },
    items,
    itemToString(item: SelectOption<SpendingOption> | null) {
      return item ? item.label : '';
    },
  });

  const onOpenClickHandler = useCallback(() => {
    openMenu();
  }, [openMenu]);

  useEffect(() => {
    if (!isOpen) {
      // Clear the filter when we are currently open and moving to a closed state, this makes it so that if we re-open
      // the menu it is not filtered.
      setItems(options);
    }
  }, [options, isOpen]);

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
          ref={inputWrapperRef}
        >
          <input
            {...getInputProps({
              id,
              className: styles.selectSpendingInput,
              onFocus: openMenu,
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
            items.map((item, index) => (
              <li
                className={selectStyles.option}
                data-selected={selectedItem?.value === item.value}
                key={item.label}
                {...getItemProps({
                  item,
                  index,
                })}
              >
                <SelectSpendingOptionComponent {...item} selected={selectedItem?.value === item.value} />
              </li>
            ))}
        </ul>
      </div>
    </div>
  );
}
