import path from 'path';
import { defineConfig } from 'rspress/config';

export default defineConfig({
  globalStyles: path.join(__dirname, 'src/styles/styles.css'),
  root: 'src/docs',
  title: 'monetr',
  lang: 'en',
  route: {
    cleanUrls: true,
  },
  // locales: [
  //   {
  //     lang: 'en-US',
  //     label: 'English',
  //     title: 'Doc Tools',
  //     description: 'Doc Tools',
  //   },
  // ],
  themeConfig: {
    darkMode: true,
    nav: [
      {
        text: 'Pricing',
        link: '/pricing/',
        position: 'right',
      },
      {
        text: 'Blog',
        link: '/blog/',
        position: 'right',
      },
      {
        text: 'Documentation',
        link: '/documentation/',
        position: 'right',
      },
    ],
  },
});
