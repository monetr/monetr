import Typography from '@monetr/interface/components/Typography';

import styles from './SettingsSection.module.scss';

export interface SettingsSectionProps {
  title: string;
  description?: string;
  children: React.ReactNode;
}

export function SettingsSection(props: SettingsSectionProps): React.JSX.Element {
  return (
    <section className={styles.settingsSection}>
      <div className={styles.settingsSectionHeader}>
        <Typography color='emphasis' size='lg' weight='semibold'>
          {props.title}
        </Typography>
        {Boolean(props.description) && (
          <Typography color='subtle' size='sm'>
            {props.description}
          </Typography>
        )}
      </div>
      <div className={styles.settingsSectionCard}>{props.children}</div>
    </section>
  );
}

export interface SettingsRowProps {
  label: string;
  description?: string;
  children?: React.ReactNode;
}

export function SettingsRow(props: SettingsRowProps): React.JSX.Element {
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
