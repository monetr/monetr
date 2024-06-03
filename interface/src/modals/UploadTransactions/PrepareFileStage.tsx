import React, { useEffect } from 'react';
import { FilePresentOutlined } from '@mui/icons-material';
import axios, { AxiosError, AxiosResponse } from 'axios';
import { enqueueSnackbar } from 'notistack';

import MSpan from '@monetr/interface/components/MSpan';
import { useSelectedBankAccountId } from '@monetr/interface/hooks/bankAccounts';
import { UploadTransactionStage } from '@monetr/interface/modals/UploadTransactions/UploadTransactionsModal';
import MonetrFile from '@monetr/interface/models/File';
import TransactionUpload from '@monetr/interface/models/TransactionUpload';

interface PrepareFileStageProps {
  close: () => void;
  file: MonetrFile;
  setResult: (result: TransactionUpload) => void;
  setStage: (stage: UploadTransactionStage) => void;
}

export default function PrepareFileStage(props: PrepareFileStageProps): JSX.Element {
  const selectedBankAccountId = useSelectedBankAccountId();

  useEffect(() => {
    if (!selectedBankAccountId) return;

    axios.post(`/api/bank_accounts/${selectedBankAccountId}/transactions/uploads`, {
      fileId: props.file.fileId,
    })
      .then((result: AxiosResponse<TransactionUpload>) => {
        props.setResult(new TransactionUpload(result.data));
        props.setStage(UploadTransactionStage.Processing);
      })
      .catch((error: AxiosError<any>) => {
        props.setStage(UploadTransactionStage.Error);
        const message = error.response.data.error || 'Unkown error';
        enqueueSnackbar(`Failed to create upload session for file: ${message}`, {
          variant: 'error',
          disableWindowBlurListener: true,
        });
        props.close();
      });
  }, [props.file, selectedBankAccountId, props]);

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
            <MSpan size='lg'>Preparing</MSpan>
            <div className='w-full bg-gray-200 rounded-full h-1.5 my-2 dark:bg-gray-700 relative'>
              <div className='absolute top-0 bg-blue-600 h-1.5 rounded-full dark:bg-blue-600' style={ { width: '25%' } }></div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
