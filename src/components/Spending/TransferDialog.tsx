import { Button, Dialog, DialogActions, DialogContent, DialogTitle, List, ListItem } from "@material-ui/core";
import Spending from 'data/Spending';
import React, { Component } from "react";
import { connect } from 'react-redux';
import { getSpendingById } from 'shared/spending/selectors/getSpendingById';

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

interface State {
  from: Spending | null;
  to: Spending | null;
}

const SafeToSpend = new Spending({
  spendingId: -1, // Indicates that this is safe to spend.
  name: 'Safe-To-Spend',
});

class TransferDialog extends Component<WithConnectionPropTypes, {}> {

  componentDidMount() {

  }

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
            <ListItem key="to" button>
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

export default connect(
  (state, props: PropTypes) => {
    let from: Spending, to: Spending;

    switch (props.fromSpendingId) {
      case null:
      case undefined:
        break;
      case 0:
        from = SafeToSpend;
        break;
      default:
        from = getSpendingById(props.fromSpendingId)(state);
    }

    switch (props.toSpendingId) {
      case null:
      case undefined:
        break;
      case 0:
        to = SafeToSpend;
        break;
      default:
        to = getSpendingById(props.toSpendingId)(state);
    }

    return {
      from,
      to,
    };
  },
  {}
)(TransferDialog);
