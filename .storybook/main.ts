import type { StorybookConfig } from '@storybook/react-webpack5';

import getParentWebpackConfig from '../webpack.config.cjs';

import { RuleSetRule } from 'webpack';

const webpackConfig = getParentWebpackConfig({}, {});

const config: StorybookConfig = {
  stories: ['../ui/**/*.mdx', '../ui/**/*.stories.@(js|jsx|ts|tsx)'],
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
    name: '@storybook/react-webpack5',
    options: {
      fastRefresh: true,
    },
  },
  docs: {
    autodocs: 'tag',
  },
  webpackFinal: async config => {
    config.resolve = {
      ...config.resolve,
      ...webpackConfig.resolve,
    };


    // @ts-ignore
    const fileLoaderRule = config.module.rules.filter(
      // @ts-ignore
      rule => rule.test && rule.test.test('.svg'),
    );
    // @ts-ignore
    fileLoaderRule!.forEach(rule => rule.exclude = /\.svg$/);

    config!.module!.rules?.push({
      test: /\.(svg)$/,
      type: 'asset',
      parser: {
        dataUrlCondition: {
          maxSize: 4 * 1024,
        },
      },
      generator: {
        filename: 'assets/img/[hash][ext][query]',
      },
    });


    return config;
  },
};
export default config;
