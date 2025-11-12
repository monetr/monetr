// vim: nospell
import type { Config } from 'tailwindcss';

import baseConfig from '../tailwind.config.ts';

import { resolve } from 'path';

const root = resolve(__dirname, '../');
const uiDir = resolve(root, 'interface/src');

const config: Config = {
  ...baseConfig,
  theme: {
    ...baseConfig.theme,
    extend: {
      ...baseConfig.theme?.extend,
      animation: {
        ...baseConfig.theme?.extend?.animation,
        first: 'moveVertical 60s ease infinite',
        second: 'moveInCircle 40s reverse infinite',
        third: 'moveInCircle 80s linear infinite',
        fourth: 'moveHorizontal 80s ease infinite',
        fifth: 'moveInCircle 40s ease infinite',
      },
      keyframes: {
        ...baseConfig.theme?.extend?.keyframes,
        moveHorizontal: {
          '0%': {
            transform: 'translateX(-50%) translateY(-10%)',
          },
          '50%': {
            transform: 'translateX(50%) translateY(10%)',
          },
          '100%': {
            transform: 'translateX(-50%) translateY(-10%)',
          },
        },
        moveInCircle: {
          '0%': {
            transform: 'rotate(0deg)',
          },
          '50%': {
            transform: 'rotate(180deg)',
          },
          '100%': {
            transform: 'rotate(360deg)',
          },
        },
        moveVertical: {
          '0%': {
            transform: 'translateY(-50%)',
          },
          '50%': {
            transform: 'translateY(50%)',
          },
          '100%': {
            transform: 'translateY(-50%)',
          },
        },
      },
    },
  },
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,md,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,md,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,md,mdx}',
    './theme.config.tsx',
    `${uiDir}/**/*.@(js|jsx|ts|tsx)`,
  ],
};

export default config;
