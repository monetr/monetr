import { defineConfig } from '@rsbuild/core';
import { pluginReact } from '@rsbuild/plugin-react';

export default defineConfig({
  plugins: [pluginReact()],
  source: {
    entry: {
      index: './src/preview/index.tsx',
    },
  },
  html: {
    template: './src/preview/index.html',
  },
  server: {
    port: 3100,
  },
});
