import type { Config } from 'postcss-load-config';

const config: Config = {
  plugins: [require('tailwindcss/nesting'), require('tailwindcss'), require('autoprefixer')],
};

export default config;
