import React from 'react';
import { useParams } from 'react-router-dom';
import { Receipt, Save, Trash } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import FormButton from '@monetr/interface/components/FormButton';
import MSpan from '@monetr/interface/components/MSpan';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import { useLink } from '@monetr/interface/hooks/links';

export default function LinkDetails(): React.JSX.Element {
  const { linkId } = useParams();
  const { data: link, isLoading } = useLink(linkId);

  if (isLoading) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <MSpan className='text-5xl'>
          One moment...
        </MSpan>
      </div>
    );
  }

  return (
    <div className='flex w-full h-full flex-col'>
      <MTopNavigation
        icon={ Receipt }
        title='Expenses'
        breadcrumb={ link.getName() }
      >
        <Button variant='destructive' >
          <Trash />
          Remove
        </Button>
        <FormButton variant='primary' type='submit' role='form'>
          <Save />
          Save
        </FormButton>
      </MTopNavigation>
    </div>
  );

}
