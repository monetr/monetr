import { rsbuildPluginEmail } from './src/build/rsbuildPluginEmail';

import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { defineConfig } from '@rsbuild/core';
import { pluginReact } from '@rsbuild/plugin-react';
import { pluginSass } from '@rsbuild/plugin-sass';

const __dirname = dirname(fileURLToPath(import.meta.url));
const emailsSource = resolve(__dirname, 'src');

// rsbuild sets NODE_ENV before loading config:
//   rsbuild dev   -> NODE_ENV=development
//   rsbuild build -> NODE_ENV=production
const isDev = process.env.NODE_ENV !== 'production';

const outDir = process.env.EMAIL_OUT_DIR
  ? resolve(process.env.EMAIL_OUT_DIR)
  : join(__dirname, 'dist', 'emails');

const devEnvironments = {
  web: {
    source: {
      entry: { index: './src/preview/index.tsx' },
    },
    html: {
      template: './src/preview/index.html',
    },
    output: {
      target: 'web' as const,
    },
  },
};

const buildEnvironments = {
  node: {
    source: {
      entry: {
        '__email_ssg__/email-bundle': './src/build/entry.ts',
      },
    },
    output: {
      // Use web target so rspack extracts CSS to separate files instead of
      // discarding it (node target only exports class name mappings).
      // Safe because the email templates are pure React with no Node APIs.
      target: 'web' as const,
      distPath: { root: 'dist/server', js: '', css: '' },
      filename: { js: '[name].cjs', css: '[name].css' },
      minify: false,
      injectStyles: false,
    },
    // rspack needs these flags to preserve all named exports from the bundle.
    // Without them, the template components get tree-shaken or lost during
    // module concatenation since nothing inside the bundle consumes them.
    tools: {
      rspack: {
        optimization: {
          usedExports: false,
          concatenateModules: false,
        },
        output: {
          library: {
            type: 'commonjs2',
          },
        },
      },
    },
  },
};

export default defineConfig({
  plugins: [
    pluginReact(),
    pluginSass(),
    ...(!isDev ? [rsbuildPluginEmail({ outDir })] : []),
  ],
  resolve: {
    alias: {
      '@monetr/emails': emailsSource,
    },
  },
  environments: isDev ? devEnvironments : buildEnvironments,
  server: {
    port: 3100,
  },
});
