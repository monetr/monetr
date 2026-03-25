import { resolve, join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';
import { defineConfig } from '@rsbuild/core';
import { pluginReact } from '@rsbuild/plugin-react';
import { rsbuildPluginEmail } from './src/build/rsbuildPluginEmail';
import tailwindConfig from './tailwind.config';

const __dirname = dirname(fileURLToPath(import.meta.url));

// rsbuild sets NODE_ENV before loading config:
//   rsbuild dev   → NODE_ENV=development
//   rsbuild build → NODE_ENV=production
const isDev = process.env.NODE_ENV !== 'production';

const outDir = process.env.EMAIL_OUT_DIR
  ? resolve(process.env.EMAIL_OUT_DIR)
  : join(__dirname, 'dist', 'emails');

export default defineConfig({
  plugins: [
    pluginReact(),
    // Only include the email rendering plugin for production builds
    ...(!isDev ? [rsbuildPluginEmail({ outDir, tailwindConfig })] : []),
  ],
  environments: {
    // Dev: web preview server
    ...(isDev ? {
      web: {
        source: {
          entry: { index: './src/preview/index.tsx' },
        },
        html: {
          template: './src/preview/index.html',
        },
        output: {
          target: 'web',
        },
      },
    } : {}),
    // Build: node environment for server-side rendering of templates
    ...(!isDev ? {
      node: {
        source: {
          entry: {
            '__email_ssg__/email-bundle': './src/build/entry.ts',
          },
        },
        output: {
          target: 'node',
          distPath: { root: 'dist/server' },
          filename: { js: '[name].cjs' },
          minify: false,
        },
        // Preserve all exports (prevent tree-shaking of template components)
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
    } : {}),
  },
  server: {
    port: 3100,
  },
});
