import path, { resolve } from 'path';

const root = resolve(__dirname, '../');
const uiDir = resolve(root, 'interface/src');
const emailDir = resolve(root, 'emails/src');
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
        emailDir,
        'node_modules',
      ],
      alias: {
        ...config?.resolve?.alias,
        '@monetr/interface': uiDir,
      }
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
