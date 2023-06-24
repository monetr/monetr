import React, { Fragment } from 'react';
import { Checkbox, Divider, List, ListItem, ListItemAvatar, ListItemButton, ListItemText, SwipeableDrawer } from '@mui/material';

import VerticalPuller from 'components/VerticalPuller';
import { useCurrentBalance } from 'hooks/balances';
import { useSpendingSink } from 'hooks/spending';
import formatAmount from 'util/formatAmount';

export interface Props {
  open: boolean;
  onClose: () => void;
  value: number | null;
  onChange: (_value: number | null) => void;
}

export default function TransactionSpentFromSelectionMobile(props: Props): JSX.Element {
  const { open, onClose } = props;
  const { result: allSpending } = useSpendingSink();
  const balances = useCurrentBalance();

  function changeSelection(spendingId: number | null): () => void {
    return () => {
      props.onChange(spendingId);
      props.onClose();
    };
  }

  const items: Array<JSX.Element> = [
    {
      spendingId: null,
      name: 'Free-To-Use',
      currentAmount: balances?.free,
    },
    ...(Array.from(allSpending.values()).sort((a, b) => a.name.toLowerCase() > b.name.toLowerCase() ? 1 : -1)),
  ]
    .map(item => (
      <Fragment>
        <ListItem
          key={ item.spendingId }
          secondaryAction={
            <span className="opacity-90 text-md">
              {formatAmount(item.currentAmount)}
            </span>
          }
          disablePadding
        >
          <ListItemButton
            role={ undefined }
            onClick={ changeSelection(item.spendingId) }
            dense
          >
            <ListItemAvatar
              style={ {
                minWidth: 'unset',
              } }
            >
              <Checkbox
                edge="start"
                checked={ props.value === item.spendingId }
                tabIndex={ -1 }
                disableRipple
              />
            </ListItemAvatar>
            <ListItemText
              primary={ item.name }
              primaryTypographyProps={ {
                className: 'text-lg',
              } }
            />
          </ListItemButton>
        </ListItem>
        <Divider />
      </Fragment>
    ));


  return (
    <SwipeableDrawer
      style={ {
        zIndex: 1300,
      } }
      anchor='right'
      open={ open }
      onClose={ onClose }
      disableSwipeToOpen
      onOpen={ () => { } }
    >
      <div className="w-[90vw] pl-5">
        <VerticalPuller />
        <div className="w-full flex p-3 justify-center">
          <p>What should this be spent from?</p>
        </div>
        <List disablePadding>
          <Divider />
          {items}
        </List>
      </div>
    </SwipeableDrawer>
  );
}

