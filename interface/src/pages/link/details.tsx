import type React from 'react';
import { useCallback } from 'react';
import type { AxiosError } from 'axios';
import type { FormikHelpers } from 'formik';
import { ChevronRight, Landmark, Save, Trash } from 'lucide-react';
import { useSnackbar } from 'notistack';
import { Link, useParams } from 'react-router-dom';

import Badge from '@monetr/interface/components/Badge';
import { Button } from '@monetr/interface/components/Button';
import Divider from '@monetr/interface/components/Divider';
import FormButton from '@monetr/interface/components/FormButton';
import FormTextField from '@monetr/interface/components/FormTextField';
import MForm from '@monetr/interface/components/MForm';
import MSpan from '@monetr/interface/components/MSpan';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import { useBankAccountsForLink } from '@monetr/interface/hooks/useBankAccountsForLink';
import { useLink } from '@monetr/interface/hooks/useLink';
import { usePatchLink } from '@monetr/interface/hooks/usePatchLink';
import { showRemoveLinkModal } from '@monetr/interface/modals/RemoveLinkModal';
import type BankAccount from '@monetr/interface/models/BankAccount';
import capitalize from '@monetr/interface/util/capitalize';
import type { APIError } from '@monetr/interface/util/request';

interface LinkValues {
  institutionName: string;
}

export default function LinkDetails(): React.JSX.Element {
  const { enqueueSnackbar } = useSnackbar();
  const { linkId } = useParams();
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
        .catch((error: AxiosError<APIError>) =>
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
    showRemoveLinkModal({ link: link });
  }, [link]);

  if (linkIsLoading || bankAccountsLoading) {
    return (
      <div className='w-full h-full flex items-center justify-center flex-col gap-2'>
        <MSpan className='text-5xl'>One moment...</MSpan>
      </div>
    );
  }

  const initialValues: LinkValues = {
    institutionName: link.institutionName,
  };

  return (
    <MForm className='flex w-full h-full flex-col' initialValues={initialValues} onSubmit={submit}>
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
      <div className='w-full h-full overflow-y-auto min-w-0 p-4 pb-16 md:pb-4'>
        <div className='flex flex-col md:flex-row w-full gap-8 items-center md:items-stretch'>
          <div className='w-full md:w-1/2 flex flex-col items-center'>
            <MSpan className='text-xl my-2 w-full'>Details</MSpan>
            <FormTextField
              className='w-full'
              data-1p-ignore
              label='Instituion / Budget Name'
              name='institutionName'
              placeholder='Budget Name'
              required
            />
          </div>
          <Divider className='block md:hidden w-1/2' />
          <div className='w-full md:w-1/2 flex flex-col gap-2'>
            <MSpan className='text-xl my-2'>Accounts</MSpan>
            <ul className='flex flex-col gap-2'>
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
    <li className='group relative w-full'>
      <Link
        className='group flex h-full gap-1 rounded-lg px-2 py-1 group-hover:bg-zinc-600 md:gap-4 items-center'
        to={path}
      >
        <div className='flex min-w-0 flex-col overflow-hidden grow'>
          <div className='flex gap-2'>
            <MSpan className='group-hover:underline' color='emphasis' ellipsis size='md' weight='semibold'>
              {props.bankAccount.name}
            </MSpan>
            {Boolean(props.bankAccount.deletedAt) && <Badge size='sm'>Archived</Badge>}
          </div>
          <MSpan color='default' ellipsis size='sm' weight='medium'>
            {capitalize(props.bankAccount.accountSubType)}
          </MSpan>
        </div>
        <ChevronRight className='text-dark-monetr-content-subtle group-hover:text-dark-monetr-content-emphasis' />
      </Link>
    </li>
  );
}
