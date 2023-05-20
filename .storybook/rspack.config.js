const path = require('path');

const root = path.resolve(__dirname, '../');
const uiDir = path.resolve(root, 'ui');

module.exports = ({ config, mode }) => {
  config = {
    ...config,
    devServer: {
      ...config?.devServer,
      client: {
        ...config?.devServer?.client,
        progress: true,
      },
    },
    resolve: {
      ...config?.resolve,
      modules: [
        ...config?.resolve?.modules,
        uiDir,
      ],
    },
    module: {
      ...config?.module,
      rules: [
        ...config?.module?.rules,
        {
          test: /\.css$/,
          use: [
            {
              loader: 'postcss-loader',
            },
          ],
        },
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
          ],
          type: 'css',
        },
      ],
    },
  };
  return config;
}
