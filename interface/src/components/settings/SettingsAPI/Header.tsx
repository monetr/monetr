import { Book, Plus } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import { showCreateAPIKeyModal } from '@monetr/interface/components/settings/SettingsAPI/CreateAPIKeyModal';
import Typography from '@monetr/interface/components/Typography';

import styles from './Header.module.scss';

export default function Header(): React.JSX.Element {
  return (
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
        <Button onClick={showCreateAPIKeyModal}>
          <Plus />
          New API Key
        </Button>
      </div>
    </div>
  );
}
