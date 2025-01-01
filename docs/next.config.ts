import { NextConfig } from 'next';
import nextra from 'nextra';

import path from 'path';

const nextConfig: NextConfig = {
  reactStrictMode: false,
  output: 'export',
  distDir: './out',
  trailingSlash: true,
  experimental: {
    scrollRestoration: true,
  },
  images: {
    unoptimized: true,
  },
  generateBuildId: async () => {
    return 'monetr.app';
  },
  sassOptions: {
    implementation: 'sass-embedded',
  },
  typescript: {
    // Fuck you next
    ignoreBuildErrors: true,
  },
  webpack: (
    config,
    _,
  ) => {
    // Important: return the modified config
    config.resolve = {
      ...config?.resolve,
      alias: {
        ...config?.resolve?.alias,
        '@monetr/docs': path.resolve(__dirname, 'src'),
      },
      modules: [
        ...config?.resolve?.modules,
      ],
      extensions: [
        ...config?.resolve?.extensions,
        '.svg',
      ],
      extensionAlias: {
        ...config?.resolve?.extensionAlias,
        '.js': ['.ts', '.tsx', '.js', '.jsx'],
        '.mjs': ['.mts', '.mjs'],
        '.cjs': ['.cts', '.cjs'],
        '.svg': ['.svg'],
      },
    };
    return config;
  },
};

const withNextra = nextra({
  theme: 'nextra-theme-docs',
  themeConfig: './theme.config.tsx',
  search: {
    codeblocks: false,
  },
});

module.exports = withNextra(nextConfig);
