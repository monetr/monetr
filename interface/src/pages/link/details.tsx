import type React from 'react';
import { Fragment, useCallback } from 'react';
import type { FormikHelpers } from 'formik';
import { Landmark, LogIn, Plug, Save, Trash } from 'lucide-react';
import { useParams } from 'wouter';

import type { ApiError } from '@monetr/interface/api/client';
import { Button } from '@monetr/interface/components/Button';
import FormButton from '@monetr/interface/components/FormButton';
import FormTextField from '@monetr/interface/components/FormTextField';
import { layoutVariants } from '@monetr/interface/components/Layout';
import LinkAccountList from '@monetr/interface/components/link/LinkAccountList';
import LinkAlerts from '@monetr/interface/components/link/LinkAlerts';
import LinkStatusCard from '@monetr/interface/components/link/LinkStatusCard';
import MForm from '@monetr/interface/components/MForm';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import Typography from '@monetr/interface/components/Typography';
import { useBankAccountsForLink } from '@monetr/interface/hooks/useBankAccountsForLink';
import { useLink } from '@monetr/interface/hooks/useLink';
import { usePatchLink } from '@monetr/interface/hooks/usePatchLink';
import { showRemoveLinkModal } from '@monetr/interface/modals/RemoveLinkModal';
import { showUpdatePlaidAccountOverlay } from '@monetr/interface/modals/UpdatePlaidAccountOverlay';
import type { ID } from '@monetr/interface/models/ID';
import type LinkModel from '@monetr/interface/models/Link';
import type { APIError } from '@monetr/interface/util/request';
import { useSnackbar } from '@monetr/notify';

import styles from './details.module.scss';

interface LinkValues {
  institutionName: string;
  description: string;
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

  const handleReauthenticateLink = useCallback(() => {
    if (!link) {
      return;
    }
    showUpdatePlaidAccountOverlay({
      link: link,
    });
  }, [link]);

  const handleUpdateAccountSelection = useCallback(() => {
    if (!link) {
      return;
    }
    showUpdatePlaidAccountOverlay({
      link: link,
      updateAccountSelection: true,
    });
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
    description: link.description ?? '',
  };

  return (
    <MForm className={styles.form} initialValues={initialValues} onSubmit={submit}>
      <MTopNavigation icon={Landmark} title={link.getName()} />
      <div className={styles.body}>
        <div className={styles.content}>
          <LinkAlerts link={link} />
          <LinkStatusCard link={link} />
          <div className={styles.columns}>
            <div className={styles.columnDetails}>
              <Typography className={styles.heading} color='emphasis' component='h3' size='xl' weight='semibold'>
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
              <FormTextField
                className={layoutVariants({ width: 'full' })}
                data-1p-ignore
                label='Description'
                name='description'
                placeholder='Optional notes, like whose login this is'
              />
              <div className={styles.saveDesktop}>
                <FormButton role='form' type='submit' variant='primary'>
                  <Save />
                  Save Changes
                </FormButton>
              </div>
              {link.getIsPlaid() && (
                <Fragment>
                  <Typography className={styles.connectionHeading} color='emphasis' component='h3' weight='semibold'>
                    Connection
                  </Typography>
                  <dl className={styles.connection}>
                    <dt>Provider</dt>
                    <dd>
                      <Typography className={styles.connectionValue} component='code'>
                        Plaid
                      </Typography>
                    </dd>
                    <dt>Institution</dt>
                    <dd>
                      <Typography className={styles.connectionValue} component='code' ellipsis>
                        {link.plaidLink?.institutionName}
                      </Typography>
                    </dd>
                  </dl>
                  <div className={styles.connectionButtons}>
                    <Button onClick={handleUpdateAccountSelection} variant='outlined'>
                      <Plug />
                      Update Account Selection
                    </Button>
                    {(link.getIsError() || link.getIsPendingExpiration()) && (
                      <Button onClick={handleReauthenticateLink} variant='outlined'>
                        <LogIn />
                        Reauthenticate
                      </Button>
                    )}
                  </div>
                </Fragment>
              )}
              {link.getIsLunchFlow() && (
                <Fragment>
                  <Typography className={styles.connectionHeading} color='emphasis' component='h3' weight='semibold'>
                    Connection
                  </Typography>
                  <dl className={styles.connection}>
                    <dt>Provider</dt>
                    <dd>
                      <Typography className={styles.connectionValue} component='code'>
                        Lunch Flow
                      </Typography>
                    </dd>
                    <dt>Connection</dt>
                    <dd>
                      <Typography className={styles.connectionValue} component='code' ellipsis>
                        {link.lunchFlowLink?.name}
                      </Typography>
                    </dd>
                    <dt>API URL</dt>
                    <dd>
                      <Typography className={styles.connectionValue} component='code' ellipsis>
                        {link.lunchFlowLink?.apiUrl}
                      </Typography>
                    </dd>
                  </dl>
                </Fragment>
              )}
            </div>
            <div className={styles.columnAccounts}>
              <LinkAccountList bankAccounts={bankAccounts} link={link} />
            </div>
          </div>
          <section className={styles.remove}>
            <div className={styles.removeText}>
              <Typography color='emphasis' component='h3' size='lg' weight='semibold'>
                Remove {link.getName()}
              </Typography>
              <Typography color='subtle' component='p'>
                All expenses, goals and transactions related to {link.getName()} will be deleted. This cannot be undone.
              </Typography>
            </div>
            <Button onClick={handleRemoveLink} variant='destructive'>
              <Trash />
              Remove {link.getName()}
            </Button>
          </section>
        </div>
      </div>
      <div className={styles.saveMobile}>
        <FormButton className={styles.saveMobileButton} role='form' type='submit' variant='primary'>
          <Save />
          Save Changes
        </FormButton>
      </div>
    </MForm>
  );
}
