import type React from 'react';
import { useCallback, useEffect, useId, useMemo, useRef, useState } from 'react';
import { type UseComboboxSelectedItemChange, useCombobox } from 'downshift';

import { SelectSpendingOptionComponent } from '@monetr/interface/components/MSelectSpending';
import { defaultFilterImplementation, SelectIndicator, type SelectOption } from '@monetr/interface/components/Select';
import { Skeleton } from '@monetr/interface/components/Skeleton';
import { useCurrentBalance } from '@monetr/interface/hooks/useCurrentBalance';
import { useSpendings } from '@monetr/interface/hooks/useSpendings';
import { useUpdateTransaction } from '@monetr/interface/hooks/useUpdateTransaction';
import Spending from '@monetr/interface/models/Spending';
import Transaction from '@monetr/interface/models/Transaction';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import styles from './TransactionItemSelectSpending.module.scss';
import inputStyles from '../FormTextField.module.scss';
import selectStyles from '../Select.module.scss';

const FREE_TO_USE = 'spnd_freeToUse';

export interface TransactionItemSelectSpendingProps {
  transaction: Transaction;
}

export default function TransactionItemSelectSpending(props: TransactionItemSelectSpendingProps): React.JSX.Element {
  const id = useId();
  const { data: spending, isLoading: spendingIsLoading } = useSpendings();
  const { data: balances, isLoading: balancesIsLoading } = useCurrentBalance();
  const updateTransaction = useUpdateTransaction();

  const options: Array<SelectOption<Spending>> = useMemo(
    () => [
      {
        label: 'Free-To-Use',
        value: new Spending({
          spendingId: FREE_TO_USE,
          // It is possible for the "safe" balance to not be present when switching bank accounts. This is a pseudo race
          // condition. Instead we want to gracefully handle the value not being present initially, and print a nicer string
          // until the balance is loaded.
          currentAmount: balances?.free,
        }),
      },
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
    async (newValue: SelectOption<Spending>) => {
      if (newValue.value.spendingId === props.transaction.spendingId) {
        return Promise.resolve();
      }

      const newSpendingId = newValue.value.spendingId === FREE_TO_USE ? null : newValue.value.spendingId;

      const updatedTransaction = new Transaction({
        ...props.transaction,
        spendingId: newSpendingId,
      });

      return await updateTransaction(updatedTransaction).finally(() => {
        // Needs to be in a timeout for some reason. But basically re-focus the select after we have updated the
        // spending.
        setTimeout(() => {
          document.getElementById(id).focus();
        }, 50);
      });
    },
    [props, id, updateTransaction],
  );

  if (spendingIsLoading || balancesIsLoading) {
    return (
      <div className={styles.selectSpendingRoot}>
        <div className={styles.selectSpendingBlock}>
          <div className={mergeTailwind(inputStyles.input, selectStyles.select, styles.selectSpendingWrapper)}>
            <Skeleton className={styles.selectSpendingSkeleton} />
          </div>
        </div>
      </div>
    );
  }

  return <InnerSelect id={id} options={options} value={value} onChange={onChange} />;
}

interface InnerSelectProps<Spending> {
  id: string;
  value?: SelectOption<Spending>;
  options: Array<SelectOption<Spending>>;
  onChange: (newValue: SelectOption<Spending>) => void;
}

function InnerSelect({ id, value, options, onChange }: InnerSelectProps<Spending>): React.JSX.Element {
  const inputWrapperRef = useRef<HTMLDivElement>(null);
  const [items, setItems] = useState<Array<SelectOption<Spending>>>(options);
  const { isOpen, getMenuProps, getInputProps, getItemProps, openMenu, selectedItem } = useCombobox({
    selectedItem: value,
    // By default the highest item should be "highlighted" unless the user moves the highlight themselves.
    defaultHighlightedIndex: 0,
    // But the initially highlighted item should be the one they have selected otherwise fallback to the first item.
    initialHighlightedIndex: value ? options.indexOf(value) : 0,
    onInputValueChange({ inputValue, isOpen }) {
      // Only filter items if we are open!
      if (isOpen) {
        setItems(options.filter(defaultFilterImplementation<Spending>(inputValue)));
      }
    },
    onSelectedItemChange(changes: UseComboboxSelectedItemChange<SelectOption<Spending>>) {
      if (changes.selectedItem) {
        onChange(changes.selectedItem);
      }
    },
    items,
    itemToString(item: SelectOption<Spending>) {
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

  return (
    <div className={styles.selectSpendingRoot}>
      <div className={styles.selectSpendingBlock}>
        {/** biome-ignore lint/a11y/noStaticElementInteractions: Need to account for weird padding here */}
        <div
          onClick={onOpenClickHandler}
          ref={inputWrapperRef}
          className={mergeTailwind(inputStyles.input, selectStyles.select, styles.selectSpendingWrapper)}
        >
          <input
            {...getInputProps({
              id,
              className: styles.selectSpendingInput,
              onFocus: openMenu,
              spellCheck: false,
              'data-freetouse': value.value.spendingId === FREE_TO_USE,
              autoComplete: 'off',
            })}
          />
          <SelectIndicator open={isOpen} />
        </div>
        <ul
          className={mergeTailwind(selectStyles.unorderedList, '', {
            hidden: !(isOpen && items.length),
          })}
          {...getMenuProps()}
          style={renderStyles}
        >
          {isOpen &&
            items.map((item, index) => (
              <li
                key={item.label}
                className={mergeTailwind(
                  [
                    'text-dark-monetr-content-emphasis',
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
                <SelectSpendingOptionComponent {...item} selected={selectedItem?.value === item.value} />
              </li>
            ))}
        </ul>
      </div>
    </div>
  );
}
