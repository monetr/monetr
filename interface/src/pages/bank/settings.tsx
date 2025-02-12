import React from 'react';
import { AxiosError } from 'axios';
import { FormikHelpers } from 'formik';
import { FlaskConical, HeartCrack, Save, Settings } from 'lucide-react';
import { useSnackbar } from 'notistack';

import { Button } from '@monetr/interface/components/Button';
import Card from '@monetr/interface/components/Card';
import MForm from '@monetr/interface/components/MForm';
import MSpan from '@monetr/interface/components/MSpan';
import MTextField from '@monetr/interface/components/MTextField';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import SelectCurrency from '@monetr/interface/components/SelectCurrency';
import { useSelectedBankAccount, useUpdateBankAccount } from '@monetr/interface/hooks/bankAccounts';
import { APIError } from '@monetr/interface/util/request';

interface BankAccountValues {
  name: string;
  currency: string;
}

export default function BankAccountSettingsPage(): JSX.Element {
  const { data: bankAccount, isLoading, isError } = useSelectedBankAccount();
  const updateBankAccount = useUpdateBankAccount();
  const { enqueueSnackbar } = useSnackbar();

  if (isLoading) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <MSpan className='text-5xl'>
          One moment...
        </MSpan>
      </div>
    );
  }

  if (isError) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <HeartCrack className='dark:text-dark-monetr-content h-24 w-24' />
        <MSpan className='text-5xl'>
          Something isn't right...
        </MSpan>
        <MSpan className='text-2xl'>
          We weren't able to load details for the bank account specified...
        </MSpan>
      </div>
    );
  }

  async function submit(values: BankAccountValues, helpers: FormikHelpers<BankAccountValues>) {
    helpers.setSubmitting(true);

    return await updateBankAccount({
      bankAccountId: bankAccount.bankAccountId,
      name: values.name,
      currency: values.currency,
    })
      .then(() => enqueueSnackbar(
        'Updated bank account successfully',
        {
          variant: 'success',
          disableWindowBlurListener: true,
        },
      ))
      .catch((error: AxiosError<APIError>) => enqueueSnackbar(
        error?.response?.data?.error || 'Failed to update bank account',
        {
          variant: 'error',
          disableWindowBlurListener: true,
        },
      ))
      .finally(() => helpers.setSubmitting(false));
  }

  const initialValues: BankAccountValues = {
    name: bankAccount.name,
    currency: bankAccount.currency,
  };

  return (
    <MForm
      initialValues={ initialValues }
      onSubmit={ submit }
      className='w-full h-full flex flex-col'
    >
      <MTopNavigation
        icon={ Settings }
        title={ bankAccount.name }
        base={ `/bank/${bankAccount.bankAccountId}/transactions` }
        breadcrumb='Settings'
      >
        <Button variant='primary' type='submit'>
          <Save />
          Save Changes
        </Button>
      </MTopNavigation>
      <div className='w-full h-full overflow-y-auto min-w-0 p-4'>
        <div className='flex flex-col md:flex-row w-full gap-8 items-center md:items-stretch'>
          <div className='w-full md:w-1/2 flex flex-col items-center'>
            <Card className='w-full mb-4'>
              <MSpan>
                <FlaskConical className='w-16 h-16' />
                This page is still a work in progress, however it has been made available to make it possible to more
                currencies sooner. This page will be changed over the next several releases to improve the UX and
                functionality.
              </MSpan>
            </Card>

            <MTextField
              id='bank-account-name-search'
              label='Name'
              placeholder='Bank account name...'
              name='name'
              className='w-full'
            />

            <SelectCurrency
              name='currency'
              className='w-full'
            />
          </div>
        </div>
      </div>
    </MForm>
  );
}
