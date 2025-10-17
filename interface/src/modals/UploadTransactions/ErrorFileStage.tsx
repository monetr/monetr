import React from 'react';
import { FilePresentOutlined } from '@mui/icons-material';

import { Button } from '@monetr/interface/components/Button';
import MSpan from '@monetr/interface/components/MSpan';

interface ErrorFileStageProps {
  close: () => void;
  error: { message: string; filename: string };
}

export default function ErrorFileStage(props: ErrorFileStageProps): JSX.Element {
  return (
    <div className='h-full flex flex-col gap-2 p-2 justify-between'>
      <div className='flex flex-col gap-2 h-full'>
        <div className='flex justify-between'>
          <MSpan weight='bold' size='xl'>
            Upload Transactions
          </MSpan>
          <div>{/* TODO Close button */}</div>
        </div>

        <div className='flex gap-2 items-center border rounded-md w-full p-2 border-dark-monetr-border'>
          <FilePresentOutlined className='text-6xl text-dark-monetr-content' />
          <div className='flex flex-col py-1 w-full'>
            <MSpan size='lg'>{props.error.filename}</MSpan>
            <MSpan>Failed to import data: {props.error.message}</MSpan>
          </div>
        </div>
      </div>
      <div className='flex justify-end gap-2 mt-2'>
        <Button variant='secondary' onClick={props.close}>
          Close
        </Button>
      </div>
    </div>
  );
}
