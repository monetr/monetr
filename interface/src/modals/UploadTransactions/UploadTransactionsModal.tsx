import React, { FormEvent, useCallback, useRef, useState } from 'react';
import { useDropzone } from 'react-dropzone';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { Close, FilePresentOutlined, UploadFileOutlined } from '@mui/icons-material';
import { AxiosProgressEvent } from 'axios';
import { useSnackbar } from 'notistack';

import { MBaseButton } from '@monetr/interface/components/MButton';
import MModal, { MModalRef } from '@monetr/interface/components/MModal';
import MSpan from '@monetr/interface/components/MSpan';
import { useSelectedBankAccountId } from '@monetr/interface/hooks/bankAccounts';
import MonetrFile from '@monetr/interface/models/File';
import fileSize from '@monetr/interface/util/fileSize';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';
import request from '@monetr/interface/util/request';
import { ExtractProps } from '@monetr/interface/util/typescriptEvils';

enum UploadTransactionStage {
  FileUpload = 1,
  FieldMapping = 2,
  Processing = 3,
  Completed = 4,
  Error = 5,
}

function UploadTransactionsModal(): JSX.Element {
  const modal = useModal();
  const ref = useRef<MModalRef>(null);

  const [stage, setStage] = useState<UploadTransactionStage>(UploadTransactionStage.FileUpload);

  function CurrentStage(): JSX.Element {
    switch (stage) {
      case UploadTransactionStage.FileUpload:
        return <UploadFileStage setResult={ () => {} } setStage={ setStage } close={ modal.remove } />;
      case UploadTransactionStage.Processing:
        
      default:
        return null;
    }

  }
  
  return (
    <MModal open={ modal.visible } ref={ ref } className='sm:max-w-xl'>
      <CurrentStage />
    </MModal>
  );
}

interface StageProps {
  close: () => void;
  setResult: (file: MonetrFile) => void;
  setStage: (stage: UploadTransactionStage) => void;
}

function UploadFileStage(props: StageProps) {
  const { enqueueSnackbar } = useSnackbar();
  const [file, setFile] = useState<File|null>(null);
  const selectedBankAccountId = useSelectedBankAccountId();
  const [uploadProgress, setUploadProgress] = useState(-1);
  const onDrop = useCallback((acceptedFiles: Array<File>) => {
    const selectedFile = acceptedFiles[0];
    setFile(selectedFile);
  }, []);
  const { getRootProps, getInputProps, isDragActive } = useDropzone({ onDrop });

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();

    const formData = new FormData();
    formData.append('data', file, file.name);

    const config = {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: function (progressEvent: AxiosProgressEvent) {
        const percentCompleted = Math.round((progressEvent.loaded * 100) / progressEvent.total);
        setUploadProgress(percentCompleted);
      },
    };
    setUploadProgress(0);

    return request()
      .post(`/bank_accounts/${ selectedBankAccountId }/files`, formData, config)
      .then(result => {
        setTimeout(() => {
          switch (file.type) {
            case 'text/csv':
              props.setStage(UploadTransactionStage.FieldMapping);
              break;
            default:
              props.setStage(UploadTransactionStage.Processing);
              break;
          }
          props.setResult(new MonetrFile(result.data));
        }, 1000);
      })
      .catch(error => {
        console.error('file upload failed', error);
        // props.setStage(UploadTransactionStage.Error);
        const message = error.response.data.error || 'Unkown error';
        enqueueSnackbar(`Failed to upload file: ${message}`, {
          variant: 'error',
          disableWindowBlurListener: true,
        });
        props.close();
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
            <div>
              { /* TODO Close button */ }
            </div>
          </div>

          <div className='flex gap-2 items-center border rounded-md w-full p-2 border-dark-monetr-border'>
            <FilePresentOutlined className='text-6xl text-dark-monetr-content' />
            <div className='flex flex-col py-1 w-full'>
              <MSpan size='lg'>{ file.name }</MSpan>
              <div className='w-full bg-gray-200 rounded-full h-1.5 my-2 dark:bg-gray-700 relative'>
                <div className='absolute top-0 bg-green-600 h-1.5 rounded-full dark:bg-green-600' style={ { width: `${uploadProgress}%` } }></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (file) {
    return (
      <form onSubmit={ handleSubmit } className='h-full flex flex-col gap-2 p-2 justify-between'>
        <div className='flex flex-col gap-2 h-full'>
          <div className='flex justify-between'>
            <MSpan weight='bold' size='xl'>
            Upload Transactions
            </MSpan>
            <div>
              { /* TODO Close button */ }
            </div>
          </div>
          <MSpan>
            Upload a QFX to import transaction data manually into your account. Maximum of 5MB.
          </MSpan>

          <div className='flex gap-2 items-center border rounded-md w-full p-2 border-dark-monetr-border'>
            <FilePresentOutlined className='text-6xl text-dark-monetr-content' />
            <div className='flex flex-col py-1 w-full'>
              <MSpan size='lg'>{ file.name }</MSpan>
              <MSpan>{ fileSize(file.size) }</MSpan>
            </div>
            <Close className='mr-2 text-dark-monetr-content-subtle hover:text-dark-monetr-content cursor-pointer' />
          </div>
        </div>
        <div className='flex justify-end gap-2 mt-2'>
          <MBaseButton color='secondary' onClick={ props.close }>
            Cancel
          </MBaseButton>
          <MBaseButton color='primary' type='submit'>
            Upload
          </MBaseButton>
        </div>
      </form>
    );
  }

  return (
    <form onSubmit={ handleSubmit } className='h-full flex flex-col gap-2 p-2 justify-between'>
      <div className='flex flex-col gap-2 h-full'>
        <div className='flex justify-between'>
          <MSpan weight='bold' size='xl'>
            Upload Transactions
          </MSpan>
          <div>
            { /* TODO Close button */ }
          </div>
        </div>
        <MSpan>
          Upload a QFX to import transaction data manually into your account. Maximum of 5MB.
        </MSpan>

        <div { ...getRootProps() } className={ uploadClassNames }>
          <input { ...getInputProps() } />
          <UploadFileOutlined className='text-6xl text-dark-monetr-content' />
          <MSpan size='lg' weight='semibold'>Drag QFX file here</MSpan>
          <MSpan>Or click to browse</MSpan>
        </div>
      </div>
      <div className='flex justify-end gap-2 mt-2'>
        <MBaseButton color='secondary' onClick={ props.close }>
          Cancel
        </MBaseButton>
        <MBaseButton color='primary' type='submit'>
          Upload
        </MBaseButton>
      </div>
    </form>
  );
}

const uploadTransactionsModal = NiceModal.create(UploadTransactionsModal);

export default uploadTransactionsModal;

export function showUploadTransactionsModal(): Promise<void> {
  return NiceModal.show<void, ExtractProps<typeof uploadTransactionsModal>, {}>(uploadTransactionsModal);
}
