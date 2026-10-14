import { Fragment } from 'react';
import { CircleCheck, Copy, User } from 'lucide-react';

import Badge from '@monetr/interface/components/Badge';
import { Button } from '@monetr/interface/components/Button';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import { SettingsRow, SettingsSection } from '@monetr/interface/components/settings/SettingsSection';
import Typography from '@monetr/interface/components/Typography';
import { useAuthentication } from '@monetr/interface/hooks/useAuthentication';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { DateLength, formatDate } from '@monetr/interface/util/formatDate';
import { useSnackbar } from '@monetr/notify';

import styles from './overview.module.scss';

export default function SettingsOverview(): React.JSX.Element {
  const { data: me } = useAuthentication();
  const { data: locale } = useLocale();
  const { inTimezone } = useTimezone();
  const { enqueueSnackbar } = useSnackbar();
  // The clipboard API is only there on secure origins, so a self hosted monetr over plain http won't have it
  const canCopy = typeof navigator.clipboard?.writeText === 'function';

  function copyAccountId() {
    navigator.clipboard
      .writeText(me?.user?.accountId ?? '')
      .then(() =>
        enqueueSnackbar('Copied Account ID to the clipboard', {
          variant: 'success',
          disableWindowBlurListener: true,
        }),
      )
      .catch(() =>
        enqueueSnackbar('Failed to copy to the clipboard.', {
          variant: 'error',
          disableWindowBlurListener: true,
        }),
      );
  }

  return (
    <Fragment>
      <MTopNavigation icon={User} title='Overview' />
      <div className={styles.root}>
        <SettingsSection description='Who you are in monetr.' title='Profile'>
          <SettingsRow description='Used in emails from monetr.' label='Name'>
            <Typography size='sm' weight='medium'>
              {me?.user?.name()}
            </Typography>
          </SettingsRow>
          <SettingsRow description='Where receipts and sign in emails go.' label='Email Address'>
            <Typography size='sm' weight='medium'>
              {me?.user?.login?.email}
            </Typography>
            {me?.user?.login?.getIsEmailVerified() && (
              <Badge size='sm' variant='success'>
                <CircleCheck />
                Verified
              </Badge>
            )}
          </SettingsRow>
        </SettingsSection>

        <SettingsSection description='Shared by everyone on this monetr account.' title='Account'>
          <SettingsRow description='Paydays and due dates happen at midnight in this timezone.' label='Timezone'>
            <Typography size='sm' weight='medium'>
              {me?.user?.account?.timezone}
            </Typography>
          </SettingsRow>
          <SettingsRow description='How dates and amounts are formatted.' label='Locale'>
            <Typography size='sm' weight='medium'>
              {me?.user?.account?.locale}
            </Typography>
          </SettingsRow>
          <SettingsRow
            description="Based on your locale, used when a bank account doesn't have its own currency."
            label='Default Currency'
          >
            <Typography size='sm' weight='medium'>
              {me?.defaultCurrency}
            </Typography>
          </SettingsRow>
          <SettingsRow label='Member Since'>
            <Typography size='sm' weight='medium'>
              {me?.user?.account?.createdAt &&
                locale &&
                formatDate(me.user.account.createdAt, inTimezone, locale, DateLength.Full)}
            </Typography>
          </SettingsRow>
          <SettingsRow description='Handy if you ever need to reach out for support.' label='Account ID'>
            <Typography component='code' size='sm'>
              {me?.user?.accountId}
            </Typography>
            {canCopy && (
              <Button aria-label='Copy Account ID' onClick={copyAccountId} variant='text'>
                <Copy />
              </Button>
            )}
          </SettingsRow>
        </SettingsSection>
      </div>
    </Fragment>
  );
}
