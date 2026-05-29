import { format } from 'date-fns';

import Divider from '@monetr/interface/components/Divider';
import Typography from '@monetr/interface/components/Typography';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';

import styles from './about.module.scss';

export default function SettingsAbout(): JSX.Element {
  const {
    data: { release, revision, buildType, buildTime },
  } = useAppConfiguration();

  return (
    <div className={styles.root}>
      <div className={styles.section}>
        <Typography className={styles.sectionTitle} color='emphasis' size='2xl' weight='bold'>
          About monetr
        </Typography>
        <Divider />
        <div className={styles.row}>
          <Typography className={styles.rowLabel} size='lg' weight='semibold'>
            Version
          </Typography>
          <Typography className={styles.rowValue} component='code' size='lg'>
            {release || 'Unknown'}
          </Typography>
        </div>
        <Divider />

        <div className={styles.row}>
          <Typography className={styles.rowLabel} size='lg' weight='semibold'>
            Revision
          </Typography>
          <Typography className={styles.rowValue} component='code' size='lg'>
            {revision ? revision.slice(0, 7) : 'Unknown'}
          </Typography>
        </div>
        <Divider />

        <div className={styles.row}>
          <Typography className={styles.rowLabel} size='lg' weight='semibold'>
            Build Type
          </Typography>
          <Typography className={styles.rowValue} component='code' size='lg'>
            {buildType || 'Unknown'}
          </Typography>
        </div>
        <Divider />

        <div className={styles.row}>
          <Typography className={styles.rowLabel} size='lg' weight='semibold'>
            Build Time
          </Typography>
          <Typography className={styles.rowValue} component='code' ellipsis size='lg'>
            {format(buildTime, 'LLLL do yyyy, h:mmaaa OOOO')}
          </Typography>
        </div>
        <Divider />
      </div>

      <div className={styles.section}>
        <Typography className={styles.sectionTitle} color='emphasis' size='2xl' weight='bold'>
          Need Help?
        </Typography>
        <Divider />

        <div className={styles.row}>
          <Typography className={styles.rowLabel} size='lg' weight='semibold'>
            Source Code
          </Typography>
          <AboutHyperlink href='https://github.com/monetr/monetr'>https://github.com/monetr/monetr</AboutHyperlink>
        </div>
        <Divider />

        <div className={styles.row}>
          <Typography className={styles.rowLabel} size='lg' weight='semibold'>
            Email
          </Typography>
          <AboutHyperlink href='mailto:support@monetr.app'>support@monetr.app</AboutHyperlink>
        </div>
        <Divider />

        <div className={styles.row}>
          <Typography className={styles.rowLabel} size='lg' weight='semibold'>
            Github Discussions
          </Typography>
          <AboutHyperlink href='https://github.com/monetr/monetr/discussions'>
            https://github.com/monetr/monetr/discussions
          </AboutHyperlink>
        </div>
        <Divider />

        <div className={styles.row}>
          <Typography className={styles.rowLabel} size='lg' weight='semibold'>
            Discord
          </Typography>
          <AboutHyperlink href='https://discord.gg/68wTCXrhuq'>Join Discord Server</AboutHyperlink>
        </div>
        <Divider />

        <div className={styles.row}>
          <Typography className={styles.rowLabel} size='lg' weight='semibold'>
            Terms & Conditions
          </Typography>
          <AboutHyperlink href='https://monetr.app/policy/terms'>https://monetr.app/policy/terms</AboutHyperlink>
        </div>
        <Divider />

        <div className={styles.row}>
          <Typography className={styles.rowLabel} size='lg' weight='semibold'>
            Privacy Policy
          </Typography>
          <AboutHyperlink href='https://monetr.app/policy/privacy'>https://monetr.app/policy/privacy</AboutHyperlink>
        </div>
        <Divider />
      </div>
    </div>
  );
}

interface AboutHyperlinkProps {
  href: string;
  children: React.ReactNode;
}

function AboutHyperlink(props: AboutHyperlinkProps): JSX.Element {
  return (
    <a className={styles.hyperlink} href={props.href} rel='noopener noreferrer' target='_blank'>
      {props.children}
    </a>
  );
}
