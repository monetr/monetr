// vim: nospell
import type { Config } from 'tailwindcss';

import baseConfig from '../tailwind.config.ts';

const config: Config = {
  ...baseConfig,
  content: ['src/**/*.@(js|jsx|ts|tsx)'],
};
export default config;
