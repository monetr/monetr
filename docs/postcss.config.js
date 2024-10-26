// TODO Convert this to a typescript file: https://github.com/vercel/next.js/pull/69827
/** @type {import('postcss').Postcss} */
module.exports = {
  plugins: [
    'tailwindcss/nesting',
    'tailwindcss',
    'autoprefixer',
    [
      'cssnano',
      {
        preset: ['default', {
          discardComments: {
            removeAll: true,
          },
        }],
      },
    ],
  ],
};
