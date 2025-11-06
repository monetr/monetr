import resolveConfig from 'tailwindcss/resolveConfig';
import type { ThemeConfig } from 'tailwindcss/types/config.js';

// eslint-disable-next-line no-relative-import-paths/no-relative-import-paths
import tailwindConfig from '../../tailwind.config.ts';

const realTailwindConfig = resolveConfig(tailwindConfig);

export type ColorScheme = 'dark' | 'light';

export interface MonetrTheme {
  tailwind: Partial<ThemeConfig>;
  mediaColorSchema: ColorScheme;
}

export default function useTheme(): MonetrTheme {
  return {
    tailwind: realTailwindConfig.theme,
    mediaColorSchema: 'dark',
  };
}
