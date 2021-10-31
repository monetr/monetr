import React, { Component, Fragment } from 'react';
import { Dialog, DialogContent, DialogTitle, IconButton, Typography } from '@mui/material';
import { Close } from '@mui/icons-material';
import { connect } from 'react-redux';
import { getLinks } from 'shared/links/selectors/getLinks';
import { AppState } from 'store';

export interface PropTypes {
  open: boolean;
  onClose: () => void;
}

class AddManualBankAccountDialog extends Component<PropTypes, any> {

  render() {
    const { open, onClose } = this.props;

    return (
      <Fragment>
        <Dialog open={ open }>
          <DialogTitle>
            <div className="flex items-center">
              <span className="text-2xl flex-auto mr-5">
                Create a manual bank account
              </span>
              <IconButton className="flex-none" onClick={ onClose }>
                <Close/>
              </IconButton>
            </div>
          </DialogTitle>
          <DialogContent>
            <Typography>
              What do you want to call your manual bank account?
            </Typography>

          </DialogContent>
        </Dialog>
      </Fragment>
    )
  }
}

export default connect(
  (state: AppState) => ({
    links: getLinks(state)
  }),
  {}
)(AddManualBankAccountDialog);
