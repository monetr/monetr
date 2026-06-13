import type React from 'react';
import { useCallback } from 'react';
import type { FormikHelpers } from 'formik';
import { ChevronRight, Landmark, Save, Trash } from 'lucide-react';
import { Link, useParams } from 'wouter';

import type { ApiError } from '@monetr/interface/api/client';
import Badge from '@monetr/interface/components/Badge';
import { Button } from '@monetr/interface/components/Button';
import Divider from '@monetr/interface/components/Divider';
import FormButton from '@monetr/interface/components/FormButton';
import FormTextField from '@monetr/interface/components/FormTextField';
import { layoutVariants } from '@monetr/interface/components/Layout';
import MForm from '@monetr/interface/components/MForm';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import Typography from '@monetr/interface/components/Typography';
import { useBankAccountsForLink } from '@monetr/interface/hooks/useBankAccountsForLink';
import { useLink } from '@monetr/interface/hooks/useLink';
import { usePatchLink } from '@monetr/interface/hooks/usePatchLink';
import { showRemoveLinkModal } from '@monetr/interface/modals/RemoveLinkModal';
import type BankAccount from '@monetr/interface/models/BankAccount';
import type { ID } from '@monetr/interface/models/ID';
import type LinkModel from '@monetr/interface/models/Link';
import capitalize from '@monetr/interface/util/capitalize';
import type { APIError } from '@monetr/interface/util/request';
import { useSnackbar } from '@monetr/notify';

import styles from './details.module.scss';

interface LinkValues {
  institutionName: string;
}

export default function LinkDetails(): React.JSX.Element {
  const { enqueueSnackbar } = useSnackbar();
  const { linkId } = useParams<{ linkId: ID<LinkModel> }>();
  const { data: link, isLoading: linkIsLoading } = useLink(linkId);
  const { data: bankAccounts, isLoading: bankAccountsLoading } = useBankAccountsForLink(linkId);
  const patchLink = usePatchLink();

  const submit = useCallback(
    async (values: LinkValues, helpers: FormikHelpers<LinkValues>) => {
      helpers.setSubmitting(true);

      return await patchLink({
        linkId: linkId,
        ...values,
      })
        .then(() =>
          enqueueSnackbar('Updated link successfully', {
            variant: 'success',
            disableWindowBlurListener: true,
          }),
        )
        .catch((error: ApiError<APIError>) =>
          enqueueSnackbar(error?.response?.data?.error || 'Failed to update link', {
            variant: 'error',
            disableWindowBlurListener: true,
          }),
        )
        .finally(() => helpers.setSubmitting(false));
    },
    [enqueueSnackbar, linkId, patchLink],
  );

  const handleRemoveLink = useCallback(() => {
    if (!link) {
      return;
    }
    showRemoveLinkModal({ link: link });
  }, [link]);

  if (linkIsLoading || bankAccountsLoading || !link || !bankAccounts) {
    return (
      <div className={styles.centerState}>
        <Typography size='5xl'>One moment...</Typography>
      </div>
    );
  }

  const initialValues: LinkValues = {
    institutionName: link.institutionName,
  };

  return (
    <MForm className={styles.form} initialValues={initialValues} onSubmit={submit}>
      <MTopNavigation icon={Landmark} title={link.getName()}>
        <Button onClick={handleRemoveLink} variant='destructive'>
          <Trash />
          Remove
        </Button>
        <FormButton role='form' type='submit' variant='primary'>
          <Save />
          Save Changes
        </FormButton>
      </MTopNavigation>
      <div className={styles.body}>
        <div className={styles.columns}>
          <div className={styles.column}>
            <Typography className={styles.headingFull} size='xl'>
              Details
            </Typography>
            <FormTextField
              className={layoutVariants({ width: 'full' })}
              data-1p-ignore
              label='Instituion / Budget Name'
              name='institutionName'
              placeholder='Budget Name'
              required
            />
          </div>
          <Divider className={styles.dividerMobile} />
          <div className={styles.columnAccounts}>
            <Typography className={styles.heading} size='xl'>
              Accounts
            </Typography>
            <ul className={styles.accountList}>
              {bankAccounts.map(account => (
                <BankAccountItem bankAccount={account} key={account.bankAccountId} />
              ))}
            </ul>
          </div>
        </div>
      </div>
    </MForm>
  );
}

interface BankAccountItemProps {
  bankAccount: BankAccount;
}

function BankAccountItem(props: BankAccountItemProps): React.JSX.Element {
  const path = `/bank/${props.bankAccount.bankAccountId}/settings`;
  return (
    <li className={styles.item}>
      <Link className={styles.itemLink} to={path}>
        <div className={styles.itemText}>
          <div className={styles.itemNameRow}>
            <Typography className={styles.itemName} color='emphasis' ellipsis size='md' weight='semibold'>
              {props.bankAccount.name}
            </Typography>
            {Boolean(props.bankAccount.deletedAt) && <Badge size='sm'>Archived</Badge>}
          </div>
          <Typography color='default' ellipsis size='sm' weight='medium'>
            {capitalize(props.bankAccount.accountSubType)}
          </Typography>
        </div>
        <ChevronRight className={styles.itemChevron} />
      </Link>
    </li>
  );
}
