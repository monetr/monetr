import type { StorybookConfig } from "@storybook/react-webpack5";

import getParentWebpackConfig from '../webpack.config.cjs';

const webpackConfig = getParentWebpackConfig({}, {});

const config: StorybookConfig = {
  stories: ["../ui/**/*.mdx", "../ui/**/*.stories.@(js|jsx|ts|tsx)"],
  addons: [
    "@storybook/addon-links",
    "@storybook/addon-essentials",
    "@storybook/addon-interactions",
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
    name: "@storybook/react-webpack5",
    options: {},
  },
  docs: {
    autodocs: "tag",
  },
  webpackFinal: async (config) => {
    config.resolve = {
      ...config.resolve,
      ...webpackConfig.resolve,
    }

    return config;
  },
};
export default config;
