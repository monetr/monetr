import React from 'react';
import { format } from 'date-fns';

import MDivider from '@monetr/interface/components/MDivider';
import MSpan, { MSpanDeriveClasses } from '@monetr/interface/components/MSpan';
import { ReactElement } from '@monetr/interface/components/types';
import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';

export default function SettingsAbout(): JSX.Element {
  const {
    data: { release, revision, buildType, buildTime },
  } = useAppConfiguration();

  return (
    <div className='w-full flex flex-col p-4 gap-8'>
      <div className='w-full flex flex-col max-w-2xl'>
        <MSpan size='2xl' weight='bold' color='emphasis' className='mb-4'>
          About monetr
        </MSpan>
        <MDivider />
        <div className='flex flex-col md:flex-row p-4 gap-2'>
          <MSpan className='md:w-1/3' size='lg' weight='semibold'>
            Version
          </MSpan>
          <MSpan component='code' size='lg' className='max-w-fit'>
            {release || 'Unknown'}
          </MSpan>
        </div>
        <MDivider />

        <div className='flex flex-col md:flex-row p-4 gap-2'>
          <MSpan className='md:w-1/3' size='lg' weight='semibold'>
            Revision
          </MSpan>
          <MSpan component='code' size='lg' className='max-w-fit'>
            {revision ? revision.slice(0, 7) : 'Unknown'}
          </MSpan>
        </div>
        <MDivider />

        <div className='flex flex-col md:flex-row p-4 gap-2'>
          <MSpan className='md:w-1/3' size='lg' weight='semibold'>
            Build Type
          </MSpan>
          <MSpan component='code' size='lg' className='max-w-fit'>
            {buildType || 'Unknown'}
          </MSpan>
        </div>
        <MDivider />

        <div className='flex flex-col md:flex-row p-4 gap-2'>
          <MSpan className='md:w-1/3' size='lg' weight='semibold'>
            Build Time
          </MSpan>
          <MSpan component='code' size='lg' className='max-w-fit' ellipsis>
            {format(buildTime, 'LLLL do yyyy, h:mmaaa OOOO')}
          </MSpan>
        </div>
        <MDivider />
      </div>

      <div className='w-full flex flex-col max-w-2xl'>
        <MSpan size='2xl' weight='bold' color='emphasis' className='mb-4'>
          Need Help?
        </MSpan>
        <MDivider />

        <div className='flex flex-col md:flex-row p-4 gap-2'>
          <MSpan className='md:w-1/3' size='lg' weight='semibold'>
            Source Code
          </MSpan>
          <AboutHyperlink href='https://github.com/monetr/monetr'>https://github.com/monetr/monetr</AboutHyperlink>
        </div>
        <MDivider />

        <div className='flex flex-col md:flex-row p-4 gap-2'>
          <MSpan className='md:w-1/3' size='lg' weight='semibold'>
            Email
          </MSpan>
          <AboutHyperlink href='mailto:support@monetr.app'>support@monetr.app</AboutHyperlink>
        </div>
        <MDivider />

        <div className='flex flex-col md:flex-row p-4 gap-2'>
          <MSpan className='md:w-1/3' size='lg' weight='semibold'>
            Github Discussions
          </MSpan>
          <AboutHyperlink href='https://github.com/monetr/monetr/discussions'>
            https://github.com/monetr/monetr/discussions
          </AboutHyperlink>
        </div>
        <MDivider />

        <div className='flex flex-col md:flex-row p-4 gap-2'>
          <MSpan className='md:w-1/3' size='lg' weight='semibold'>
            Discord
          </MSpan>
          <AboutHyperlink href='https://discord.gg/68wTCXrhuq'>Join Discord Server</AboutHyperlink>
        </div>
        <MDivider />

        <div className='flex flex-col md:flex-row p-4 gap-2'>
          <MSpan className='md:w-1/3' size='lg' weight='semibold'>
            Terms & Conditions
          </MSpan>
          <AboutHyperlink href='https://monetr.app/policy/terms'>https://monetr.app/policy/terms</AboutHyperlink>
        </div>
        <MDivider />

        <div className='flex flex-col md:flex-row p-4 gap-2'>
          <MSpan className='md:w-1/3' size='lg' weight='semibold'>
            Privacy Policy
          </MSpan>
          <AboutHyperlink href='https://monetr.app/policy/privacy'>https://monetr.app/policy/privacy</AboutHyperlink>
        </div>
        <MDivider />
      </div>
    </div>
  );
}

interface AboutHyperlinkProps {
  href: string;
  children: ReactElement;
}

function AboutHyperlink(props: AboutHyperlinkProps): JSX.Element {
  const className = MSpanDeriveClasses({
    size: 'lg',
    className:
      'block dark:text-dark-monetr-blue hover:underline focus:ring-2 focus:ring-dark-monetr-blue focus:underline text-ellipsis min-w-0 truncate',
  });

  return (
    <a className={className} target='_blank' href={props.href}>
      {props.children}
    </a>
  );
}
