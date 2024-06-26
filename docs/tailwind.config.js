
const { resolve } = require('path');
const baseConfig = require('../tailwind.config.js');

const root = resolve(__dirname, '../');
const uiDir = resolve(root, 'interface/src');

/** @type {import('tailwindcss').Config} */
module.exports = {
  ...baseConfig,
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,md,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,md,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,md,mdx}',
    './theme.config.tsx',
    `${uiDir}/**/*.@(js|jsx|ts|tsx)`,
  ],
};
