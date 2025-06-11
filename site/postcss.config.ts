import type { Config } from 'postcss-load-config';

const config: Config = {
  plugins: [
    require('tailwindcss/nesting'),
    require('tailwindcss'),
    require('autoprefixer'),
    require('cssnano')({
      preset: ['default', {
        discardComments: {
          removeAll: true,
        },
      }],
    }),
  ],
};

export default config;
