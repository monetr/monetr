// vim: nospell

const baseConfig = require('../tailwind.config.js');

/** @type {import('tailwindcss').Config} */
module.exports = {
  ...baseConfig,
  content: [
    'src/**/*.@(js|jsx|ts|tsx)',
  ],
};
