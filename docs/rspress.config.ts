import rehypeMathPostProcess from './plugins/rehypeMathPostProcess';
import trailingSlashRedirects from './plugins/trailingSlashRedirects';

import path from 'node:path';
import { pluginSass } from '@rsbuild/plugin-sass';
import { defineConfig } from '@rspress/core';
import { pluginRss } from '@rspress/plugin-rss';
import { pluginSitemap } from '@rspress/plugin-sitemap';
import { transformerNotationHighlight } from '@shikijs/transformers';
import rehypeKatex from 'rehype-katex';
import remarkMath from 'remark-math';
import readingTimePlugin from 'rspress-plugin-reading-time';

const envName = process.env.NODE_ENV ?? 'development';
const branch = process.env.GIT_BRANCH ?? 'main';
const isDevelopment = envName !== 'production';

// rspress-plugin-reading-time attaches a computed reading time to
// pageData.readingTimeData (a reading-time result: { text, minutes, words, ... }).
// Its default behaviour also injects its own component after the first heading,
// but <LedgerMeta> renders the value itself, so we keep only the data hook.
const readingTimeBase = readingTimePlugin();
const readingTimeDataPlugin = {
  name: readingTimeBase.name,
  extendPageData: readingTimeBase.extendPageData,
};

export default defineConfig({
  outDir: process.env.OUTPUT_DIR ?? 'doc_build',
  root: 'src',
  title: 'monetr',
  description:
    'Take control of your finances, paycheck by paycheck, with monetr. Put aside what you need, spend what you want, and confidently manage your money with ease.',
  logo: '/logo.svg',
  icon: '/favicon.ico',
  lang: 'en',
  locales: [{ lang: 'en', label: 'English' }],
  multiVersion: {
    default: 'v1',
    versions: ['v1'],
  },
  globalStyles: path.join(__dirname, 'styles/globals.scss'),
  themeConfig: {
    darkMode: true,
    socialLinks: [
      {
        icon: 'github',
        mode: 'link',
        content: 'https://github.com/monetr/monetr',
      },
      {
        icon: 'discord',
        mode: 'link',
        content: 'https://discord.gg/68wTCXrhuq',
      },
    ],
    footer: {
      message: '',
    },
    editLink: {
      docRepoBaseUrl: `https://github.com/monetr/monetr/blob/${branch}/docs/src`,
    },
    lastUpdated: true,
  },
  markdown: {
    remarkPlugins: [remarkMath],
    rehypePlugins: [rehypeKatex, rehypeMathPostProcess],
    shiki: {
      langs: [
        {
          name: 'math',
          scopeName: 'source.math',
          patterns: [{ match: '.', name: 'text.math' }],
          // Required for some reason?
          repository: {},
        },
      ],
      transformers: [transformerNotationHighlight()],
    },
    link: {
      checkDeadLinks: true,
    },
    image: {
      checkDeadImages: true,
    },
  },
  plugins: [
    pluginSitemap({ siteUrl: 'https://monetr.app' }),
    pluginRss({
      siteUrl: 'https://monetr.app',
      // Use RSS 2.0; override the default .rss extension to the more familiar .xml so the feed sits at /rss/blog.xml.
      output: { type: 'rss' },
      feed: {
        id: 'blog',
        test: '/blog/',
        title: 'monetr Blog',
        description: 'Blog posts and announcements from monetr.',
        language: 'en',
        output: { filename: 'blog.xml' },
        // TODO: channel <link> defaults to siteUrl ("https://monetr.app"); per RSS 2.0 it should point to the HTML page
        // corresponding to the feed ("https://monetr.app/blog"). Override by adding `link` here.
        // TODO: emit <atom:link rel="self">. The feed lib only does so when `feed: '<self URL>'` is set on the channel
        // options. Without it, validators warn with MissingAtomSelfLink.
        // TODO: <guid> currently renders as the relative routePath with isPermaLink="false". Item.link is already the
        // absolute URL, so we could leave guid unset and let the feed lib reuse link as a permalink guid; or set guid
        // explicitly to the absolute URL.
        // Existing blog frontmatter uses `authors` (plural) with { name, github, email } entries. plugin-rss reads
        // `author` (singular), so map it here instead of touching every blog post. Also drop the rendered HTML body
        // from content:encoded: it would otherwise ship root-relative URLs and rspress-specific markup that feed
        // readers can't resolve cleanly. Subscribers see the summary and click through for the full post.
        item(item, page) {
          const next: typeof item = { ...item, content: '' };
          const raw = page.frontmatter?.authors;
          if (!Array.isArray(raw)) {
            return next;
          }
          const authors = raw.flatMap(entry => {
            if (!entry || typeof entry !== 'object') {
              return [];
            }
            const { name, github, email } = entry as {
              name?: unknown;
              github?: unknown;
              email?: unknown;
            };
            if (typeof name !== 'string' || name.length === 0) {
              return [];
            }
            return [
              {
                name,
                // RSS 2.0 requires an email on <author>; fall back to a noreply address so we emit "email (Name)"
                // rather than just the name.
                email: typeof email === 'string' && email.length > 0 ? email : 'noreply@monetr.app',
                link: typeof github === 'string' ? `https://github.com/${github}` : undefined,
              },
            ];
          });
          return authors.length > 0 ? { ...next, author: authors } : next;
        },
      },
    }),
    trailingSlashRedirects({ siteUrl: 'https://monetr.app' }),
    readingTimeDataPlugin,
  ],
  builderConfig: {
    plugins: [pluginSass()],
    output: {
      cleanDistPath: true,
      cssModules: isDevelopment
        ? undefined
        : {
            // Hash class names in production builds to match interface/ convention.
            localIdentName: '[hash:base64:6]',
          },
    },
    security: {
      sri: {
        enable: !isDevelopment,
        algorithm: 'sha512',
      },
    },
    resolve: {
      alias: {
        '@monetr/docs': path.resolve(__dirname, '.'),
      },
    },
    performance: {
      preload: {
        // Prevents dumb screen flash where the font is missing.
        type: 'all-assets',
        include: [/inter-latin-wght-normal.*\.woff2$/],
      },
    },
    html: {
      tags: [
        {
          tag: 'script',
          // Specify the default theme mode, which can be `dark` or `light`
          children: "window.RSPRESS_THEME = 'dark';",
        },
        {
          // Only include umami if we are doing a production build
          tag: 'script',
          attrs: !isDevelopment
            ? {
                defer: true,
                src: 'https://a.monetr.app/script.js',
                'data-website-id': 'ccbdfaf9-683f-4487-b97f-5516e1353715',
              }
            : {},
        },
      ],
    },
  },
  route: {
    cleanUrls: true,
  },
});
