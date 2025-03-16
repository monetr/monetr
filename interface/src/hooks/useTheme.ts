import { Theme } from '@mui/material';
import resolveConfig from 'tailwindcss/resolveConfig';
import { ThemeConfig } from 'tailwindcss/types/config.js';

import theme from '@monetr/interface/theme';

// eslint-disable-next-line no-relative-import-paths/no-relative-import-paths
import tailwindConfig from '../../tailwind.config.ts';

const realTailwindConfig = resolveConfig(tailwindConfig);

export type ColorScheme = 'dark' | 'light';

export interface MonetrTheme {
  tailwind: Partial<ThemeConfig>;
  material: Theme;
  mediaColorSchema: ColorScheme;
}

export default function useTheme(): MonetrTheme {
  return {
    tailwind: realTailwindConfig.theme,
    material: theme,
    mediaColorSchema: 'dark',
  };
}
