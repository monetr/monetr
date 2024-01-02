import React, { FormEvent, useCallback, useRef, useState } from 'react';
import { useDropzone } from 'react-dropzone';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { UploadFileOutlined } from '@mui/icons-material';
import { AxiosProgressEvent } from 'axios';

import { MBaseButton } from '@monetr/interface/components/MButton';
import MModal, { MModalRef } from '@monetr/interface/components/MModal';
import MSpan from '@monetr/interface/components/MSpan';
import { useSelectedBankAccountId } from '@monetr/interface/hooks/bankAccounts';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';
import request from '@monetr/interface/util/request';
import { ExtractProps } from '@monetr/interface/util/typescriptEvils';

function UploadTransactionsModal(): JSX.Element {
  const modal = useModal();
  const selectedBankAccountId = useSelectedBankAccountId();
  const [file, setFile] = useState<File|null>(null);
  const [uploadProgress, setUploadProgress] = useState(0);
  const ref = useRef<MModalRef>(null);
  const onDrop = useCallback((acceptedFiles: Array<File>) => {
    const selectedFile = acceptedFiles[0];
    setFile(selectedFile);
  }, []);
  const { getRootProps, getInputProps, isDragActive } = useDropzone({ onDrop });

  function handleSubmit(event: FormEvent) {
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

    return request()
      .post(`/bank_accounts/${ selectedBankAccountId }/files`, formData, config)
      .then(result => {
        console.log(result);
      });

  }

  const uploadClassNames = mergeTailwind(
    'border-dashed rounded-md w-full border p-8 flex justify-center flex-col items-center cursor-pointer',
    { 'border-dark-monetr-border hover:border-dark-monetr-border-string': !isDragActive },
    { 'border-dark-monetr-brand-subtle': isDragActive },
  );
  
  return (
    <MModal open={ modal.visible } ref={ ref } className='sm:max-w-xl'>
      <form onSubmit={ handleSubmit }  className='h-full flex flex-col gap-2 p-2 justify-between'>
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
            Upload a CSV to import transaction data manually into your account. Maximum of 5MB.
          </MSpan>

          <div { ...getRootProps() } className={ uploadClassNames }>
            <input { ...getInputProps() } />
            <UploadFileOutlined className='text-6xl text-dark-monetr-content' />
            <MSpan size='lg' weight='semibold'>Drag CSV here</MSpan>
            <MSpan>Or click to browse</MSpan>
          </div>
        </div>
        <div className='flex justify-end gap-2'>
          <MBaseButton color='secondary' onClick={ modal.remove }>
            Cancel
          </MBaseButton>
          <MBaseButton color='primary' type='submit'>
            Upload
          </MBaseButton>
        </div>
      </form>
    </MModal>
  );
}

const uploadTransactionsModal = NiceModal.create(UploadTransactionsModal);

export default uploadTransactionsModal;

export function showUploadTransactionsModal(): Promise<void> {
  return NiceModal.show<void, ExtractProps<typeof uploadTransactionsModal>, {}>(uploadTransactionsModal);
}
