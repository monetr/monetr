import { useCallback } from 'react';
import { Book, KeyRound, Plus, RefreshCcw, ServerCrash } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import Card from '@monetr/interface/components/Card';
import SettingsAPIHeader from '@monetr/interface/components/settings/SettingsAPI/Header';
import Typography from '@monetr/interface/components/Typography';
import useApiKeys from '@monetr/interface/hooks/useApiKeys';

import styles from './api.module.scss';

export default function SettingsAPIKeys(): React.JSX.Element {
  const { data: keys, isLoading, isError, refetch, isFetching } = useApiKeys();

  const refresh = useCallback(() => refetch(), [refetch]);

  if (isLoading) {
    return <div>Loading placeholder </div>;
  }

  if (isError) {
    return (
      <div className={styles.root}>
        <SettingsAPIHeader />
        <Card className={styles.cardRoot}>
          <ServerCrash className={styles.errorLogo} />
          <div className={styles.cardEmptyText}>
            <Typography size='lg' weight='bold'>
              We couldn't load your keys
            </Typography>
            <Typography color='subtle' size='lg' weight='normal'>
              monetr didn't response as expected; your keys are safe and nothing was changed. Try again in just a
              moment.
            </Typography>
          </div>
          <div className={styles.cardEmptyActions}>
            <Button disabled={isFetching} onClick={refresh}>
              <RefreshCcw />
              Try Again
            </Button>
          </div>
        </Card>
      </div>
    );
  }

  return (
    <div className={styles.root}>
      <SettingsAPIHeader />
      <Card className={styles.cardRoot}>
        <div className={styles.keyLogos}>
          <KeyRound className={styles.keyLogosBack} />
          <KeyRound className={styles.keyLogosFront} />
        </div>
        <div className={styles.cardEmptyText}>
          <Typography size='lg' weight='bold'>
            No API Keys yet
          </Typography>
          <Typography color='subtle' size='lg' weight='normal'>
            API Keys let scripts and tools read your monetr data and manage it on your behalf. You'll only see the
            secret once at creation time...
          </Typography>
        </div>
        <div className={styles.cardEmptyActions}>
          <Button variant='outlined'>
            <Book />
            Read the Docs
          </Button>
          <Button>
            <Plus />
            Create your first API Key
          </Button>
        </div>
      </Card>
    </div>
  );
}
