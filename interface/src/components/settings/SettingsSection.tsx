import Typography from '@monetr/interface/components/Typography';

import styles from './SettingsSection.module.scss';

export interface SettingsSectionProps {
  title: string;
  description?: string;
  children: React.ReactNode;
}

export default function SettingsSection(props: SettingsSectionProps): React.JSX.Element {
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
