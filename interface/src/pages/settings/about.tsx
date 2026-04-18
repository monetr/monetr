import { format } from 'date-fns';

import Divider from '@monetr/interface/components/Divider';
import Typography, { textVariants } from '@monetr/interface/components/Typography';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export default function SettingsAbout(): JSX.Element {
  const {
    data: { release, revision, buildType, buildTime },
  } = useAppConfiguration();

  return (
    <div className='w-full flex flex-col p-4 gap-8'>
      <div className='w-full flex flex-col max-w-2xl'>
        <Typography className='mb-4' color='emphasis' size='2xl' weight='bold'>
          About monetr
        </Typography>
        <Divider />
        <div className='flex flex-col md:flex-row p-4 gap-2'>
          <Typography className='md:w-1/3' size='lg' weight='semibold'>
            Version
          </Typography>
          <Typography className='max-w-fit' component='code' size='lg'>
            {release || 'Unknown'}
          </Typography>
        </div>
        <Divider />

        <div className='flex flex-col md:flex-row p-4 gap-2'>
          <Typography className='md:w-1/3' size='lg' weight='semibold'>
            Revision
          </Typography>
          <Typography className='max-w-fit' component='code' size='lg'>
            {revision ? revision.slice(0, 7) : 'Unknown'}
          </Typography>
        </div>
        <Divider />

        <div className='flex flex-col md:flex-row p-4 gap-2'>
          <Typography className='md:w-1/3' size='lg' weight='semibold'>
            Build Type
          </Typography>
          <Typography className='max-w-fit' component='code' size='lg'>
            {buildType || 'Unknown'}
          </Typography>
        </div>
        <Divider />

        <div className='flex flex-col md:flex-row p-4 gap-2'>
          <Typography className='md:w-1/3' size='lg' weight='semibold'>
            Build Time
          </Typography>
          <Typography className='max-w-fit' component='code' ellipsis size='lg'>
            {format(buildTime, 'LLLL do yyyy, h:mmaaa OOOO')}
          </Typography>
        </div>
        <Divider />
      </div>

      <div className='w-full flex flex-col max-w-2xl'>
        <Typography className='mb-4' color='emphasis' size='2xl' weight='bold'>
          Need Help?
        </Typography>
        <Divider />

        <div className='flex flex-col md:flex-row p-4 gap-2'>
          <Typography className='md:w-1/3' size='lg' weight='semibold'>
            Source Code
          </Typography>
          <AboutHyperlink href='https://github.com/monetr/monetr'>https://github.com/monetr/monetr</AboutHyperlink>
        </div>
        <Divider />

        <div className='flex flex-col md:flex-row p-4 gap-2'>
          <Typography className='md:w-1/3' size='lg' weight='semibold'>
            Email
          </Typography>
          <AboutHyperlink href='mailto:support@monetr.app'>support@monetr.app</AboutHyperlink>
        </div>
        <Divider />

        <div className='flex flex-col md:flex-row p-4 gap-2'>
          <Typography className='md:w-1/3' size='lg' weight='semibold'>
            Github Discussions
          </Typography>
          <AboutHyperlink href='https://github.com/monetr/monetr/discussions'>
            https://github.com/monetr/monetr/discussions
          </AboutHyperlink>
        </div>
        <Divider />

        <div className='flex flex-col md:flex-row p-4 gap-2'>
          <Typography className='md:w-1/3' size='lg' weight='semibold'>
            Discord
          </Typography>
          <AboutHyperlink href='https://discord.gg/68wTCXrhuq'>Join Discord Server</AboutHyperlink>
        </div>
        <Divider />

        <div className='flex flex-col md:flex-row p-4 gap-2'>
          <Typography className='md:w-1/3' size='lg' weight='semibold'>
            Terms & Conditions
          </Typography>
          <AboutHyperlink href='https://monetr.app/policy/terms'>https://monetr.app/policy/terms</AboutHyperlink>
        </div>
        <Divider />

        <div className='flex flex-col md:flex-row p-4 gap-2'>
          <Typography className='md:w-1/3' size='lg' weight='semibold'>
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
  const className = mergeTailwind(
    textVariants({ size: 'lg' }),
    'block dark:text-dark-monetr-blue hover:underline focus:ring-2 focus:ring-dark-monetr-blue focus:underline text-ellipsis min-w-0 truncate',
  );

  return (
    <a className={className} href={props.href} rel='noopener' target='_blank'>
      {props.children}
    </a>
  );
}
