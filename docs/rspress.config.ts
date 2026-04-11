import rehypeMathPostProcess from './plugins/rehypeMathPostProcess';

import path from 'node:path';
import { pluginSass } from '@rsbuild/plugin-sass';
import { defineConfig } from '@rspress/core';
import { pluginSitemap } from '@rspress/plugin-sitemap';
import { transformerNotationHighlight } from '@shikijs/transformers';
import rehypeKatex from 'rehype-katex';
import remarkMath from 'remark-math';

const envName = process.env.NODE_ENV ?? 'development';
const branch = process.env.GIT_BRANCH ?? 'main';
const isDevelopment = envName !== 'production';

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
  globalStyles: path.join(__dirname, 'theme/index.css'),
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
  plugins: [pluginSitemap({ siteUrl: 'https://monetr.app' })],
  builderConfig: {
    plugins: [pluginSass()],
    output: {
      cleanDistPath: true,
    },
    resolve: {
      alias: {
        '@monetr/docs': path.resolve(__dirname, '.'),
      },
    },
    performance: {
      preload: {
        // Prevents dumb screen flash where the font is missing
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
