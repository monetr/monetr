import React, { useState } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle } from '@mui/material';
import { AxiosError } from 'axios';
import { useSnackbar } from 'notistack';

import { useRemoveFundingSchedule } from 'hooks/fundingSchedules';
import FundingSchedule from 'models/FundingSchedule';

export interface RemoveFundingScheduleDialogProps {
  fundingSchedule: FundingSchedule;
}

function RemoveFundingScheduleDialog(props: RemoveFundingScheduleDialogProps): JSX.Element {
  const modal = useModal();
  const removeFundingSchedule = useRemoveFundingSchedule();
  const { enqueueSnackbar } = useSnackbar();
  const [loading, setLoading] = useState<boolean>(false);

  async function removeSchedule(): Promise<void> {
    setLoading(true);
    return removeFundingSchedule(props.fundingSchedule)
      .then(() => modal.remove())
      .catch((error: AxiosError) => {
        setLoading(false);
        const message = error.response.data['error'] || 'Failed to remove funding schedule.';
        enqueueSnackbar(message, {
          variant: 'error',
          disableWindowBlurListener: true,
        });
      });
  }

  return (
    <Dialog
      open={ modal.visible }
      maxWidth="xs"
    >
      <DialogTitle>
        Remove funding schedule?
      </DialogTitle>
      <DialogContent>
        <DialogContentText>
          Are you sure you want to remove the funding schedule { props.fundingSchedule.name }?
        </DialogContentText>
      </DialogContent>
      <DialogActions>
        <Button
          color="secondary"
          onClick={ modal.remove }
          disabled={ loading }
        >
          Cancel
        </Button>
        <Button
          disabled={ loading }
          onClick={ removeSchedule }
          color="primary"
          type="submit"
        >
          Remove
        </Button>
      </DialogActions>
    </Dialog>
  );
}

const modal = NiceModal.create(RemoveFundingScheduleDialog);

export default modal;

export function showRemoveFundingScheduleDialog(props: RemoveFundingScheduleDialogProps): void {
  NiceModal.show(modal, props);
}
