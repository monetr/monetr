import { useEffect } from 'react';
import { FilePresentOutlined } from '@mui/icons-material';
import { type UseQueryResult, useQuery, useQueryClient } from '@tanstack/react-query';

import { Button } from '@monetr/interface/components/Button';
import MSpan from '@monetr/interface/components/MSpan';
import type { UploadTransactionStage } from '@monetr/interface/modals/UploadTransactions/UploadTransactionsModal';
import TransactionUpload from '@monetr/interface/models/TransactionUpload';

interface ProcessingFileStageProps {
  close: () => void;
  upload: TransactionUpload;
  setStage: (stage: UploadTransactionStage) => void;
}

export default function ProcessingFileStage(props: ProcessingFileStageProps): JSX.Element {
  const { data } = useTransactionUploadProgress(props.upload.bankAccountId, props.upload.transactionUploadId);

  return (
    <div className='h-full flex flex-col gap-2 p-2 justify-between'>
      <div className='flex flex-col gap-2 h-full'>
        <div className='flex justify-between'>
          <MSpan weight='bold' size='xl'>
            Upload Transactions
          </MSpan>
          <div>{/* TODO Close button */}</div>
        </div>
        <MSpan>
          Your upload is currently pending or processing, keep this window open to see its status. You can also safely
          close this window. Your upload will not be interuptted.
        </MSpan>

        <div className='flex gap-2 items-center border rounded-md w-full p-2 border-dark-monetr-border'>
          <FilePresentOutlined className='text-6xl text-dark-monetr-content' />
          <div className='flex flex-col py-1 w-full'>
            <MSpan size='lg'>{props.upload.file.name}</MSpan>
            <MSpan>Import {data?.status || 'processing'}!</MSpan>
          </div>
        </div>
      </div>
      <div className='flex justify-end gap-2 mt-2'>
        <Button variant='secondary' onClick={props.close}>
          Done
        </Button>
      </div>
    </div>
  );
}

function useTransactionUploadProgress(
  bankAccountId: string,
  transactionUploadId: string,
): UseQueryResult<TransactionUpload, unknown> {
  const queryClient = useQueryClient();
  // Bootstrap the socket to listen for the actual changes.
  useEffect(() => {
    const protocol = location.protocol === 'https:' ? 'wss' : 'ws';
    // eslint-disable-next-line max-len
    const socket = new WebSocket(
      `${protocol}://${location.host}/api/bank_accounts/${bankAccountId}/transactions/upload/${transactionUploadId}/progress`,
    );
    socket.onopen = () => {
      // eslint-disable-next-line no-console
      console.log('Listening for transaction upload progress messages for', {
        bankAccountId,
        transactionUploadId,
      });
    };

    // Whenever we receive a progress message, update our state to represent the new status.
    socket.onmessage = event => {
      if (!event.data) {
        return;
      }
      const data: Partial<TransactionUpload> = JSON.parse(event.data);
      const queryKey = [`/bank_accounts/${bankAccountId}/transactions/upload/${transactionUploadId}`];
      // Take the current upload data stored in state (if its there) and merge it with the message we received. The
      // first message will contain the entire transaction upload object so if its not already in the state then this
      // will persist it. Subsequent messages will contain changes to the status.
      queryClient.setQueryData(queryKey, (item: TransactionUpload | null) => {
        return {
          ...(item ?? {}),
          ...data,
        };
      });
    };

    // On unmount close the socket
    return () => socket.close();
  }, [bankAccountId, queryClient, transactionUploadId]);

  // Subscribe to changes for the transaction upload
  return useQuery<Partial<TransactionUpload>, unknown, TransactionUpload>({
    queryKey: [`/bank_accounts/${bankAccountId}/transactions/upload/${transactionUploadId}`],
    initialData: () => null, // Don't do the initial fetch, rely on the websocket instead.
    select: data => new TransactionUpload(data),
  });
}
