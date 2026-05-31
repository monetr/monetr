import { type FormEvent, useCallback, useRef, useState } from 'react';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { useQueryClient } from '@tanstack/react-query';
import { FileUp } from 'lucide-react';
import { useDropzone } from 'react-dropzone';

import { Button } from '@monetr/interface/components/Button';
import MModal, { type MModalRef } from '@monetr/interface/components/MModal';
import Typography from '@monetr/interface/components/Typography';
import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import ErrorFileStage from '@monetr/interface/modals/UploadTransactions/ErrorFileStage';
import ProcessingFileStage from '@monetr/interface/modals/UploadTransactions/ProcessingFileStage';
import TransactionUpload from '@monetr/interface/models/TransactionUpload';
import fileSize from '@monetr/interface/util/fileSize';
import mergeClasses from '@monetr/interface/util/mergeClasses';
import request, { type ApiResponse } from '@monetr/interface/util/request';
import type { ExtractProps } from '@monetr/interface/util/typescriptEvils';

import styles from './UploadTransactionsModal.module.scss';

export enum UploadTransactionStage {
  FileUpload = 1,
  FieldMapping = 2,
  Preparing = 3,
  Processing = 4,
  Completed = 5,
  Error = 6,
}

function UploadTransactionsModal(): React.JSX.Element {
  const modal = useModal();
  const ref = useRef<MModalRef>(null);
  const queryClient = useQueryClient();
  const selectedBankAccountId = useSelectedBankAccountId();

  const [stage, setStage] = useState<UploadTransactionStage>(UploadTransactionStage.FileUpload);
  const [error, setError] = useState<{ message: string; filename: string } | null>(null);
  const [monetrUpload, setMonetrUpload] = useState<TransactionUpload | null>(null);
  const onClose = useCallback(() => {
    if (stage === UploadTransactionStage.Processing) {
      queryClient.invalidateQueries({ queryKey: [`/api/bank_accounts/${selectedBankAccountId}/transactions`] });
      queryClient.invalidateQueries({ queryKey: [`/api/bank_accounts/${selectedBankAccountId}/balances`] });
    }
    return modal.remove();
  }, [stage, modal, queryClient, selectedBankAccountId]);

  return (
    <MModal className={styles.modal} open={modal.visible} ref={ref}>
      {(() => {
        switch (stage) {
          case UploadTransactionStage.FileUpload:
            return (
              <UploadFileStage
                close={modal.remove}
                setError={setError}
                setResult={setMonetrUpload}
                setStage={setStage}
              />
            );
          case UploadTransactionStage.Processing:
            // We only reach the processing stage once the upload has been set, but narrow here so the stage component
            // always receives a defined upload.
            return monetrUpload ? (
              <ProcessingFileStage close={onClose} setStage={setStage} upload={monetrUpload} />
            ) : null;
          case UploadTransactionStage.Completed:
            return null;
          case UploadTransactionStage.Error:
            // Likewise the error stage is only reached once an error has been set.
            return error ? <ErrorFileStage close={onClose} error={error} /> : null;
          default:
            return null;
        }
      })()}
    </MModal>
  );
}

interface StageProps {
  close: () => void;
  setResult: (result: TransactionUpload) => void;
  setStage: (stage: UploadTransactionStage) => void;
  setError: (error: { message: string; filename: string }) => void;
}

