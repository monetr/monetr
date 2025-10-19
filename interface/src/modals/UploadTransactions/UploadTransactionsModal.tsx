import { type FormEvent, useCallback, useRef, useState } from 'react';
import { useDropzone } from 'react-dropzone';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { Close, FilePresentOutlined, UploadFileOutlined } from '@mui/icons-material';
import { useQueryClient } from '@tanstack/react-query';
import axios, { type AxiosProgressEvent, type AxiosResponse } from 'axios';

import { Button } from '@monetr/interface/components/Button';
import MModal, { type MModalRef } from '@monetr/interface/components/MModal';
import MSpan from '@monetr/interface/components/MSpan';
import { useSelectedBankAccountId } from '@monetr/interface/hooks/useSelectedBankAccountId';
import ErrorFileStage from '@monetr/interface/modals/UploadTransactions/ErrorFileStage';
import ProcessingFileStage from '@monetr/interface/modals/UploadTransactions/ProcessingFileStage';
import TransactionUpload from '@monetr/interface/models/TransactionUpload';
import fileSize from '@monetr/interface/util/fileSize';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';
import type { ExtractProps } from '@monetr/interface/util/typescriptEvils';

export enum UploadTransactionStage {
  FileUpload = 1,
  FieldMapping = 2,
  Preparing = 3,
  Processing = 4,
  Completed = 5,
  Error = 6,
}

function UploadTransactionsModal(): JSX.Element {
  const modal = useModal();
  const ref = useRef<MModalRef>(null);
  const queryClient = useQueryClient();
  const selectedBankAccountId = useSelectedBankAccountId();

  const [stage, setStage] = useState<UploadTransactionStage>(UploadTransactionStage.FileUpload);
  const [error, setError] = useState<{ message: string; filename: string } | null>(null);
  const [monetrUpload, setMonetrUpload] = useState<TransactionUpload | null>(null);
  const onClose = useCallback(() => {
    if (stage === UploadTransactionStage.Processing) {
      queryClient.invalidateQueries({ queryKey: [`/bank_accounts/${selectedBankAccountId}/transactions`] });
      queryClient.invalidateQueries({ queryKey: [`/bank_accounts/${selectedBankAccountId}/balances`] });
    }
    return modal.remove();
  }, [stage, modal, queryClient, selectedBankAccountId]);

  function CurrentStage(): JSX.Element {
    switch (stage) {
      case UploadTransactionStage.FileUpload:
        return (
          <UploadFileStage setResult={setMonetrUpload} setStage={setStage} setError={setError} close={modal.remove} />
        );
      case UploadTransactionStage.Processing:
        return <ProcessingFileStage upload={monetrUpload} setStage={setStage} close={onClose} />;
      case UploadTransactionStage.Completed:
        return null;
      case UploadTransactionStage.Error:
        return <ErrorFileStage error={error} close={onClose} />;
      default:
        return null;
    }
  }

  return (
    <MModal open={modal.visible} ref={ref} className='sm:max-w-xl'>
      <CurrentStage />
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

    const config = {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: (progressEvent: AxiosProgressEvent) => {
        const percentCompleted = Math.round((progressEvent.loaded * 100) / progressEvent.total);
        setUploadProgress(percentCompleted);
      },
    };
    setUploadProgress(0);

    return axios
      .post(`/api/bank_accounts/${selectedBankAccountId}/transactions/upload`, formData, config)
      .then((result: AxiosResponse<TransactionUpload>) => {
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

  const uploadClassNames = mergeTailwind(
    'border-dashed rounded-md w-full border p-8 flex justify-center flex-col items-center cursor-pointer',
    { 'border-dark-monetr-border hover:border-dark-monetr-border-string': !isDragActive },
    { 'border-dark-monetr-brand-subtle': isDragActive },
  );

  if (uploadProgress >= 0) {
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
              <MSpan size='lg'>{file.name}</MSpan>
              <div className='w-full bg-gray-200 rounded-full h-1.5 my-2 dark:bg-gray-700 relative'>
                <div
                  className='absolute top-0 bg-green-600 h-1.5 rounded-full dark:bg-green-600'
                  style={{ width: `${uploadProgress}%` }}
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (file) {
    return (
      <form onSubmit={handleSubmit} className='h-full flex flex-col gap-2 p-2 justify-between'>
        <div className='flex flex-col gap-2 h-full'>
          <div className='flex justify-between'>
            <MSpan weight='bold' size='xl'>
              Upload Transactions
            </MSpan>
            <div>{/* TODO Close button */}</div>
          </div>
          <MSpan>Upload a QFX or OFX file to import transaction data manually into your account. Maximum of 5MB.</MSpan>

          <div className='flex gap-2 items-center border rounded-md w-full p-2 border-dark-monetr-border'>
            <FilePresentOutlined className='text-6xl text-dark-monetr-content' />
            <div className='flex flex-col py-1 w-full'>
              <MSpan size='lg'>{file.name}</MSpan>
              <MSpan>{fileSize(file.size)}</MSpan>
            </div>
            <Close className='mr-2 text-dark-monetr-content-subtle hover:text-dark-monetr-content cursor-pointer' />
          </div>
        </div>
        <div className='flex justify-end gap-2 mt-2'>
          <Button variant='secondary' onClick={props.close}>
            Cancel
          </Button>
          <Button variant='primary' type='submit'>
            Upload
          </Button>
        </div>
      </form>
    );
  }

  return (
    <form onSubmit={handleSubmit} className='h-full flex flex-col gap-2 p-2 justify-between'>
      <div className='flex flex-col gap-2 h-full'>
        <div className='flex justify-between'>
          <MSpan weight='bold' size='xl'>
            Upload Transactions
          </MSpan>
          <div>{/* TODO Close button */}</div>
        </div>
        <MSpan>Upload a QFX or OFX file to import transaction data manually into your account. Maximum of 5MB.</MSpan>

        <div {...getRootProps()} className={uploadClassNames}>
          <input {...getInputProps()} />
          <UploadFileOutlined className='text-6xl text-dark-monetr-content' />
          <MSpan size='lg' weight='semibold'>
            Drag OFX file here
          </MSpan>
          <MSpan>Or click to browse</MSpan>
        </div>
      </div>
      <div className='flex justify-end gap-2 mt-2'>
        <Button variant='secondary' onClick={props.close}>
          Cancel
        </Button>
        <Button variant='primary' type='submit'>
          Upload
        </Button>
      </div>
    </form>
  );
}

const uploadTransactionsModal = NiceModal.create(UploadTransactionsModal);

export default uploadTransactionsModal;

export function showUploadTransactionsModal(): Promise<void> {
  return NiceModal.show<void, ExtractProps<typeof uploadTransactionsModal>, {}>(uploadTransactionsModal);
}
