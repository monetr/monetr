/* eslint-disable max-len */
import React, { Fragment } from 'react';
import Image from 'next/image';
import Link from 'next/link';
import type { DocsThemeConfig } from 'nextra-theme-docs';

import Logo from '@monetr/docs/assets/logo.svg';
import GithubStars from '@monetr/docs/components/GithubStars';
import Head from '@monetr/docs/components/Head';
import SignUp from '@monetr/docs/components/SignUp';

const branch = process.env.GIT_BRANCH ?? 'main';

const config: DocsThemeConfig = {
  head: Head,
  darkMode: false,
  nextThemes: {
    forcedTheme: 'dark',
  },
  logo: (
    <Fragment>
      <Image src={ Logo } alt='monetr logo' className='w-8 h-8 lg:w-10 lg:h-10' />
      <div className='flex items-center justify-center ml-3'>
        <span className='absolute mx-auto flex border w-fit bg-gradient-to-r blur-xl opacity-50 from-purple-100 via-purple-200 to-purple-300 bg-clip-text text-2xl lg:text-3xl box-content font-extrabold text-transparent text-center select-none'>
          monetr
        </span>
        <h1 className='relative top-0 justify-center flex bg-gradient-to-r items-center from-purple-100 via-purple-200 to-purple-300 bg-clip-text text-2xl lg:text-3xl font-extrabold text-transparent text-center select-auto'>
          monetr
        </h1>
      </div>
    </Fragment>
  ),
  docsRepositoryBase: `https://github.com/monetr/monetr/blob/${branch}/docs`,
  banner: {
    dismissible: true,
    key: 'monetr-go-live-january-2025',
    content: (<p>
      🎉 monetr is going live January 3rd, 2025!
      Check out the announcement <Link className='font-bold hover:text-dark-monetr-blue hover:underline' href='/blog/2024-12-30-introduction'>here</Link>.
    </p>
    ),
  },
  project: {
    icon: GithubStars,
    link: 'https://github.com/monetr/monetr',
  },
  navbar: {
    extraContent: <SignUp />,
  },
  sidebar: {
    defaultMenuCollapseLevel: 1,
  },
  chat: {
    link: 'https://discord.gg/68wTCXrhuq',
    icon: (
      <svg xmlns='http://www.w3.org/2000/svg' width='28' height='28' fill='currentColor' viewBox='0 0 16 16' className='hidden sm:block'>
        <path d='M13.545 2.907a13.227 13.227 0 0 0-3.257-1.011.05.05 0 0 0-.052.025c-.141.25-.297.577-.406.833a12.19 12.19 0 0 0-3.658 0 8.258 8.258 0 0 0-.412-.833.051.051 0 0 0-.052-.025c-1.125.194-2.22.534-3.257 1.011a.041.041 0 0 0-.021.018C.356 6.024-.213 9.047.066 12.032c.001.014.01.028.021.037a13.276 13.276 0 0 0 3.995 2.02.05.05 0 0 0 .056-.019c.308-.42.582-.863.818-1.329a.05.05 0 0 0-.01-.059.051.051 0 0 0-.018-.011 8.875 8.875 0 0 1-1.248-.595.05.05 0 0 1-.02-.066.051.051 0 0 1 .015-.019c.084-.063.168-.129.248-.195a.05.05 0 0 1 .051-.007c2.619 1.196 5.454 1.196 8.041 0a.052.052 0 0 1 .053.007c.08.066.164.132.248.195a.051.051 0 0 1-.004.085 8.254 8.254 0 0 1-1.249.594.05.05 0 0 0-.03.03.052.052 0 0 0 .003.041c.24.465.515.909.817 1.329a.05.05 0 0 0 .056.019 13.235 13.235 0 0 0 4.001-2.02.049.049 0 0 0 .021-.037c.334-3.451-.559-6.449-2.366-9.106a.034.034 0 0 0-.02-.019Zm-8.198 7.307c-.789 0-1.438-.724-1.438-1.612 0-.889.637-1.613 1.438-1.613.807 0 1.45.73 1.438 1.613 0 .888-.637 1.612-1.438 1.612Zm5.316 0c-.788 0-1.438-.724-1.438-1.612 0-.889.637-1.613 1.438-1.613.807 0 1.451.73 1.438 1.613 0 .888-.631 1.612-1.438 1.612Z' />
      </svg>
    ),
  },
  footer: {
    content: (
      <div className='flex w-full items-center sm:items-start justify-between md:px-20'>
        <p className='text-sm'>
          © {new Date().getFullYear()} monetr LLC.
        </p>
        <div className='gap-2 sm:gap-4 flex flex-col sm:flex-row'>
          <a href='https://status.monetr.app/' target='_blank' className='hover:underline text-sm'>
            Status
          </a>
          <a href='/contact' className='hover:underline text-sm'>
            Contact
          </a>
          <a href='/policy/terms' className='hover:underline text-sm'>
            Terms & Conditions
          </a>
          <a href='/policy/privacy' className='hover:underline text-sm'>
            Privacy
          </a>
        </div>
      </div>
    ),
  },
  // ... other theme options
};
export default config;
