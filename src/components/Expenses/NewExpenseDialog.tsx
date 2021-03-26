import { Button, Dialog, DialogActions, DialogContent, DialogTitle, IconButton, Typography } from "@material-ui/core";
import React, { Component } from "react";

enum NewExpenseStep {
  Name,
  Amount,
  Recurrence,
  Funding,
}

export interface PropTypes {
  onClose: { (): void };
  isOpen: boolean;
}

export interface State {
  step
}

export class NewExpenseDialog extends Component<PropTypes, any> {

  renderStepContent = () => {

  };

  render() {
    const { onClose, isOpen } = this.props;

    return (
      <Dialog onClose={ onClose } open={ isOpen }>
        <DialogTitle>
          Create a new expense
        </DialogTitle>
        <DialogContent>

        </DialogContent>
        <DialogActions>
          <Button color="primary">
            Disagree
          </Button>
          <Button color="primary">
            Agree
          </Button>
        </DialogActions>
      </Dialog>
    )
  }
}