function UploadFileStage(props: StageProps) {
  const selectedBankAccountId = useSelectedBankAccountId();
  const [file, setFile] = useState<File | null>(null);
  const [uploadProgress, setUploadProgress] = useState(-1);
  const onDrop = useCallback((acceptedFiles: Array<File>) => {
    const selectedFile = acceptedFiles[0];
    setFile(selectedFile);
  }, []);
  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: {
      'application/vnd.intu.QFX': ['.ofx', '.qfx'],
    },
  });

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();

    // If the file has not been presented then do nothing!
    if (!file) {
      return;
    }

    const formData = new FormData();
    formData.append('data', file, file.name);

    setUploadProgress(0);

    return request<TransactionUpload>({
      method: 'POST',
      url: `/api/bank_accounts/${selectedBankAccountId}/transactions/upload`,
      data: formData,
      onUploadProgress: progressEvent => {
        const percentCompleted = Math.round((progressEvent.loaded * 100) / progressEvent.total);
        setUploadProgress(percentCompleted);
      },
    })
      .then((result: ApiResponse<TransactionUpload>) => {
        props.setResult(new TransactionUpload(result.data));
        props.setStage(UploadTransactionStage.Processing);
      })
      .catch(error => {
        console.error('file upload failed', error);
        const message = error.response.data.error || 'Unkown error';
        props.setError({
          message,
          filename: file.name,
        });
        props.setStage(UploadTransactionStage.Error);
      });
  }

  const uploadClassNames = mergeClasses(styles.dropzone, { [styles.dropzoneActive]: isDragActive });

  if (uploadProgress >= 0) {
    return (
      <div className={styles.stage}>
        <div className={styles.stageBody}>
          <div className={styles.stageHeader}>
            <Typography size='xl' weight='bold'>
              Upload Transactions
            </Typography>
            <div>{/* TODO Close button */}</div>
          </div>

          <div className={styles.filePreview}>
            <FileUp className={styles.fileIcon} />
            <div className={styles.fileInfo}>
              <Typography size='lg'>{file?.name}</Typography>
              <div className={styles.progressTrack}>
                <div className={styles.progressBar} style={{ width: `${uploadProgress}%` }} />
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (file) {
    return (
      <form className={styles.stage} onSubmit={handleSubmit}>
        <div className={styles.stageBody}>
          <div className={styles.stageHeader}>
            <Typography size='xl' weight='bold'>
              Upload Transactions
            </Typography>
            <div>{/* TODO Close button */}</div>
          </div>
          <Typography>
            Upload a QFX or OFX file to import transaction data manually into your account. Maximum of 5MB.
          </Typography>

          <div className={styles.filePreview}>
            <FileUp className={styles.fileIcon} />
            <div className={styles.fileInfo}>
              <Typography size='lg'>{file.name}</Typography>
              <Typography>{fileSize(file.size)}</Typography>
            </div>
          </div>
        </div>
        <div className={styles.actions}>
          <Button onClick={props.close} variant='secondary'>
            Cancel
          </Button>
          <Button type='submit' variant='primary'>
            Upload
          </Button>
        </div>
      </form>
    );
  }

  return (
    <form className={styles.stage} onSubmit={handleSubmit}>
      <div className={styles.stageBody}>
        <div className={styles.stageHeader}>
          <Typography size='xl' weight='bold'>
            Upload Transactions
          </Typography>
          <div>{/* TODO Close button */}</div>
        </div>
        <Typography>
          Upload a QFX or OFX file to import transaction data manually into your account. Maximum of 5MB.
        </Typography>

        <div {...getRootProps()} className={uploadClassNames}>
          <input {...getInputProps()} />
          <FileUp className={styles.fileIcon} />
          <Typography size='lg' weight='semibold'>
            Drag OFX file here
          </Typography>
          <Typography>Or click to browse</Typography>
        </div>
      </div>
      <div className={styles.actions}>
        <Button onClick={props.close} variant='secondary'>
          Cancel
        </Button>
        <Button type='submit' variant='primary'>
          Upload
        </Button>
      </div>
    </form>
  );
}

const uploadTransactionsModal = NiceModal.create(UploadTransactionsModal);

export default uploadTransactionsModal;

export function showUploadTransactionsModal(): Promise<void> {
  return NiceModal.show<
    void,
    ExtractProps<typeof uploadTransactionsModal>,
    Partial<ExtractProps<typeof uploadTransactionsModal>>
  >(uploadTransactionsModal);
}
