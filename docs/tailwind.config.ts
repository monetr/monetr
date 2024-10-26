// vim: nospell
import type { Config } from 'tailwindcss';

// eslint-disable-next-line no-relative-import-paths/no-relative-import-paths
import baseConfig from '../tailwind.config.ts';

import { resolve } from 'path';

const root = resolve(__dirname, '../');
const uiDir = resolve(root, 'interface/src');

const config: Config = {
  ...baseConfig,
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,md,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,md,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,md,mdx}',
    './theme.config.tsx',
    `${uiDir}/**/*.@(js|jsx|ts|tsx)`,
  ],
};

export default config;
