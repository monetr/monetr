import React, { useEffect, useState } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { Button, CircularProgress, Dialog, DialogActions, DialogContent, DialogTitle } from '@mui/material';
import axios from 'axios';


function NoticeDialog(): JSX.Element {
  const modal = useModal();
  const [noticeString, setNoticeString] = useState<string | null>();

  useEffect(() => {
    if (!noticeString) {
      axios.get<string>('/api/NOTICE')
        .then(result => result.data && setNoticeString(result.data));
    }
  }, [noticeString]);

  return (
    <Dialog open={ modal.visible } maxWidth='md'>
      <DialogTitle>
        Third Party Notices
      </DialogTitle>
      <DialogContent>
        { noticeString &&
          <pre className='whitespace-pre-line max-w-4xl'>
            { noticeString }
          </pre>
        }
        { !noticeString &&
          <div className='w-full flex justify-center'>
            <CircularProgress />
          </div>
        }
      </DialogContent>
      <DialogActions>
        <Button
          color='secondary'
          onClick={ modal.remove }
        >
          Close
        </Button>
      </DialogActions>
    </Dialog>
  );
}

const noticeModal = NiceModal.create(NoticeDialog);

export default noticeModal;

export function showNoticeDialog(): void {
  NiceModal.show(noticeModal);
}
