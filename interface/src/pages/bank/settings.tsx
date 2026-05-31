import { useCallback } from 'react';
import type { FormikHelpers } from 'formik';
import { Archive, FlaskConical, HeartCrack, Save, Settings } from 'lucide-react';
import { useLocation } from 'wouter';

import type { ApiError } from '@monetr/interface/api/client';
import { Button } from '@monetr/interface/components/Button';
import Card from '@monetr/interface/components/Card';
import FormTextField from '@monetr/interface/components/FormTextField';
import MForm from '@monetr/interface/components/MForm';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import SelectCurrency from '@monetr/interface/components/SelectCurrency';
import Typography from '@monetr/interface/components/Typography';
import { useArchiveBankAccount } from '@monetr/interface/hooks/useArchiveBankAccount';
import { useCurrentLink } from '@monetr/interface/hooks/useCurrentLink';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import { useUpdateBankAccount } from '@monetr/interface/hooks/useUpdateBankAccount';
import type { APIError } from '@monetr/interface/util/request';
import { useSnackbar } from '@monetr/notify';

import styles from './settings.module.scss';

interface BankAccountValues {
  name: string;
  currency: string;
}

export default function BankAccountSettingsPage(): JSX.Element | null {
  const { data: link } = useCurrentLink();
  const { data: bankAccount, isLoading, isError } = useSelectedBankAccount();
  const updateBankAccount = useUpdateBankAccount();
  const archiveBankAccount = useArchiveBankAccount();
  const { enqueueSnackbar } = useSnackbar();
  const [, navigate] = useLocation();

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
      <div className={styles.centerState}>
        <Typography size='5xl'>One moment...</Typography>
      </div>
    );
  }

  if (isError) {
    return (
      <div className={styles.centerState}>
        <HeartCrack className={styles.errorIcon} />
        <Typography size='5xl'>Something isn't right...</Typography>
        <Typography size='2xl'>We weren't able to load details for the bank account specified...</Typography>
      </div>
    );
  }

  // By this point we are neither loading nor in an error state, so we should have a bank account. Guard anyway to keep
  // things type safe before we start reading fields off of it.
  if (!bankAccount) {
    return null;
  }

  async function submit(values: BankAccountValues, helpers: FormikHelpers<BankAccountValues>) {
    if (!bankAccount) {
      return Promise.resolve();
    }

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
      .catch((error: ApiError<APIError>) =>
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
    <MForm className={styles.form} initialValues={initialValues} onSubmit={submit}>
      <MTopNavigation
        base={`/bank/${bankAccount.bankAccountId}/transactions`}
        breadcrumb='Settings'
        icon={Settings}
        title={bankAccount.name}
      >
        {!bankAccount.deletedAt && Boolean(link?.getIsManual()) && (
          <Button onClick={archive} variant='destructive'>
            <Archive />
            Archive
          </Button>
        )}
        <Button type='submit' variant='primary'>
          <Save />
          Save Changes
        </Button>
      </MTopNavigation>
      <div className={styles.content}>
        <div className={styles.row}>
          <div className={styles.column}>
            <Card className={styles.card}>
              <Typography size='inherit'>
                <FlaskConical className={styles.cardIcon} />
                This page is still a work in progress, however it has been made available to make it possible to switch
                the currencies for your bank account sooner. This page will be changed over the next several releases to
                improve the UX and functionality.
              </Typography>
            </Card>
            <FormTextField
              className={styles.input}
              data-1p-ignore
              label='Name'
              name='name'
              placeholder='Bank account name...'
            />
            <SelectCurrency className={styles.input} disabled={link?.getIsPlaid()} name='currency' />
          </div>
        </div>
      </div>
    </MForm>
  );
}
