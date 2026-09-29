import { Fragment, useCallback } from 'react';
import type { FormikHelpers } from 'formik';
import { Archive, HeartCrack, Save, Settings } from 'lucide-react';
import { useLocation } from 'wouter';

import type { ApiError } from '@monetr/interface/api/client';
import { Button } from '@monetr/interface/components/Button';
import FormAmountField from '@monetr/interface/components/FormAmountField';
import FormButton from '@monetr/interface/components/FormButton';
import FormTextField from '@monetr/interface/components/FormTextField';
import MForm from '@monetr/interface/components/MForm';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import SelectBankAccountSubType from '@monetr/interface/components/SelectBankAccountSubType';
import SelectBankAccountType from '@monetr/interface/components/SelectBankAccountType';
import SelectCurrency from '@monetr/interface/components/SelectCurrency';
import Typography from '@monetr/interface/components/Typography';
import { useArchiveBankAccount } from '@monetr/interface/hooks/useArchiveBankAccount';
import { useCurrentLink } from '@monetr/interface/hooks/useCurrentLink';
import useLocaleCurrency from '@monetr/interface/hooks/useLocaleCurrency';
import { useSelectedBankAccount } from '@monetr/interface/hooks/useSelectedBankAccount';
import { useUpdateBankAccount } from '@monetr/interface/hooks/useUpdateBankAccount';
import type { BankAccountSubType, BankAccountType } from '@monetr/interface/models/BankAccount';
import { amountToFriendly, friendlyToAmount } from '@monetr/interface/util/amounts';
import type { APIError } from '@monetr/interface/util/request';
import { useSnackbar } from '@monetr/notify';

import styles from './settings.module.scss';

interface BankAccountValues {
  name: string;
  mask: string;
  currency: string;
  accountType: BankAccountType;
  accountSubType: BankAccountSubType;
  availableBalance: number;
  currentBalance: number;
  limitBalance: number | null;
}

export default function BankAccountSettingsPage(): React.JSX.Element | null {
  const { data: link } = useCurrentLink();
  const { data: bankAccount, isLoading, isError } = useSelectedBankAccount();
  const { data: locale, isLoading: isLocaleLoading, isError: isLocaleError } = useLocaleCurrency();
  const updateBankAccount = useUpdateBankAccount();
  const archiveBankAccount = useArchiveBankAccount();
  const { enqueueSnackbar } = useSnackbar();
  const [, navigate] = useLocation();

  const archive = useCallback(async () => {
    if (!bankAccount) {
      return Promise.resolve();
    }

    if (window.confirm(`Are you sure you want to archive bank account: ${bankAccount.name}`)) {
      return await archiveBankAccount(bankAccount.bankAccountId).then(() => navigate('/'));
    }

    return Promise.resolve();
  }, [bankAccount, archiveBankAccount, navigate]);

  if (isLoading || isLocaleLoading) {
    return (
      <div className={styles.centerState}>
        <Typography size='5xl'>One moment...</Typography>
      </div>
    );
  }

  if (isError || isLocaleError) {
    return (
      <div className={styles.centerState}>
        <HeartCrack className={styles.errorIcon} />
        <Typography size='5xl'>Something isn&apos;t right...</Typography>
        <Typography size='2xl'>We weren&apos;t able to load details for the bank account specified...</Typography>
      </div>
    );
  }

  // By this point we are neither loading nor in an error state, so we should have a bank account. Guard anyway to keep
  // things type safe before we start reading fields off of it.
  if (!bankAccount || !locale) {
    return null;
  }

  async function submit(values: BankAccountValues, helpers: FormikHelpers<BankAccountValues>) {
    if (!bankAccount || !locale) {
      return Promise.resolve();
    }

    helpers.setSubmitting(true);

    return await updateBankAccount({
      bankAccountId: bankAccount.bankAccountId,
      name: values.name,
      ...((link?.getIsManual() || link?.getIsLunchFlow()) && {
        mask: values.mask || null,
      }),
      ...(link?.getIsManual() && {
        currency: values.currency,
        accountType: values.accountType,
        accountSubType: values.accountSubType,
        // We don't use the locale.friendlyToAmount helper here because it doesn't accept a currency argument, since the
        // currency could be changed here AND the balance could be changed; we need to handle for that.
        availableBalance: friendlyToAmount(values.availableBalance, locale.locale, values.currency),
        currentBalance: friendlyToAmount(values.currentBalance, locale.locale, values.currency),
        limitBalance:
          values.limitBalance !== null && values.limitBalance !== undefined
            ? friendlyToAmount(values.limitBalance, locale.locale, values.currency)
            : null,
      }),
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
    mask: bankAccount.mask ?? '',
    currency: bankAccount.currency,
    accountType: bankAccount.accountType,
    accountSubType: bankAccount.accountSubType,
    availableBalance: amountToFriendly(bankAccount.availableBalance, locale.locale, bankAccount.currency),
    currentBalance: amountToFriendly(bankAccount.currentBalance, locale.locale, bankAccount.currency),
    limitBalance: bankAccount.limitBalance
      ? amountToFriendly(bankAccount.limitBalance, locale.locale, bankAccount.currency)
      : null,
  };

  return (
    <MForm className={styles.form} initialValues={initialValues} onSubmit={submit}>
      {({ values: { currency } }) => (
        <Fragment>
          <MTopNavigation
            base={`/bank/${bankAccount.bankAccountId}/transactions`}
            breadcrumb='Settings'
            icon={Settings}
            title={bankAccount.name}
          />
          <div className={styles.content}>
            <div className={styles.row}>
              <div className={styles.column}>
                <div className={styles.inputRow}>
                  <FormTextField
                    className={styles.input}
                    data-1p-ignore
                    label='Name'
                    name='name'
                    placeholder='Bank account name...'
                  />
                  <FormTextField
                    className={styles.maskInput}
                    data-1p-ignore
                    disabled={!link?.getIsManual() && !link?.getIsLunchFlow()}
                    inputMode='numeric'
                    label='Mask'
                    maxLength={4}
                    name='mask'
                    placeholder='1234'
                  />
                </div>
                <SelectCurrency className={styles.input} disabled={link?.getIsPlaid()} name='currency' />
                <div className={styles.inputRow}>
                  <SelectBankAccountType className={styles.input} disabled={!link?.getIsManual()} name='accountType' />
                  <SelectBankAccountSubType
                    className={styles.input}
                    disabled={!link?.getIsManual()}
                    name='accountSubType'
                  />
                </div>
                <FormAmountField
                  className={styles.input}
                  currency={currency}
                  disabled={!link?.getIsManual()}
                  label='Available Balance'
                  name='availableBalance'
                />
                <FormAmountField
                  className={styles.input}
                  currency={currency}
                  disabled={!link?.getIsManual()}
                  label='Current Balance'
                  name='currentBalance'
                />
                <FormAmountField
                  className={styles.input}
                  currency={currency}
                  disabled={!link?.getIsManual()}
                  label='Limit Balance'
                  name='limitBalance'
                  placeholder='No Limit Balance'
                />
                <div className={styles.formButtons}>
                  {Boolean(link?.getCanArchiveBankAccount()) && (
                    <Button disabled={Boolean(bankAccount.deletedAt)} onClick={archive} variant='destructive'>
                      <Archive />
                      Archive
                    </Button>
                  )}
                  <FormButton role='form' type='submit' variant='primary'>
                    <Save />
                    Save Changes
                  </FormButton>
                </div>
              </div>
            </div>
          </div>
        </Fragment>
      )}
    </MForm>
  );
}
