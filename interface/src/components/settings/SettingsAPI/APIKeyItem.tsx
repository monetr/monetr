import { KeyRound, Trash } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import Card from '@monetr/interface/components/Card';
import Code from '@monetr/interface/components/Code';
import { Skeleton } from '@monetr/interface/components/Skeleton';
import Typography from '@monetr/interface/components/Typography';
import { useLocale } from '@monetr/interface/hooks/useLocale';
import useTimezone from '@monetr/interface/hooks/useTimezone';
import { useUser } from '@monetr/interface/hooks/useUser';
import type ApiKey from '@monetr/interface/models/ApiKey';
import { DateLength, formatDate } from '@monetr/interface/util/formatDate';

import styles from './APIKeyItem.module.scss';

interface APIKeyItemProps {
  apiKey: ApiKey;
}

export default function APIKeyItem(props: APIKeyItemProps): React.JSX.Element {
  const { inTimezone } = useTimezone();
  const { data: locale, isLoading: localeIsLoading } = useLocale();
  const { data: createdByUser } = useUser(props.apiKey.createdBy);

  if (localeIsLoading) {
    return (
      <Card className={styles.itemSkeleton}>
        <div className={styles.itemSkeletonContent}>
          <Skeleton className={styles.itemSkeletonName} />
          <Skeleton className={styles.itemSkeletonKeyId} />
          <div className={styles.itemSkeletonMeta}>
            <Skeleton className={styles.itemSkeletonMetaLine} />
            <Skeleton className={styles.itemSkeletonMetaLine} />
          </div>
        </div>
        <Skeleton className={styles.itemSkeletonAction} />
      </Card>
    );
  }

  return (
    <Card className={styles.item} key={props.apiKey.apiKeyId}>
      <div className={styles.itemContent}>
        <Typography size='lg' weight='bold'>
          {props.apiKey.name}
        </Typography>
        <Code className={styles.itemContentKeyId} icon={KeyRound} label='Key ID'>
          {props.apiKey.apiKeyId}
        </Code>
        <div className={styles.itemContentMetadata}>
          <Typography component='p' ellipsis size='sm'>
            Created By: <b>{createdByUser?.name() ?? '...'}</b>
          </Typography>
          <Typography component='p' ellipsis size='sm'>
            Created On: <b>{formatDate(props.apiKey.createdAt, inTimezone, locale!, DateLength.Full)}</b>
          </Typography>
        </div>
      </div>
      <div>
        <Button onClick={() => showRevokeAPIKeyModal({ apiKey: props.apiKey })} variant='destructive'>
          <Trash />
          Revoke
        </Button>
      </div>
    </Card>
  );
}
