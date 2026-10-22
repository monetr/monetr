import { isFuture } from 'date-fns';
import { CreditCard, Info, KeyRound, Shield, User } from 'lucide-react';

import Divider from '@monetr/interface/components/Divider';
import { layoutVariants } from '@monetr/interface/components/Layout';
import SettingsSidebarItem from '@monetr/interface/components/Layout/SettingsSidebarItem';
import Typography from '@monetr/interface/components/Typography';
import useApiKeys from '@monetr/interface/hooks/useApiKeys';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import { useAuthentication } from '@monetr/interface/hooks/useAuthentication';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { DateLength, formatDate } from '@monetr/interface/util/formatDate';

import styles from './SettingsSidebar.module.scss';

export default function SettingsSidebar(): React.JSX.Element {
  const { data: config } = useAppConfiguration();
  const { data: auth } = useAuthentication();
  const { data: apiKeys } = useApiKeys();
  const { data: locale } = useLocale();
  const { inTimezone } = useTimezone();

  let apiKeysCaption = '';
  if (apiKeys?.length === 1) {
    apiKeysCaption = '1 key';
  } else if (apiKeys) {
    apiKeysCaption = `${apiKeys.length} keys`;
  }

  // Same checks as the status badge on the billing page
  let billingCaption = 'Expired';
  if (auth?.hasSubscription && auth?.isActive) {
    billingCaption = 'Active';
  } else if (auth?.trialingUntil && isFuture(auth.trialingUntil)) {
    billingCaption = 'Trialing';
    if (locale) {
      billingCaption = `Trial ends ${formatDate(auth.trialingUntil, inTimezone, locale, DateLength.Medium)}`;
    }
  }

  return (
    <div className={styles.settingsSidebarRoot} data-sidebar-secondary>
      <div className={styles.title}>
        <Typography color='emphasis' ellipsis size='xl' weight='semibold'>
          Settings
        </Typography>
      </div>
      <Divider className={layoutVariants({ width: '1/2' })} />
      <div className={styles.content}>
        <SettingsSidebarItem caption='Profile and account' icon={User} title='Overview' to='/settings/overview' />
        <SettingsSidebarItem
          caption='Password, email, multi-factor'
          icon={Shield}
          title='Security'
          to='/settings/security'
        />
        <SettingsSidebarItem
          caption={apiKeysCaption}
          data-testid='settings-sidebar-api-keys'
          icon={KeyRound}
          title='API Keys'
          to='/settings/api'
        />
        {config?.billingEnabled && (
          <SettingsSidebarItem
            caption={billingCaption}
            data-testid='settings-sidebar-billing'
            icon={CreditCard}
            title='Billing'
            to='/settings/billing'
          />
        )}
        <SettingsSidebarItem caption='Version and server info' icon={Info} title='About' to='/settings/about' />
      </div>
    </div>
  );
}
