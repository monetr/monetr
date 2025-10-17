import { useEffect } from 'react';
import { FilePresentOutlined } from '@mui/icons-material';
import axios, { type AxiosError, type AxiosResponse } from 'axios';

import MSpan from '@monetr/interface/components/MSpan';
import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import { UploadTransactionStage } from '@monetr/interface/modals/UploadTransactions/UploadTransactionsModal';
import type MonetrFile from '@monetr/interface/models/File';
import TransactionUpload from '@monetr/interface/models/TransactionUpload';

interface PrepareFileStageProps {
  close: () => void;
  file: MonetrFile;
  setResult: (result: TransactionUpload) => void;
  setStage: (stage: UploadTransactionStage) => void;
  setError: (error: string) => void;
}

export default function PrepareFileStage(props: PrepareFileStageProps): JSX.Element {
  const selectedBankAccountId = useSelectedBankAccountId();

  useEffect(() => {
    if (!selectedBankAccountId) { return; }

    axios
      .post(`/api/bank_accounts/${selectedBankAccountId}/transactions/upload`, {
        fileId: props.file.fileId,
      })
      .then((result: AxiosResponse<TransactionUpload>) => {
        props.setResult(new TransactionUpload(result.data));
        props.setStage(UploadTransactionStage.Processing);
      })
      .catch((error: AxiosError<any>) => {
        const message = error.response.data.error || 'Unkown error';
        props.setError(message);
        props.setStage(UploadTransactionStage.Error);
      });
  }, [props.file, selectedBankAccountId, props]);

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
            <MSpan size='lg'>{props.file.name}</MSpan>
            <MSpan>Preparing...</MSpan>
          </div>
        </div>
      </div>
    </div>
  );
}
