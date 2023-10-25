import React, { useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import NiceModal, { useModal } from '@ebay/nice-modal-react';
import { DeleteOutlined } from '@mui/icons-material';
import { useSnackbar } from 'notistack';

import { MBaseButton } from 'components/MButton';
import MModal, { MModalRef } from 'components/MModal';
import MSpan from 'components/MSpan';
import { useRemoveLink } from 'hooks/links';
import Link from 'models/Link';
import { ExtractProps } from 'util/typescriptEvils';

export interface RemoveLinkModalProps {
  link: Link;
}

function RemoveLinkModal(props: RemoveLinkModalProps): JSX.Element {
  const modal = useModal();
  const ref = useRef<MModalRef>(null);
  const { enqueueSnackbar } = useSnackbar();
  const [submitting, setSubmitting] = useState(false);
  const removeLink = useRemoveLink();
  const navigate = useNavigate();

  async function submit() {
    setSubmitting(true);
    return removeLink(props.link.linkId)
      .then(() => {
        navigate('/');
        modal.remove();
      })
      .catch(error => {
        setSubmitting(false);
        enqueueSnackbar(error?.response?.data?.error || `Failed to remove ${ props.link.getName() }`, {
          variant: 'error',
          disableWindowBlurListener: true,
        });
      });
  }

  return (
    <MModal open={ modal.visible } ref={ ref } className='py-4 md:max-w-md'>
      <div className='h-full flex flex-col gap-4 p-2 justify-between'>
        <div className='flex flex-col'>
          <MSpan weight='bold' size='xl' className='mb-2'>
            <DeleteOutlined />
            Remove { props.link.getName() }?
          </MSpan>
          <MSpan size='lg' weight='medium'>
            Are you sure you want to remove your {props.link.getName()} data?
          </MSpan>
          <MSpan size='lg'>
            All expenses, goals and transactions related to this will be deleted. This cannot be undone.
          </MSpan>
        </div>
        <div className='flex justify-end gap-2'>
          <MBaseButton disabled={ submitting } color='secondary' onClick={ modal.remove }>
            Cancel
          </MBaseButton>
          <MBaseButton disabled={ submitting } color='cancel' type='submit' onClick={ submit }>
            Remove
          </MBaseButton>
        </div>
      </div>
    </MModal>
  );
}

const removeLinkModal = NiceModal.create<RemoveLinkModalProps>(RemoveLinkModal);

export default removeLinkModal;

export function showRemoveLinkModal(props: RemoveLinkModalProps): Promise<void> {
  return NiceModal.show<void, ExtractProps<typeof removeLinkModal>, {}>(removeLinkModal, props);
}
