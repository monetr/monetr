import React from 'react';
import { Checkbox, List, ListItem, ListItemIcon, Typography } from '@mui/material';
import * as R from 'ramda';

import { useSpendingSink } from 'hooks/spending';
import Spending from 'models/Spending';

interface Props {
  value: number | null;
  onChange: { (spending: Spending | null): void };
  disabled?: boolean;
  excludeIds?: number[];
  excludeSafeToSpend?: boolean;
}

export default function SpendingSelectionList(props: Props): JSX.Element {
  const { result: spending } = useSpendingSink();

  const selectItem = (spendingId: number | null) => () => {
    const { onChange, value } = props;
    if (spendingId === value) {
      return;
    }

    return onChange(spending.get(spendingId));
  };

  const { value, disabled, excludeIds, excludeSafeToSpend } = props;
  const items = R.pipe(
    R.filter((item: Spending) => !excludeIds?.includes(item.spendingId)),
    R.sortBy(item => item.name.toLowerCase()),
    R.map(item => (
      <ListItem
        key={ `${ item.spendingId }` }
        onClick={ selectItem(item.spendingId) }
        button
      >
        <ListItemIcon>
          <Checkbox
            edge="start"
            checked={ value === item.spendingId }
            tabIndex={ -1 }
            color="primary"
            disabled={ !!disabled }
          />
        </ListItemIcon>
        <div className="w-full grid grid-cols-3 grid-rows-1 grid-flow-col gap-1">
          <div className="col-span-2">
            <Typography>{ item.name }</Typography>
          </div>
          <div className="flex justify-end col-span-1">
            <Typography>{ item.getCurrentAmountString() }</Typography>
          </div>
        </div>
      </ListItem>
    )),
  )(Array.from(spending.values()));

  return (
    <div className="w-full spending-selection-list">
      <List className="p-0">
        { !excludeSafeToSpend &&
          <ListItem
            key="safe"
            onClick={ selectItem(null) }
            button
          >
            <ListItemIcon>
              <Checkbox
                edge="start"
                checked={ !value }
                tabIndex={ -1 }
                color="primary"
                disabled={ !!disabled }
              />
            </ListItemIcon>
            <div className="w-full grid grid-cols-3 grid-rows-1 grid-flow-col gap-1">
              <div className="col-span-3">
                <Typography>Safe To Spend</Typography>
              </div>
            </div>
          </ListItem>
        }
        { items }
      </List>
    </div>
  );
}

