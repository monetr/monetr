import React from 'react';
import { Close } from '@mui/icons-material';
import { Button, Dialog, DialogActions, DialogContent, DialogTitle, IconButton, Typography } from '@mui/material';

interface Props {
  open: boolean;
  onCancel: () => void;
  onConfirm: () => void;
}

export default function DuplicateInstitutionDialog(props: Props): JSX.Element {
  const { open, onCancel, onConfirm } = props;
  return (
    <Dialog open={ open } disableEnforceFocus={ true } maxWidth="xs">
      <DialogTitle>
        <div className="flex items-center">
          <span className="text-2xl flex-auto">
              There is already a link for this bank account
          </span>
          <IconButton className="flex-none" onClick={ onCancel }>
            <Close />
          </IconButton>
        </div>
      </DialogTitle>
      <DialogContent className="mb-5">
        <Typography>
          There is already a link for this bank, are you sure that this bank account has not already been linked?
          This is just a warning to make sure that the same bank account does not get added twice.
        </Typography>
        <Typography>
          If you think this is a mistake you can click continue, if you have already authenticated this bank account
          then the existing link should be updated instead.
        </Typography>
      </DialogContent>
      <DialogActions>
        <Button color="secondary" onClick={ onCancel }>
          Cancel
        </Button>
        <Button color="primary" onClick={ onConfirm }>
          Continue
        </Button>
      </DialogActions>
    </Dialog>
  );
}

