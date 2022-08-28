import React, { Fragment } from 'react';
import { Close } from '@mui/icons-material';
import { Dialog, DialogContent, DialogTitle, IconButton, Typography } from '@mui/material';


interface Props {
  open: boolean;
  onClose: () => void;
}

export default function AddManualBankAccountDialog(props: Props): JSX.Element {
  const { open, onClose } = props;

  return (
    <Fragment>
      <Dialog open={ open }>
        <DialogTitle>
          <div className="flex items-center">
            <span className="text-2xl flex-auto mr-5">
                Create a manual bank account
            </span>
            <IconButton className="flex-none" onClick={ onClose }>
              <Close />
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
  );
}
