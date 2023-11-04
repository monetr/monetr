import type { StorybookConfig } from '@storybook/types';

const envName = process.env.NODE_ENV;
const isDevelopment = envName !== 'production';

const config: StorybookConfig = {
  stories: ['../interface/src/**/*.stories.@(js|jsx|ts|tsx)'],
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
    '@storybook/addon-coverage',
    'storybook-addon-react-router-v6',
    {
      name: '@storybook/addon-styling',
      options: {
        // Check out https://github.com/storybookjs/addon-styling/blob/main/docs/api.md
        // For more details on this addon's options.
        postCss: true,
      },
    },
    'storycap',
  ],
  framework: {
    name: 'storybook-react-rspack',
    options: {
      fastRefresh: isDevelopment,
    },
  },
  docs: {
    autodocs: 'tag',
  },
};
export default config;
