const path = require('path');

/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: false,
  output: 'export',
  distDir: './out',
  trailingSlash: true,
  images: {
    unoptimized: true,
  },
  generateBuildId: async () => {
    return 'monetr.app';
  },
  webpack: (
    config,
    nextShit,
  ) => {
    // Important: return the modified config
    config.resolve = {
      ...config?.resolve,
      alias: {
        ...config?.resolve?.alias,
        '@monetr/docs': path.resolve(__dirname, 'src'),
        '@monetr/interface': path.resolve(__dirname, '../interface/src'),
      }
    }
    return config
  },
}

const withNextra = require('nextra')({
  theme: 'nextra-theme-docs',
  themeConfig: './theme.config.tsx',
  flexsearch: {
    codeblocks: false,
  }
});

module.exports = withNextra(nextConfig);
