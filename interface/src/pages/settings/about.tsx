import React from 'react';

import MDivider from 'components/MDivider';
import MSpan, { MSpanDeriveClasses } from 'components/MSpan';
import { ReactElement } from 'components/types';
import { format } from 'date-fns';
import { useAppConfiguration } from 'hooks/useAppConfiguration';

export default function SettingsAbout(): JSX.Element {
  const {
    release,
    revision,
    buildType,
    buildTime,
  } = useAppConfiguration();

  return (
    <div className='w-full flex flex-col p-4 gap-8'>
      <div className='w-full flex flex-col max-w-2xl'>
        <MSpan size='2xl' weight='bold' color='emphasis' className='mb-4'>
          About monetr
        </MSpan>
        <MDivider />
        <div className='flex p-4'>
          <MSpan className='w-1/3' size='lg' weight='semibold'>
            Version
          </MSpan>
          <MSpan component='code' size='lg'>
            { release || 'Unknown' }
          </MSpan>
        </div>
        <MDivider />

        <div className='flex p-4'>
          <MSpan className='w-1/3' size='lg' weight='semibold'>
            Revision
          </MSpan>
          <MSpan component='code' size='lg'>
            { revision ? revision.slice(0, 7) : 'Unknown' }
          </MSpan>
        </div>
        <MDivider />

        <div className='flex p-4'>
          <MSpan className='w-1/3' size='lg' weight='semibold'>
            Build Type
          </MSpan>
          <MSpan component='code' size='lg'>
            { buildType || 'Unknown' }
          </MSpan>
        </div>
        <MDivider />

        <div className='flex p-4'>
          <MSpan className='w-1/3' size='lg' weight='semibold'>
            Build Time
          </MSpan>
          <MSpan component='code' size='lg'>
            { format(buildTime, 'LLLL do yyyy, h:mmaaa OOOO') }
          </MSpan>
        </div>
        <MDivider />
      </div>

      <div className='w-full flex flex-col max-w-2xl'>
        <MSpan size='2xl' weight='bold' color='emphasis' className='mb-4'>
          Need Help?
        </MSpan>
        <MDivider />

        <div className='flex p-4'>
          <MSpan className='w-1/3' size='lg' weight='semibold'>
            Source Code
          </MSpan>
          <AboutHyperlink href='https://github.com/monetr/monetr'>
            https://github.com/monetr/monetr
          </AboutHyperlink>
        </div>
        <MDivider />

        <div className='flex p-4'>
          <MSpan className='w-1/3' size='lg' weight='semibold'>
            Email
          </MSpan>
          <AboutHyperlink href='mailto:support@monetr.app'>
            support@monetr.app
          </AboutHyperlink>
        </div>
        <MDivider />

        <div className='flex p-4'>
          <MSpan className='w-1/3' size='lg' weight='semibold'>
            Github Discussions
          </MSpan>
          <AboutHyperlink href='https://github.com/monetr/monetr/discussions'>
            https://github.com/monetr/monetr/discussions
          </AboutHyperlink>
        </div>
        <MDivider />

        <div className='flex p-4'>
          <MSpan className='w-1/3' size='lg' weight='semibold'>
            Discord
          </MSpan>
          <AboutHyperlink href='https://discord.gg/68wTCXrhuq'>
            Join Discord Server
          </AboutHyperlink>
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
    className: 'dark:text-dark-monetr-blue hover:underline focus:ring-2 focus:ring-dark-monetr-blue focus:underline',
  });

  return (
    <a className={ className } target='_blank' href={ props.href }>
      { props.children }
    </a>
  );
}
