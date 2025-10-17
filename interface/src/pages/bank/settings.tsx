import React, { useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import type { AxiosError } from 'axios';
import type { FormikHelpers } from 'formik';
import { Archive, FlaskConical, HeartCrack, Save, Settings } from 'lucide-react';
import { useSnackbar } from 'notistack';

import { Button } from '@monetr/interface/components/Button';
import Card from '@monetr/interface/components/Card';
import MForm from '@monetr/interface/components/MForm';
import MSpan from '@monetr/interface/components/MSpan';
import MTextField from '@monetr/interface/components/MTextField';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import SelectCurrency from '@monetr/interface/components/SelectCurrency';
import { useArchiveBankAccount } from '@monetr/interface/hooks/useArchiveBankAccount';
import { useCurrentLink } from '@monetr/interface/hooks/useCurrentLink';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import { useUpdateBankAccount } from '@monetr/interface/hooks/useUpdateBankAccount';
import type { APIError } from '@monetr/interface/util/request';

interface BankAccountValues {
  name: string;
  currency: string;
}

export default function BankAccountSettingsPage(): JSX.Element {
  const { data: link } = useCurrentLink();
  const { data: bankAccount, isLoading, isError } = useSelectedBankAccount();
  const updateBankAccount = useUpdateBankAccount();
  const archiveBankAccount = useArchiveBankAccount();
  const { enqueueSnackbar } = useSnackbar();
  const navigate = useNavigate();

  const archive = useCallback(async () => {
    if (!bankAccount) {
      return Promise.resolve();
    }

    if (window.confirm(`Are you sure you want to archive bank account: ${bankAccount.name}`)) {
      return archiveBankAccount(bankAccount.bankAccountId).then(() => navigate('/'));
    }

    return Promise.resolve();
  }, [bankAccount, archiveBankAccount, navigate]);

  if (isLoading) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <MSpan className='text-5xl'>One moment...</MSpan>
      </div>
    );
  }

  if (isError) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <HeartCrack className='dark:text-dark-monetr-content h-24 w-24' />
        <MSpan className='text-5xl'>Something isn't right...</MSpan>
        <MSpan className='text-2xl'>We weren't able to load details for the bank account specified...</MSpan>
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
      .then(() =>
        enqueueSnackbar('Updated bank account successfully', {
          variant: 'success',
          disableWindowBlurListener: true,
        }),
      )
      .catch((error: AxiosError<APIError>) =>
        enqueueSnackbar(error?.response?.data?.error || 'Failed to update bank account', {
          variant: 'error',
          disableWindowBlurListener: true,
        }),
      )
      .finally(() => helpers.setSubmitting(false));
  }

  const initialValues: BankAccountValues = {
    name: bankAccount.name,
    currency: bankAccount.currency,
  };

  return (
    <MForm initialValues={initialValues} onSubmit={submit} className='w-full h-full flex flex-col'>
      <MTopNavigation
        icon={Settings}
        title={bankAccount.name}
        base={`/bank/${bankAccount.bankAccountId}/transactions`}
        breadcrumb='Settings'
      >
        {!bankAccount.deletedAt && Boolean(link?.getIsManual()) && (
          <Button variant='destructive' onClick={archive}>
            <Archive />
            Archive
          </Button>
        )}
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
                This page is still a work in progress, however it has been made available to make it possible to switch
                the currencies for your bank account sooner. This page will be changed over the next several releases to
                improve the UX and functionality.
              </MSpan>
            </Card>

            <MTextField
              id='bank-account-name-search'
              label='Name'
              placeholder='Bank account name...'
              name='name'
              className='w-full'
            />

            <SelectCurrency name='currency' className='w-full' disabled={link?.getIsPlaid()} />
          </div>
        </div>
      </div>
    </MForm>
  );
}
