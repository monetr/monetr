// vim: nospell
const baseConfig = require('../tailwind.config.js');
const { resolve } = require('path');

const root = resolve(__dirname, '../');
const uiDir = resolve(root, 'interface/src');
const storiesDir = resolve(root, 'stories/stories');

/** @type {import('tailwindcss').Config} */
module.exports = {
  ...baseConfig,
  content: [
    `${uiDir}/**/*.@(js|jsx|ts|tsx)`,
    `${storiesDir}/**/*.@(js|jsx|ts|tsx)`,
  ],
};
