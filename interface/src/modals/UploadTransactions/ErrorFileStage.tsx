import { FileUp } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import Typography from '@monetr/interface/components/Typography';

import styles from './ErrorFileStage.module.scss';

interface ErrorFileStageProps {
  close: () => void;
  error: { message: string; filename: string };
}

export default function ErrorFileStage(props: ErrorFileStageProps): JSX.Element {
  return (
    <div className={styles.root}>
      <div className={styles.body}>
        <div className={styles.header}>
          <Typography size='xl' weight='bold'>
            Upload Transactions
          </Typography>
          <div>{/* TODO Close button */}</div>
        </div>

        <div className={styles.fileCard}>
          <FileUp className={styles.fileIcon} />
          <div className={styles.fileInfo}>
            <Typography size='lg'>{props.error.filename}</Typography>
            <Typography size='inherit'>Failed to import data: {props.error.message}</Typography>
          </div>
        </div>
      </div>
      <div className={styles.actions}>
        <Button onClick={props.close} variant='secondary'>
          Close
        </Button>
      </div>
    </div>
  );
}
