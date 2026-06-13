import { useEffect } from 'react';
import { type UseQueryResult, useQuery, useQueryClient } from '@tanstack/react-query';
import { FileUp } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import Typography from '@monetr/interface/components/Typography';
import type { UploadTransactionStage } from '@monetr/interface/modals/UploadTransactions/UploadTransactionsModal';
import type TransactionUpload from '@monetr/interface/models/TransactionUpload';

import styles from './ProcessingFileStage.module.scss';

interface ProcessingFileStageProps {
  close: () => void;
  upload: TransactionUpload;
  setStage: (stage: UploadTransactionStage) => void;
}

export default function ProcessingFileStage(props: ProcessingFileStageProps): React.JSX.Element {
  const { data } = useTransactionUploadProgress(props.upload.bankAccountId, props.upload.transactionUploadId);

  return (
    <div className={styles.root}>
      <div className={styles.body}>
        <div className={styles.header}>
          <Typography size='xl' weight='bold'>
            Upload Transactions
          </Typography>
          <div>{/* TODO Close button */}</div>
        </div>
        <Typography size='inherit'>
          Your upload is currently pending or processing, keep this window open to see its status. You can also safely
          close this window. Your upload will not be interuptted.
        </Typography>

        <div className={styles.fileCard}>
          <FileUp className={styles.fileIcon} />
          <div className={styles.fileInfo}>
            <Typography size='lg'>{props.upload.file?.name}</Typography>
            <Typography size='inherit'>Import {data?.status || 'processing'}!</Typography>
          </div>
        </div>
      </div>
      <div className={styles.actions}>
        <Button onClick={props.close} variant='secondary'>
          Done
        </Button>
      </div>
    </div>
  );
}

function useTransactionUploadProgress(
  bankAccountId: string,
  transactionUploadId: string,
): UseQueryResult<Partial<TransactionUpload>, unknown> {
  const queryClient = useQueryClient();
  // Bootstrap the socket to listen for the actual changes.
  useEffect(() => {
    const protocol = location.protocol === 'https:' ? 'wss' : 'ws';
    const socket = new WebSocket(
      `${protocol}://${location.host}/api/bank_accounts/${bankAccountId}/transactions/upload/${transactionUploadId}/progress`,
    );
    socket.onopen = () => {
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
      const queryKey = [`/api/bank_accounts/${bankAccountId}/transactions/upload/${transactionUploadId}`];
      // Take the current upload data stored in state (if its there) and merge it with the message we received. The
      // first message will contain the entire transaction upload object so if its not already in the state then this
      // will persist it. Subsequent messages will contain changes to the status.
      queryClient.setQueryData(queryKey, (item: Partial<TransactionUpload> | undefined) => {
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
  // The upload object is delivered incrementally over the websocket above (first message is the whole object, then
  // status updates) so we leave the query data as a partial rather than hydrating it into a TransactionUpload here.
  return useQuery<Partial<TransactionUpload>, unknown>({
    queryKey: [`/api/bank_accounts/${bankAccountId}/transactions/upload/${transactionUploadId}`],
    initialData: () => ({}), // Don't do the initial fetch, rely on the websocket instead.
  });
}
