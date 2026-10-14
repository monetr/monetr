import { Fragment } from 'react';
import { format } from 'date-fns';
import { ArrowUpRight, Book, BookOpen, Globe, Info, LinkIcon, Mail } from 'lucide-react';

import { Avatar, AvatarFallback } from '@monetr/interface/components/Avatar';
import MTopNavigation from '@monetr/interface/components/MTopNavigation';
import SettingsRow from '@monetr/interface/components/settings/SettingsRow';
import SettingsSection from '@monetr/interface/components/settings/SettingsSection';
import Typography from '@monetr/interface/components/Typography';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';

import styles from './about.module.scss';

export default function SettingsAbout(): React.JSX.Element {
  const { data: config } = useAppConfiguration();
  const { release, revision, buildType, buildTime } = config ?? {};

  return (
    <Fragment>
      <MTopNavigation icon={Info} title='About' />
      <div className={styles.root}>
        <SettingsSection description='What this server is running.' title='About monetr'>
          <SettingsRow label='Version'>
            <Typography size='sm' weight='medium'>
              {release ? release : 'Unknown'}
            </Typography>
          </SettingsRow>
          <SettingsRow label='Revision'>
            <Typography component='code' size='sm'>
              {revision ? revision.slice(0, 7) : 'Unknown'}
            </Typography>
          </SettingsRow>
          <SettingsRow label='Build Type'>
            <Typography size='sm' weight='medium'>
              {buildType || 'Unknown'}
            </Typography>
          </SettingsRow>
          <SettingsRow label='Build Time'>
            <Typography size='sm' weight='medium'>
              {buildTime ? format(buildTime, 'LLLL do yyyy, h:mmaaa OOOO') : 'Unknown'}
            </Typography>
          </SettingsRow>
        </SettingsSection>

        <SettingsSection title='Need Help?'>
          <div className={styles.links}>
            <AboutLink description='monetr.app' href='https://monetr.app' icon={Globe} title='Website' />
            <AboutLink
              description='Learn how to use monetr'
              href='https://monetr.app/documentation'
              icon={BookOpen}
              title='Documentation'
            />
            <AboutLink
              description='github.com/monetr/monetr'
              href='https://github.com/monetr/monetr'
              icon={Book}
              title='Source Code'
            />
            <AboutLink description='support@monetr.app' href='mailto:support@monetr.app' icon={Mail} title='Email' />
            <AboutLink
              description='Ask questions, share ideas'
              href='https://github.com/monetr/monetr/discussions'
              icon={LinkIcon}
              title='GitHub Discussions'
            />
            <AboutLink
              description='Join Discord Server'
              href='https://discord.gg/68wTCXrhuq'
              icon={LinkIcon}
              title='Discord'
            />
          </div>
        </SettingsSection>

        <SettingsSection title='Legal'>
          <div className={styles.links}>
            <AboutLink href='https://monetr.app/policy/terms' title='Terms And Conditions' />
            <AboutLink href='https://monetr.app/policy/privacy' title='Privacy Policy' />
          </div>
        </SettingsSection>
      </div>
    </Fragment>
  );
}

interface AboutLinkProps {
  href: string;
  title: string;
  description?: string;
  icon?: React.FC;
}

function AboutLink(props: AboutLinkProps): React.JSX.Element {
  const Icon = props.icon;

  return (
    <a className={styles.link} href={props.href} rel='noopener noreferrer' target='_blank'>
      {Icon && (
        <Avatar>
          <AvatarFallback>
            <Icon />
          </AvatarFallback>
        </Avatar>
      )}
      <span className={styles.linkText}>
        <Typography color='emphasis' ellipsis size='sm' weight='semibold'>
          {props.title}
        </Typography>
        {Boolean(props.description) && (
          <Typography color='subtle' ellipsis size='xs'>
            {props.description}
          </Typography>
        )}
      </span>
      <ArrowUpRight className={styles.linkArrow} />
    </a>
  );
}
