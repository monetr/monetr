import { Comment } from "@mui/icons-material";
import { Checkbox, Divider, IconButton, List, ListItem, ListItemAvatar, ListItemButton, ListItemIcon, ListItemText, SwipeableDrawer } from "@mui/material";
import { useCurrentBalance } from "hooks/balances";
import { useSpendingSink } from "hooks/spending";
import { useUpdateTransaction } from "hooks/transactions";
import Transaction from "models/Transaction";
import React, { Fragment } from "react";
import formatAmount from "util/formatAmount";

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
    }
  }

  const items: Array<JSX.Element> = [
    {
      spendingId: null,
      name: 'Safe-To-Spend',
      currentAmount: balances?.safe,
    },
    ...(Array.from(allSpending.values()).sort((a, b) => a.name.toLowerCase() > b.name.toLowerCase() ? 1 : -1)),
    ]
      .map(item => (
        <Fragment>
          <ListItem
            key={ item.spendingId }
            secondaryAction={
              <span className="opacity-90 text-sm">
                { formatAmount(item.currentAmount) }
              </span>
            }
            disablePadding
          >
            <ListItemButton
              role={undefined}
              onClick={ changeSelection(item.spendingId) }
              dense
            >
              <ListItemAvatar
                style={{
                  minWidth: 'unset',
                }}
              >
                <Checkbox
                  edge="start"
                  checked={ props.value === item.spendingId }
                  tabIndex={-1}
                  disableRipple
                />
              </ListItemAvatar>
              <ListItemText
                primary={item.name}
                primaryTypographyProps={{
                  className: 'text-md',
                }}
              />
            </ListItemButton>
          </ListItem>
          <Divider />
        </Fragment>
      ))


  return (
    <SwipeableDrawer
      style={{
        zIndex: 1300,
      }}
      anchor='right'
      open={ open }
      onClose={ onClose }
      disableSwipeToOpen
      onOpen={ () => {} }
    >
      <List className="w-[80vw]">
        { items }
      </List>
    </SwipeableDrawer>
  )
}

