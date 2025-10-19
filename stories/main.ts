import type { StorybookConfig } from 'storybook-react-rsbuild';

const marketingStoryOnly = process.env.MARKETING_STORY_ONLY === 'true';
let stories = [
  '../interface/src/**/*.stories.@(js|jsx|ts|tsx)',
  '../emails/**/*.stories.tsx',
  'stories/**/*.stories.tsx',
];
if (marketingStoryOnly) {
  stories = ['../interface/src/pages/app.stories.tsx'];
}

const config: StorybookConfig = {
  stories: stories,
  addons: [
    '@storybook/addon-links',
    {
      name: '@storybook/addon-essentials',
      options: {
        backgrounds: false,
      },
    },
    '@storybook/addon-interactions',
    '@storybook/addon-viewport',
    'storybook-addon-react-router-v6',
    // TODO Add storycap back in later for screenshots of stories.
    // 'storycap',
  ],
  framework: 'storybook-react-rsbuild',
  rsbuildFinal: config => {
    if (marketingStoryOnly) {
      // Marketing story only merges the storybook dist with the next.js dist from docs.
      // as a result, the subpath for storybook needs to be /_storybook in order to be served properly.
      config.output ??= {};
      config.output.assetPrefix = '/_storybook/';
    }
    return config;
  },
  docs: {
    autodocs: 'tag',
  },
};
export default config;
