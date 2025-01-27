import React from 'react';
import { HeartCrack, Save, Settings } from 'lucide-react';
import { useSnackbar } from 'notistack';

import { Button } from '@monetr/interface/components/Button';
import MForm from '@monetr/interface/components/MForm';
import MSelect from '@monetr/interface/components/MSelect';
import MSpan from '@monetr/interface/components/MSpan';
import MTextField from '@monetr/interface/components/MTextField';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import { useSelectedBankAccount } from '@monetr/interface/hooks/bankAccounts';
import { useInstalledCurrencies } from '@monetr/interface/hooks/useInstalledCurrencies';

interface BankAccountValues {
  name: string;
}


export default function BankAccountSettingsPage(): JSX.Element {
  const { data: bankAccount, isLoading, isError } = useSelectedBankAccount();
  const { data: currencies, isLoading: currenciesLoading } = useInstalledCurrencies();
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

  const initialValues: BankAccountValues = {
    name: bankAccount.name,
  };

  return (
    <MForm
      initialValues={ initialValues }
      onSubmit={ () => {} }
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
            <MTextField
              id='bank-account-name-search'
              label='Name'
              placeholder='Bank account name...'
              name='name'
              className='w-full'
            />

            <MSelect
              label='Currency'
              name='currency'
              onChange={ () => {} }
              options={ (currencies ?? []).map(currency => ({ label: currency, value: currency })) }
              isLoading={ currenciesLoading }
              placeholder='Select a funding schedule...'
              required
              value={ { label: 'USD', value: 'USD' } }
              className='w-full'
            />
          </div>
        </div>
      </div>
    </MForm>
  );
}
