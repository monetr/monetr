import React, { useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { FormikHelpers } from 'formik';
import { Landmark, Save, Trash } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import FormButton from '@monetr/interface/components/FormButton';
import MDivider from '@monetr/interface/components/MDivider';
import MForm from '@monetr/interface/components/MForm';
import MSpan from '@monetr/interface/components/MSpan';
import MTextField from '@monetr/interface/components/MTextField';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import { useBankAccountsForLink } from '@monetr/interface/hooks/bankAccounts';
import { useLink } from '@monetr/interface/hooks/links';

interface LinkValues {
  institutionName: string;
}

export default function LinkDetails(): React.JSX.Element {
  const { linkId } = useParams();
  const { data: link, isLoading: linkIsLoading } = useLink(linkId);
  const { data: bankAccounts, isLoading: bankAccountsLoading } = useBankAccountsForLink(linkId);
  const submit = useCallback(async (values: LinkValues, helpers: FormikHelpers<LinkValues>) => {
    console.log(values, helpers);
  }, []);


  if (linkIsLoading || bankAccountsLoading) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <MSpan className='text-5xl'>
          One moment...
        </MSpan>
      </div>
    );
  }

  const initialValues: LinkValues = {
    institutionName: link.institutionName,
  };

  return (
    <MForm
      initialValues={ initialValues }
      className='flex w-full h-full flex-col'
      onSubmit={ submit }
    >
      <MTopNavigation
        icon={ Landmark }
        title={ link.getName() }
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
      <div className='w-full h-full overflow-y-auto min-w-0 p-4 pb-16 md:pb-4'>
        <div className='flex flex-col md:flex-row w-full gap-8 items-center md:items-stretch'>
          <div className='w-full md:w-1/2 flex flex-col items-center'>
            <MSpan className='text-xl my-2 w-full'>
              Details
            </MSpan>
            <MTextField
              className='w-full'
              label='Instituion / Budget Name'
              placeholder='Budget Name'
              name='institutionName'
              required
              data-1p-ignore
            />
          </div>
          <MDivider className='block md:hidden w-1/2' />
          <div className='w-full md:w-1/2 flex flex-col gap-2'>
            <MSpan className='text-xl my-2'>
              Accounts
            </MSpan>
            { bankAccounts.map(account => (
              <div>{account.name}</div>
            ))}
          </div>
        </div>
      </div>
    </MForm>
  );
}
