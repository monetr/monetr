import { useCallback } from 'react';
import { Book, KeyRound, Plus, RefreshCcw, ServerCrash, Trash } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import Card from '@monetr/interface/components/Card';
import Code from '@monetr/interface/components/Code';
import SettingsAPIHeader from '@monetr/interface/components/settings/SettingsAPI/Header';
import Typography from '@monetr/interface/components/Typography';
import useApiKeys from '@monetr/interface/hooks/useApiKeys';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { DateLength, formatDate } from '@monetr/interface/util/formatDate';

import styles from './api.module.scss';

export default function SettingsAPIKeys(): React.JSX.Element {
  const { data: keys, isLoading, isError, refetch, isFetching, isSuccess } = useApiKeys();
  const { inTimezone } = useTimezone();
  const { data: locale, isLoading: localeIsLoading } = useLocale();

  const refresh = useCallback(() => refetch(), [refetch]);

  if (isLoading || localeIsLoading) {
    return <div>Loading placeholder </div>;
  }

  if (isError || !isSuccess) {
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

  if (keys.length === 0) {
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

  return (
    <div className={styles.root}>
      <SettingsAPIHeader />
      {keys.map(item => (
        <Card className={styles.item} key={item.apiKeyId}>
          <div className={styles.itemContent}>
            <Typography size='lg' weight='bold'>
              {item.name}
            </Typography>
            <Code className={styles.itemContentKeyId} icon={KeyRound} label='Key ID'>
              {item.apiKeyId}
            </Code>
            <div className={styles.itemContentMetadata}>
              <Typography component='p' ellipsis size='sm'>
                Created By: <b>[PLACEHOLDER TODO]</b>
              </Typography>
              <Typography component='p' ellipsis size='sm'>
                Created On: <b>{formatDate(item.createdAt, inTimezone, locale!, DateLength.Full)}</b>
              </Typography>
            </div>
          </div>
          <div>
            <Button variant='destructive'>
              <Trash />
              Revoke
            </Button>
          </div>
        </Card>
      ))}
    </div>
  );
}
