import { defineConfig } from '@rsbuild/core';
import { pluginReact } from '@rsbuild/plugin-react';
import { pluginSass } from '@rsbuild/plugin-sass';
import path from 'path';

const interfaceSource = path.resolve(__dirname, '../interface/src');

export default defineConfig({
  source: {
    alias: {
      '@monetr/interface': interfaceSource,
    },
    define: {
      CONFIG: JSON.stringify({}),
      REVISION: JSON.stringify(process.env.RELEASE_REVISION),
      RELEASE: JSON.stringify(process.env.RELEASE_VERSION),
      NODE_VERSION: process.version,
    },
  },
  plugins: [
    pluginReact(),
    pluginSass(),
  ],
});
