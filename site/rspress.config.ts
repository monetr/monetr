import * as path from 'node:path';
import { pluginSass } from '@rsbuild/plugin-sass';
import { defineConfig } from '@rspress/core';

export default defineConfig({
  root: path.join(__dirname, 'src'),
  lang: 'en',
  logo: 'https://monetr.app/_next/static/media/logo.77f6bc96.svg',
  logoText: 'monetr',
  themeConfig: {
    darkMode: true,
  },
  locales: [
    {
      lang: 'en',
      label: 'English',
      title: 'monetr',
      description: 'monetr',
    },
  ],
  route: {
    cleanUrls: true,
    excludeConvention: ['**/public/*'],
  },
  multiVersion: {
    default: 'v1',
    versions: ['v1'],
  },
  search: {
    versioned: true,
  },
  builderConfig: {
    resolve: {
      alias: {
        '@monetr/site': __dirname,
      }
    },
    html: {
      tags: [
        {
          tag: 'script',
          // Specify the default theme mode, which can be `dark` or `light`
          children: "window.RSPRESS_THEME = 'dark';",
        },
      ],
    },
    plugins: [
      pluginSass(),
    ]
  },
});
