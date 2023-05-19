import type { StorybookConfig } from '@storybook/types';

const config: StorybookConfig = {
  stories: ['../ui/**/*.stories.@(js|jsx|ts|tsx)'],
  addons: [
    '@storybook/addon-links',
    '@storybook/addon-essentials',
    '@storybook/addon-interactions',
    '@storybook/addon-viewport',
    '@storybook/addon-coverage',
    {
      name: '@storybook/addon-styling',
      options: {
        // Check out https://github.com/storybookjs/addon-styling/blob/main/docs/api.md
        // For more details on this addon's options.
        postCss: true,
      },
    },
  ],
  framework: {
    name: 'storybook-react-rspack',
    options: {
      fastRefresh: true,
    },
  },
  docs: {
    autodocs: 'tag',
  },
};
export default config;
