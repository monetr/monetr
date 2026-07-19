import { Book, KeyRound, Plus } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import Card from '@monetr/interface/components/Card';
import Typography from '@monetr/interface/components/Typography';

import styles from './api.module.scss';

export default function SettingsAPIKeys(): React.JSX.Element {
  return (
    <div className={styles.root}>
      <div className={styles.headerRow}>
        <div>
          <Typography color='emphasis' component='h1' size='3xl' weight='semibold'>
            API Keys
          </Typography>
          <Typography size='md' weight='normal'>
            Use API keys to connect monetr to scripts, or other custom automation tools. You can manage the keys created
            for this account.
          </Typography>
        </div>
        <div className={styles.headerRowAction}>
          <Button variant='outlined'>
            <Book />
            API Docs
          </Button>
          <Button>
            <Plus />
            New API Key
          </Button>
        </div>
      </div>
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
