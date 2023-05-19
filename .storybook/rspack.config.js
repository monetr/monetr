const path = require('path');

const root = path.resolve(__dirname, '../');
const uiDir = path.resolve(root, 'ui');

module.exports = ({ config, mode}) => {
  config.resolve.modules.push(uiDir);
  config.module.rules.push({
    test: /\.css$/,
    use: [
      {
        loader: 'postcss-loader',
      },
    ],
  });
  config.module.rules.push({
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
  });
  return config;
}
