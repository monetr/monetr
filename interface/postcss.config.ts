import type { Config } from 'postcss-load-config';

const config: Config = {
  plugins: {
    autoprefixer: {},
    cssnano: {
      preset: [
        'default',
        {
          discardComments: {
            removeAll: true,
          },
        },
      ],
    },
  },
};

export default config;
