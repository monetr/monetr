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
