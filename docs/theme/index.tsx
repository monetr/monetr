import { useEffect } from 'react';

import GithubStars from '../components/GithubStars';
import QueryClientWrapper from '../components/QueryClientWrapper';
import SignIn from '../components/SignIn';

import { useFrontmatter } from '@rspress/core/runtime';
import { Layout as BasicLayout, Link, FallbackHeading as OriginalFallbackHeading } from '@rspress/core/theme-original';

import './index.css';

function NavTitle() {
  return (
    <Link className='flex items-center gap-3 no-underline hover:brightness-110' href='/'>
      <img alt='monetr logo' className='size-8 lg:size-10 rounded-none' src='/logo.svg' />
      <div className='flex items-center justify-center ml-1'>
        <span className='absolute mx-auto flex border w-fit bg-gradient-to-r blur-xl opacity-75 from-purple-100 via-purple-200 to-purple-300 bg-clip-text text-2xl lg:text-3xl box-content font-extrabold text-transparent text-center select-none'>
          monetr
        </span>
        <span className='relative top-0 justify-center flex bg-gradient-to-r items-center from-purple-100 via-purple-200 to-purple-300 bg-clip-text text-2xl lg:text-3xl font-extrabold text-transparent text-center select-auto'>
          monetr
        </span>
      </div>
    </Link>
  );
}

function NavExtras() {
  return (
    <div className='flex items-center gap-3'>
      <GithubStars />
      <SignIn />
    </div>
  );
}

function Footer() {
  return (
    <footer className='border-t border-zinc-700 py-6 mt-8'>
      <div className='flex w-full items-center sm:items-start justify-between px-6 md:px-20'>
        <p className='text-sm text-zinc-400'>© {new Date().getFullYear()} monetr LLC.</p>
        <div className='gap-2 sm:gap-4 flex flex-col sm:flex-row'>
          <Link
            className='hover:underline text-sm text-zinc-400'
            href='https://status.monetr.app/'
            rel='noreferrer'
            target='_blank'
          >
            Status
          </Link>
          <Link className='hover:underline text-sm text-zinc-400' href='/contact'>
            Contact
          </Link>
          <Link className='hover:underline text-sm text-zinc-400' href='/policy/terms'>
            Terms & Conditions
          </Link>
          <Link className='hover:underline text-sm text-zinc-400' href='/policy/privacy'>
            Privacy
          </Link>
        </div>
      </div>
    </footer>
  );
}

const Layout = () => {
  useEffect(() => {
    // Ensure dark mode classes are present
    document.documentElement.classList.add('dark', 'rp-dark');
    document.documentElement.style.colorScheme = 'dark';
  }, []);

  return (
    <QueryClientWrapper>
      <BasicLayout
        afterNavMenu={<NavExtras />}
        // TODO This renders weird on custom pages, causing a brief flash.
        // beforeNav={
        //   <NoSSR>
        //     <Banner
        //       display={typeof window !== 'undefined'}
        //       href='/blog/2025-12-31-similar-transactions'
        //       message='🎉 Read the latest blog post about similar transactions'
        //       storageKey='monetr-launched-january-2025'
        //     />
        //   </NoSSR>
        // }
        bottom={<Footer />}
        navTitle={<NavTitle />}
      />
    </QueryClientWrapper>
  );
};

// Fixes an issue with headings on blog posts, where H1 would render if one isnt already there, but then the h1 would
// also include things like the tag from the frontmatter which i dont want. So this just makes it so if the frontmatter
// has `noFallbackHeading` then it doesnt do anything at all.
function FallbackHeading(props: { level: 1 | 2 | 3 | 4 | 5 | 6; title: string }) {
  const {
    frontmatter: { noFallbackHeading },
  } = useFrontmatter();

  if (noFallbackHeading) {
    return null;
  }

  return <OriginalFallbackHeading {...props} />;
}

export * from '@rspress/core/theme-original';
export { FallbackHeading, Layout };
