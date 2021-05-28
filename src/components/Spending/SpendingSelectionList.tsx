import { Checkbox, List, ListItem, ListItemIcon, Typography } from '@material-ui/core';
import Spending from 'data/Spending';
import { Map } from 'immutable';
import React, { Component } from 'react';
import { connect } from 'react-redux';
import { getSpending } from 'shared/spending/selectors/getSpending';

export interface PropTypes {
  value: number | null;
  onChange: { (spending: Spending | null): void };
  disabled?: boolean;
  excludeIds?: number[];
  excludeSafeToSpend?: boolean;
}

interface WithConnectionPropTypes extends PropTypes {
  spending: Map<number, Spending>;
}

export class SpendingSelectionList extends Component<WithConnectionPropTypes, {}> {

  selectItem = (spendingId: number | null) => () => {
    const { onChange, spending, value } = this.props;
    if (spendingId === value) {
      return;
    }

    return onChange(spending.get(spendingId, null));
  };

  render() {
    const { spending, value, disabled, excludeIds, excludeSafeToSpend } = this.props;

    return (
      <div className="w-full spending-selection-list">
        <List className="p-0">
          { !excludeSafeToSpend &&
          <ListItem
            key={ 'safe' }
            onClick={ this.selectItem(null) }
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
            <div className="grid grid-cols-3 grid-rows-1 grid-flow-col gap-1 w-full">
              <div className="col-span-3">
                <Typography>Safe To Spend</Typography>
              </div>
            </div>
          </ListItem>
          }

          {
            spending
              .filter(item => !excludeIds?.includes(item.spendingId))
              .sortBy(item => item.name)
              .map(item => (
                <ListItem
                  key={ `${ item.spendingId }` }
                  onClick={ this.selectItem(item.spendingId) }
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
                  <div className="grid grid-cols-3 grid-rows-1 grid-flow-col gap-1 w-full">
                    <div className="col-span-2">
                      <Typography>{ item.name }</Typography>
                    </div>
                    <div className="col-span-1 flex justify-end">
                      <Typography>{ item.getCurrentAmountString() }</Typography>
                    </div>
                  </div>
                </ListItem>
              )).valueSeq().toArray()
          }
        </List>
      </div>
    );
  }
}

export default connect(
  state => ({
    spending: getSpending(state),
  }),
  {}
)(SpendingSelectionList);
