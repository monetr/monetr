// vim: nospell
import type { Config } from 'tailwindcss';

// eslint-disable-next-line no-relative-import-paths/no-relative-import-paths
import baseConfig from '../tailwind.config.ts';

import { resolve } from 'node:path';

const root = resolve(__dirname, '../');
const uiDir = resolve(root, 'interface/src');
const storiesDir = resolve(root, 'stories/stories');

const config: Config = {
  ...baseConfig,
  content: [`${uiDir}/**/*.@(js|jsx|ts|tsx)`, `${storiesDir}/**/*.@(js|jsx|ts|tsx)`],
};
export default config;
