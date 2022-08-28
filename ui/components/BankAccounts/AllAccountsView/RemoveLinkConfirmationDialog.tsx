import React, { Fragment, useState } from 'react';
import { useQueryClient } from 'react-query';
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
import { AxiosError } from 'axios';
import classnames from 'classnames';

import { useLink, useRemoveLink } from 'hooks/links';

interface Props {
  open: boolean;
  onClose: () => void;
  linkId: number;
}

export default function RemoveLinkConfirmationDialog(props: Props): JSX.Element {
  const queryClient = useQueryClient();
  const link = useLink(props.linkId);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string>(null);
  const removeLink = useRemoveLink();

  async function doRemoveLink(): Promise<void> {
    setLoading(true);
    return removeLink(props.linkId)
      .then(() => props.onClose())
      .then(() => void Promise.all([
        queryClient.invalidateQueries('/links'),
        queryClient.invalidateQueries('/bank_accounts'),
      ]))
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

  const { open, onClose } = props;

  return (
    <Fragment>
      <ErrorMaybe />
      <Dialog open={ open } onClose={ onClose }>
        <DialogTitle>
          <div className="flex items-center">
            <span className="text-2xl flex-auto">
                Remove { link.getName() }
            </span>
            <IconButton
              disabled={ loading }
              className="flex-none"
              onClick={ onClose }
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
            onClick={ onClose }
          >
            Cancel
          </Button>
          <Button
            disabled={ loading }
            onClick={ doRemoveLink }
            className={ classnames({
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
