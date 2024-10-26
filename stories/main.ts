import { resolve } from 'path';
import type { StorybookConfig } from 'storybook-react-rsbuild';

const envName = process.env.NODE_ENV;
const isDevelopment = envName !== 'production';

const root = resolve(__dirname, '../');
const uiDir = resolve(root, 'interface/src');
const emailDir = resolve(root, 'emails/src');

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
    // {
    //   name: '@storybook/addon-styling',
    //   options: {
    //     // Check out https://github.com/storybookjs/addon-styling/blob/main/docs/api.md
    //     // For more details on this addon's options.
    //     postCss: true,
    //   },
    // },
    // 'storycap',
  ],
  framework: 'storybook-react-rsbuild',
  // webpackFinal: async config => {
  //   config.mode = isDevelopment ? 'development' : 'production';
  //   if (isDevelopment) {
  //     //@ts-ignore
  //     config.devServer = {
  //       //@ts-ignore
  //       ...config.devServer,
  //       hot: true,
  //     };
  //
  //     config.plugins = [
  //       ...config.plugins,
  //       new ReactRefreshWebpackPlugin(),
  //     ];
  //   }
  //   config.resolve.alias = {
  //     ...config.resolve.alias,
  //     '@monetr/interface': uiDir,
  //   };
  //   config.resolve.extensions = [
  //     ...config.resolve.extensions,
  //     '.svg',
  //     '.scss',
  //     '.css',
  //   ];
  //   config.resolve.modules = [
  //     ...config.resolve.modules,
  //     uiDir,
  //     emailDir,
  //     'node_modules',
  //   ];
  //   config.module.rules = [
  //     ...config.module.rules,
  //     {
  //       test: /\.?scss$/,
  //       use: [
  //         'style-loader',
  //         'css-loader',
  //         'postcss-loader',
  //         'sass-loader',
  //       ],
  //     },
  //     {
  //       test: /\.svg$/,
  //       parser: {
  //         dataUrlCondition: {
  //           maxSize: 1 * 1024 * 1024, // 1MB
  //         },
  //       },
  //     },
  //   ];
  //
  //   return config;
  // },
  docs: {
    autodocs: 'tag',
  },
};
export default config;
