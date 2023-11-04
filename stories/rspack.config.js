import { resolve } from 'path';

const root = resolve(__dirname, '../');
const uiDir = resolve(root, 'interface/src');
const mockServiceWorkerJS = resolve(root, 'interface/public/mockServiceWorker.js');
const envName = process.env.NODE_ENV;
const isDevelopment = envName !== 'production';

export default ({ config, mode }) => {
  // This is so ugly, but basically this is doing a "deep" merge of what config values I need and the config values that
  // rspack and storybook actually need.
  config = {
    ...config,
    experiments: {
      ...config?.experiments,
      incrementalRebuild: true,
    },
    builtins: {
      ...config?.builtins,
      react: {
        ...config?.builtins?.react,
        runtime: 'automatic',
        development: isDevelopment,
        refresh: isDevelopment,
      },
      copy: {
        ...config?.builtins?.copy,
        patterns: [
          // This makes it so that the mock service worker actually works properly with rspack and storybook.
          ...(config?.builtins?.copy?.patterns ?? []),
          {
            from: mockServiceWorkerJS,
            to: 'mockServiceWorker.js',
          },
        ],
      }
    },
    devServer: isDevelopment ? {
      ...config?.devServer,
      liveReload: true,
      client: {
        ...config?.devServer?.client,
        progress: true,
      },
    } : config?.devServer,
    resolve: {
      ...config?.resolve,
      modules: [
        ...config?.resolve?.modules,
        uiDir,
        'node_modules',
      ],
    },
    module: {
      ...config?.module,
      rules: [
        ...config?.module?.rules,
        {
          test: /\.scss$/,
          use: [
            {
              loader: 'sass-loader',
              options: {
                sassOptions: {
                  quietDeps: true,
                },
              },
            },
            {
              loader: 'postcss-loader',
            },
          ],
          type: 'css',
        },
        {
          test: /\.css$/,
          use: [
            {
              loader: 'postcss-loader',
            },
          ],
          type: 'css',
        },
        {
          test: /\.svg$/,
          type: 'asset',
          parser: {
            dataUrlCondition: {
              maxSize: 1 * 1024 * 1024, // 1MB
            },
          },
          generator: {
            filename: 'assets/img/[contenthash][ext][query]',
          },
        },
      ],
    },
  };
  return config;
}
