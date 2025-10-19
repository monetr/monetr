// vim: nospell
import type { Config } from 'tailwindcss';

// eslint-disable-next-line no-relative-import-paths/no-relative-import-paths
import baseConfig from '../tailwind.config.ts';

const config: Config = {
  ...baseConfig,
  content: ['src/**/*.@(js|jsx|ts|tsx)'],
};
export default config;
