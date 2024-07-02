import React from 'react';
import { FilePresentOutlined } from '@mui/icons-material';

import { MBaseButton } from '@monetr/interface/components/MButton';
import MSpan from '@monetr/interface/components/MSpan';


interface ErrorFileStageProps {
  close: () => void;
  error: string;
}

export default function ErrorFileStage(props: ErrorFileStageProps): JSX.Element {
  return (
    <div className='h-full flex flex-col gap-2 p-2 justify-between'>
      <div className='flex flex-col gap-2 h-full'>
        <div className='flex justify-between'>
          <MSpan weight='bold' size='xl'>
            Upload Transactions
          </MSpan>
          <div>
            { /* TODO Close button */ }
          </div>
        </div>

        <div className='flex gap-2 items-center border rounded-md w-full p-2 border-dark-monetr-border'>
          <FilePresentOutlined className='text-6xl text-dark-monetr-content' />
          <div className='flex flex-col py-1 w-full'>
            <MSpan size='lg'>{ props.file.name }</MSpan>
            <MSpan>Failed to import data: { props.error }</MSpan>
          </div>
        </div>
      </div>
      <div className='flex justify-end gap-2 mt-2'>
        <MBaseButton color='secondary' onClick={ props.close }>
          Close
        </MBaseButton>
      </div>
    </div>
  );
}
