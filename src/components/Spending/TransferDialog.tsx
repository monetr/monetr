import { Button, Dialog, DialogActions, DialogContent, DialogTitle, List, ListItem } from "@material-ui/core";
import Spending from 'data/Spending';
import React, { Component } from "react";

export interface PropTypes {
  fromSpendingId?: number;
  toSpendingId?: number;
  isOpen: boolean;
  onClose: { (): void }
}

interface WithConnectionPropTypes extends PropTypes {
  from: Spending | null;
  to: Spending | null;
}

const SafeToSpend = new Spending({
  spendingId: -1, // Indicates that this is safe to spend.
  name: 'Safe-To-Spend',
});

class TransferDialog extends Component<PropTypes, {}> {

  doTransfer = () => {

  };

  render() {
    const { isOpen, onClose } = this.props;
    return (
      <Dialog open={ isOpen }>
        <DialogTitle>
          Transfer Funds
        </DialogTitle>
        <DialogContent>
          <List>
            <ListItem key="from" button>
              From: Thing
            </ListItem>
            <ListItem key="to">
              To: Place
            </ListItem>
          </List>
        </DialogContent>
        <DialogActions>
          <Button
            onClick={ onClose }
          >
            Cancel
          </Button>
          <Button
            variant="outlined"
            color="primary"
            onClick={ this.doTransfer }
          >
            Transfer
          </Button>
        </DialogActions>
      </Dialog>
    );
  }
}
