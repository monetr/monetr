import Typography from '@monetr/interface/components/Typography';

import styles from './SettingsRow.module.scss';

export interface SettingsRowProps {
  label: string;
  description?: string;
  children?: React.ReactNode;
}

export default function SettingsRow(props: SettingsRowProps): React.JSX.Element {
  return (
    <div className={styles.settingsRow}>
      <div className={styles.settingsRowText}>
        <Typography color='emphasis' size='sm' weight='semibold'>
          {props.label}
        </Typography>
        {Boolean(props.description) && (
          <Typography color='subtle' size='sm'>
            {props.description}
          </Typography>
        )}
      </div>
      <div className={styles.settingsRowControl}>{props.children}</div>
    </div>
  );
}
