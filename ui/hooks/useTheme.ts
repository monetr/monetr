import { useEffect, useState } from 'react';
import { Theme } from '@mui/material';

import tailwindConfig from '../../tailwind.config.cjs';

import resolveConfig from 'tailwindcss/resolveConfig';
import { ThemeConfig } from 'tailwindcss/types/config.js';
import theme from 'theme';

const realTailwindConfig = resolveConfig(tailwindConfig);

export type ColorScheme = 'dark' | 'light';

export interface MonetrTheme {
  tailwind: Partial<ThemeConfig>;
  material: Theme;
  mediaColorSchema: ColorScheme;
}

export default function useTheme(): MonetrTheme {
  const [mode, setMode] = useState<ColorScheme>();

  useEffect(() => {
    window.matchMedia('(prefers-color-scheme: dark)')
      .addEventListener('change', event => {
        const colorScheme = event.matches ? 'dark' : 'light';
        setMode(colorScheme);
      });
  }, []);

  return {
    tailwind: realTailwindConfig.theme,
    material: theme,
    mediaColorSchema: mode,
  };
}
