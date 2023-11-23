const path = require('path');
const TsconfigPathsPlugin = require('tsconfig-paths-webpack-plugin');

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
    const interfaceDir = path.resolve(__dirname, '../interface/src')
    console.log(interfaceDir)
    // Important: return the modified config
    config.resolve = {
      ...config?.resolve,
      modules: [
        ...config?.resolve?.modules,
        interfaceDir,
      ],
      alias: {
        ...config?.resolve?.alias,
        '@monetr/docs': path.resolve(__dirname, 'src'),
        '@monetr/interface': interfaceDir,
      },
      plugins: [
        ...config?.resolve?.plugins,
        new TsconfigPathsPlugin({
          logLevel: 'info',
        }),
      ]
    }
    config.module.rules = [
      ...config?.module?.rules,
      {
        test: /interface.+\.tsx?$/,
        use: 'ts-loader',
        exclude: /node_modules/,
      },
    ]
    config.module.rules.forEach(item => {
      console.log(item)
    })
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
