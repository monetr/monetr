import React, { Fragment, useState } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { Close } from '@mui/icons-material';
import {
  Alert,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  Snackbar,
  Typography,
} from '@mui/material';
import { useQueryClient } from '@tanstack/react-query';
import { AxiosError } from 'axios';

import { useLink, useRemoveLink } from 'hooks/links';
import mergeTailwind from 'util/mergeTailwind';

interface RemoveLinkConfirmationDialogProps {
  linkId: number;
}

function RemoveLinkConfirmationDialog(props: RemoveLinkConfirmationDialogProps): JSX.Element {
  const modal = useModal();
  const queryClient = useQueryClient();
  const { data: link } = useLink(props.linkId);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string>(null);
  const removeLink = useRemoveLink();

  async function doRemoveLink(): Promise<void> {
    setLoading(true);
    return removeLink(props.linkId)
      .then(() => void Promise.all([
        queryClient.invalidateQueries(['/links']),
        queryClient.removeQueries([`/links/${props.linkId}`]),
        queryClient.invalidateQueries(['/bank_accounts']),
      ]))
      .then(() => modal.remove())
      .catch((error: AxiosError<{ error: string; }>) => setError(error?.response?.data?.error))
      .finally(() => setLoading(false));
  }

  function ErrorMaybe(): JSX.Element {
    if (!error) {
      return null;
    }

    const onClose = () => setError(null);

    return (
      <Snackbar open autoHideDuration={ 6000 } onClose={ onClose }>
        <Alert onClose={ onClose } severity="error">
          { error }
        </Alert>
      </Snackbar>
    );
  }

  return (
    <Fragment>
      <ErrorMaybe />
      <Dialog
        open={ modal.visible }
        onClose={ modal.remove }
      >
        <DialogTitle>
          <div className="flex items-center">
            <span className="text-2xl flex-auto">
                Remove { link.getName() }
            </span>
            <IconButton
              disabled={ loading }
              className="flex-none"
              onClick={ modal.remove }
            >
              <Close />
            </IconButton>
          </div>
        </DialogTitle>
        <DialogContent>
          <Typography>
            Are you sure you want to remove the <b>{ link.getName() }</b> link? This cannot be undone.
          </Typography>
          { link.getIsPlaid() && <Typography>You can also convert this link to be manual instead.</Typography> }
        </DialogContent>
        <DialogActions>
          <Button
            disabled={ loading }
            onClick={ modal.remove }
          >
            Cancel
          </Button>
          <Button
            disabled={ loading }
            onClick={ doRemoveLink }
            className={ mergeTailwind({
              'text-red-500': !loading,
            }) }
          >
            Remove
          </Button>
        </DialogActions>
      </Dialog>
    </Fragment>
  );
}

const removeLinkConfirmationModal = NiceModal.create(RemoveLinkConfirmationDialog);
export default removeLinkConfirmationModal;

export function showRemoveLinkConfirmationDialog(props: RemoveLinkConfirmationDialogProps): void {
  NiceModal.show(removeLinkConfirmationModal, props);
}
