import { Fragment, useEffect } from 'react';

import { useFrontmatter, usePage } from '@rspress/core/runtime';
import { Layout as BasicLayout, Link, FallbackHeading as OriginalFallbackHeading } from '@rspress/core/theme-original';

import 'katex/dist/katex.min.css';

// Self-hosted faces for the docs theme. Inter is the body and heading face,
// JetBrains Mono the code/metadata face.
import '@fontsource-variable/inter';
import '@fontsource/jetbrains-mono/400.css';
import '@fontsource/jetbrains-mono/500.css';
import '@fontsource/jetbrains-mono/600.css';

import FundingBar from '@monetr/docs/components/FundingBar';
import GithubStars from '@monetr/docs/components/GithubStars';
import GradientHeading from '@monetr/docs/components/GradientHeading/GradientHeading';
import LedgerMeta from '@monetr/docs/components/LedgerMeta';
import PageMetadata from '@monetr/docs/components/PageMetadata';
import QueryClientWrapper from '@monetr/docs/components/QueryClientWrapper';
import SignIn from '@monetr/docs/components/SignIn';

import layoutStyles from './Layout.module.scss';

function NavTitle() {
  return (
    <Link className={layoutStyles.navTitle} href='/'>
      <img alt='monetr logo' className={layoutStyles.navLogo} src='/logo.svg' />
      <GradientHeading
        as='span'
        blurClassName={layoutStyles.navBrandBlur}
        foregroundClassName={layoutStyles.navBrandForeground}
        wrapperClassName={layoutStyles.navBrandWrapper}
      >
        monetr
      </GradientHeading>
    </Link>
  );
}

function NavExtras() {
  return (
    <div className={layoutStyles.navExtras}>
      <GithubStars />
      <SignIn />
    </div>
  );
}

function Footer() {
  return (
    <footer className={layoutStyles.footer}>
      <div className={layoutStyles.footerRow}>
        <p className={layoutStyles.footerCopyright}>© {new Date().getFullYear()} monetr LLC.</p>
        <div className={layoutStyles.footerLinks}>
          <Link className={layoutStyles.footerLink} href='https://status.monetr.app/' rel='noreferrer' target='_blank'>
            Status
          </Link>
          <Link className={layoutStyles.footerLink} href='/contact'>
            Contact
          </Link>
          <Link className={layoutStyles.footerLink} href='/policy/terms'>
            Terms & Conditions
          </Link>
          <Link className={layoutStyles.footerLink} href='/policy/privacy'>
            Privacy
          </Link>
        </div>
      </div>
    </footer>
  );
}

// DocMeta renders the LedgerMeta row at the top of documentation pages. It is
// wired into the doc layout's beforeDocContent slot so pages get the row without
// any per-MDX additions. Scoped to the /documentation tree so policy/blog/custom
// pages are left alone.
function DocMeta() {
  const { page } = usePage();
  if (!page.routePath?.includes('/documentation')) {
    return null;
  }

  return <LedgerMeta />;
}

const Layout = () => {
  useEffect(() => {
    // Ensure dark mode classes are present
    document.documentElement.classList.add('dark', 'rp-dark');
    document.documentElement.style.colorScheme = 'dark';
  }, []);

  return (
    <Fragment>
      <FundingBar />
      <PageMetadata />
      <QueryClientWrapper>
        <BasicLayout
          afterNavMenu={<NavExtras />}
          // Renders the page-metadata ledger row at the top of doc pages without
          // touching any MDX. DocMeta gates this to the /documentation tree.
          beforeDocContent={<DocMeta />}
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
          // Registers <LedgerMeta> (and anything else here) as a global MDX
          // component. The top-level Layout threads `components` down to the
          // MDXProvider that wraps page content, so docs can use it import-free.
          components={{ LedgerMeta }}
          navTitle={<NavTitle />}
        />
      </QueryClientWrapper>
    </Fragment>
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
