// vim: nospell
import type { Config } from 'tailwindcss';
import plugin from 'tailwindcss/plugin';

const config: Partial<Config> = {
  important: true,
  darkMode: 'class',
  future: {
    hoverOnlyWhenSupported: true,
  },
  plugins: [
    require('tailwindcss-animate'),
    plugin(helper => {
      helper.addBase({
        ':root': {
          '--background': '0 0% 100%',
          '--foreground': '224 71.4% 4.1%',
          '--card': '0 0% 100%',
          '--card-foreground': '224 71.4% 4.1%',
          '--popover': '0 0% 100%',
          '--popover-foreground': '224 71.4% 4.1%',
          '--primary': '262.1 83.3% 57.8%',
          '--primary-foreground': '210 20% 98%',
          '--secondary': '220 14.3% 95.9%',
          '--secondary-foreground': '220.9 39.3% 11%',
          '--muted': '220 14.3% 95.9%',
          '--muted-foreground': '220 8.9% 46.1%',
          '--accent': '220 14.3% 95.9%',
          '--accent-foreground': '220.9 39.3% 11%',
          '--destructive': '0 84.2% 60.2%',
          '--destructive-foreground': '210 20% 98%',
          '--border': '220 13% 91%',
          '--input': '220 13% 91%',
          '--ring': '262.1 83.3% 57.8%',
          '--radius': '0.75rem',
          '--chart-1': '12 76% 61%',
          '--chart-2': '173 58% 39%',
          '--chart-3': '197 37% 24%',
          '--chart-4': '43 74% 66%',
          '--chart-5': '27 87% 67%',
        },
        '.dark': {
          // '--background': '224 71.4% 4.1%',
          '--background': '#19161f',
          '--foreground': '210 20% 98%',
          '--card': '224 71.4% 4.1%',
          '--card-foreground': '210 20% 98%',
          '--popover': '224 71.4% 4.1%',
          '--popover-foreground': '210 20% 98%',
          '--primary': '263.4 70% 50.4%',
          '--primary-foreground': '210 20% 98%',
          '--secondary': '215 27.9% 16.9%',
          '--secondary-foreground': '210 20% 98%',
          '--muted': '215 27.9% 16.9%',
          '--muted-foreground': '217.9 10.6% 64.9%',
          '--accent': '215 27.9% 16.9%',
          '--accent-foreground': '210 20% 98%',
          '--destructive': '0 62.8% 30.6%',
          '--destructive-foreground': '210 20% 98%',
          '--border': '215 27.9% 16.9%',
          '--input': '#71717a',
          '--ring': '263.4 70% 50.4%',
          '--chart-1': '220 70% 50%',
          '--chart-2': '160 60% 45%',
          '--chart-3': '30 80% 55%',
          '--chart-4': '280 65% 60%',
          '--chart-5': '340 75% 55%',
        },
      });
    }),
  ],
  theme: {
    extend: {
      // borderRadius: {
      //   lg: 'var(--radius)',
      //   md: 'calc(var(--radius) - 2px)',
      //   sm: 'calc(var(--radius) - 4px)',
      // },
      fontFamily: {
        sans: ['Inter Variable', 'sans-serif'],
        body: ['Inter Variable', 'sans-serif'],
      },
      backgroundImage: {
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
        'gradient-conic':
          'conic-gradient(from 180deg at 50% 50%, var(--tw-gradient-stops))',
      },
      keyframes: {
        'accordion-down': {
          from: { height: '0' },
          to: { height: 'var(--radix-accordion-content-height)' },
        },
        'accordion-up': {
          from: { height: 'var(--radix-accordion-content-height)' },
          to: { height: '0' },
        },
      },
      animation: {
        'accordion-down': 'accordion-down 0.2s ease-out',
        'accordion-up': 'accordion-up 0.2s ease-out',
      },
      gap: {
        'component': '0.125rem', // gap-0.5
        'stack': '0.5rem', // gap-2
      },
      colors: {
        inherit: 'inherit',
        current: 'currentColor',
        transparent: 'transparent',
        black: '#000',
        white: '#fff',
        border: 'hsl(var(--border))',
        input: 'var(--input)',
        ring: 'hsl(var(--ring))',
        content: {
          DEFAULT: '#d4d4d8', // zinc-200
          placeholder: '#9ca3af', // gray-400
          disabled: '#6b7280' // gray-500
        },
        background: {
          subtle: '#27272a', // zinc-800
          DEFAULT: '#19161f',
          emphasis: '#3f3f46', // zinc-700
          focused: '#131118',
          bright: '#fafafa', // zinc-50
        },
        foreground: {
          muted: '#52525b',
          subtle: '#a1a1aa', // zinc-400
          DEFAULT: '#d4d4d8', // zinc-200
          emphasis: '#fafafa', // zinc-50
        },
        primary: {
          DEFAULT: '#4E1AA0',
          foreground: 'hsl(var(--primary-foreground))',
        },
        secondary: {
          DEFAULT: 'hsl(var(--secondary))',
          foreground: 'hsl(var(--secondary-foreground))',
        },
        destructive: {
          DEFAULT: '#ef4444',
          foreground: '#d4d4d8',
        },
        muted: {
          DEFAULT: 'hsl(var(--muted))',
          foreground: 'hsl(var(--muted-foreground))',
        },
        accent: {
          DEFAULT: '#3f3f46', // background-emphasis
          foreground: '#fafafa', // content-emphasis
        },
        popover: {
          DEFAULT: '#19161f', // background-DEFAULT
          foreground: '#fafafa', // content-emphasis
        },
        card: {
          DEFAULT: 'hsl(var(--card))',
          foreground: 'hsl(var(--card-foreground))',
        },
        'monetr': {
          brand: {
            DEFAULT: '#4E1AA0',
          },
          background: {
            subtle: '',
            DEFAULT: '#F8F8F8',
            emphasis: '',
          },
          border: {
            DEFAULT: '', // zinc-700
          },
          content: {
            subtle: '#6b7280', // gray-500
            DEFAULT: '#111827', // gray-900
            emphasis: '', // zinc-50
          },
        },
        'dark-monetr': {
          red: {
            DEFAULT: '#ef4444',
          },
          green: {
            DEFAULT: '#22c55e',
          },
          blue: {
            DEFAULT: '#3b82f6',
          },
          brand: {
            bright: '#CFB9F4',
            faint: '#AC84EB',
            muted: '#9461E5',
            subtle: '#5D1FC1',
            DEFAULT: '#4E1AA0',
          },
          background: {
            subtle: '#27272a', // zinc-800
            DEFAULT: '#19161f',
            emphasis: '#3f3f46', // zinc-700
            focused: '#131118',
            bright: '#fafafa', // zinc-50
          },
          border: {
            subtle: '#27272a', // zinc-800
            DEFAULT: '#3f3f46', // zinc-700
            string: '#71717a',
          },
          content: {
            muted: '#52525b',
            subtle: '#a1a1aa', // zinc-400
            DEFAULT: '#d4d4d8', // zinc-200
            emphasis: '#fafafa', // zinc-50
          },
          popover: {
            DEFAULT: '0 0% 100%',
            foreground: '222.2 47.4% 11.2%',
          },
        },
      },
    },
    aspectRatio: {
      'video-vertical': '9/16',
      'video': '16/9',
    },
    animation: {
      'ping-slow': 'ping 2s cubic-bezier(0, 0, 0.2, 1) infinite',
    },
    colors: {
      slate: {
        50: '#f8fafc',
        100: '#f1f5f9',
        200: '#e2e8f0',
        300: '#cbd5e1',
        400: '#94a3b8',
        500: '#64748b',
        600: '#475569',
        700: '#334155',
        800: '#1e293b',
        900: '#0f172a',
      },
      gray: {
        50: '#f9fafb',
        100: '#f3f4f6',
        200: '#e5e7eb',
        300: '#d1d5db',
        400: '#9ca3af',
        500: '#6b7280',
        600: '#4b5563',
        700: '#374151',
        800: '#1f2937',
        900: '#111827',
      },
      zinc: {
        50: '#fafafa',
        100: '#f4f4f5',
        200: '#e4e4e7',
        300: '#d4d4d8',
        400: '#a1a1aa',
        500: '#71717a',
        600: '#52525b',
        700: '#3f3f46',
        800: '#27272a',
        900: '#19161f',
        // 900: '#1c1821',
        // 900: '#1b181f',
        // 900: '#18181b',
      },
      neutral: {
        50: '#fafafa',
        100: '#f5f5f5',
        200: '#e5e5e5',
        300: '#d4d4d4',
        400: '#a3a3a3',
        500: '#737373',
        600: '#525252',
        700: '#404040',
        800: '#262626',
        900: '#171717',
      },
      stone: {
        50: '#fafaf9',
        100: '#f5f5f4',
        200: '#e7e5e4',
        300: '#d6d3d1',
        400: '#a8a29e',
        500: '#78716c',
        600: '#57534e',
        700: '#44403c',
        800: '#292524',
        900: '#1c1917',
      },
      red: {
        50: '#fef2f2',
        100: '#fee2e2',
        200: '#fecaca',
        300: '#fca5a5',
        400: '#f87171',
        500: '#ef4444',
        600: '#dc2626',
        700: '#b91c1c',
        800: '#991b1b',
        900: '#7f1d1d',
      },
      orange: {
        50: '#fff7ed',
        100: '#ffedd5',
        200: '#fed7aa',
        300: '#fdba74',
        400: '#fb923c',
        500: '#f97316',
        600: '#ea580c',
        700: '#c2410c',
        800: '#9a3412',
        900: '#7c2d12',
      },
      amber: {
        50: '#fffbeb',
        100: '#fef3c7',
        200: '#fde68a',
        300: '#fcd34d',
        400: '#fbbf24',
        500: '#f59e0b',
        600: '#d97706',
        700: '#b45309',
        800: '#92400e',
        900: '#78350f',
      },
      yellow: {
        50: '#fefce8',
        100: '#fef9c3',
        200: '#fef08a',
        300: '#fde047',
        400: '#facc15',
        500: '#eab308',
        600: '#ca8a04',
        700: '#a16207',
        800: '#854d0e',
        900: '#713f12',
      },
      lime: {
        50: '#f7fee7',
        100: '#ecfccb',
        200: '#d9f99d',
        300: '#bef264',
        400: '#a3e635',
        500: '#84cc16',
        600: '#65a30d',
        700: '#4d7c0f',
        800: '#3f6212',
        900: '#365314',
      },
      green: {
        50: '#f0fdf4',
        100: '#dcfce7',
        200: '#bbf7d0',
        300: '#86efac',
        400: '#4ade80',
        500: '#22c55e',
        600: '#16a34a',
        700: '#15803d',
        800: '#166534',
        900: '#14532d',
      },
      emerald: {
        50: '#ecfdf5',
        100: '#d1fae5',
        200: '#a7f3d0',
        300: '#6ee7b7',
        400: '#34d399',
        500: '#10b981',
        600: '#059669',
        700: '#047857',
        800: '#065f46',
        900: '#064e3b',
      },
      teal: {
        50: '#f0fdfa',
        100: '#ccfbf1',
        200: '#99f6e4',
        300: '#5eead4',
        400: '#2dd4bf',
        500: '#14b8a6',
        600: '#0d9488',
        700: '#0f766e',
        800: '#115e59',
        900: '#134e4a',
      },
      cyan: {
        50: '#ecfeff',
        100: '#cffafe',
        200: '#a5f3fc',
        300: '#67e8f9',
        400: '#22d3ee',
        500: '#06b6d4',
        600: '#0891b2',
        700: '#0e7490',
        800: '#155e75',
        900: '#164e63',
      },
      sky: {
        50: '#f0f9ff',
        100: '#e0f2fe',
        200: '#bae6fd',
        300: '#7dd3fc',
        400: '#38bdf8',
        500: '#0ea5e9',
        600: '#0284c7',
        700: '#0369a1',
        800: '#075985',
        900: '#0c4a6e',
      },
      blue: {
        50: '#eff6ff',
        100: '#dbeafe',
        200: '#bfdbfe',
        300: '#93c5fd',
        400: '#60a5fa',
        500: '#3b82f6',
        600: '#2563eb',
        700: '#1d4ed8',
        800: '#1e40af',
        900: '#1e3a8a',
      },
      indigo: {
        50: '#eef2ff',
        100: '#e0e7ff',
        200: '#c7d2fe',
        300: '#a5b4fc',
        400: '#818cf8',
        500: '#6366f1',
        600: '#4f46e5',
        700: '#4338ca',
        800: '#3730a3',
        900: '#312e81',
      },
      violet: {
        50: '#f5f3ff',
        100: '#ede9fe',
        200: '#ddd6fe',
        300: '#c4b5fd',
        400: '#a78bfa',
        500: '#8b5cf6',
        600: '#7c3aed',
        700: '#6d28d9',
        800: '#5b21b6',
        900: '#4c1d95',
      },
      purple: {
        50: '#EDE5FB',
        100: '#D8C6F6',
        200: '#B591ED',
        300: '#8E58E4',
        400: '#6823D7',
        500: '#4E1AA0',
        600: '#3E157F',
        700: '#2F1060',
        800: '#200B42',
        900: '#0F051F',
      },
      fuchsia: {
        50: '#fdf4ff',
        100: '#fae8ff',
        200: '#f5d0fe',
        300: '#f0abfc',
        400: '#e879f9',
        500: '#d946ef',
        600: '#c026d3',
        700: '#a21caf',
        800: '#86198f',
        900: '#701a75',
      },
      pink: {
        50: '#fdf2f8',
        100: '#fce7f3',
        200: '#fbcfe8',
        300: '#f9a8d4',
        400: '#f472b6',
        500: '#ec4899',
        600: '#db2777',
        700: '#be185d',
        800: '#9d174d',
        900: '#831843',
      },
      rose: {
        50: '#fff1f2',
        100: '#ffe4e6',
        200: '#fecdd3',
        300: '#fda4af',
        400: '#fb7185',
        500: '#f43f5e',
        600: '#e11d48',
        700: '#be123c',
        800: '#9f1239',
        900: '#881337',
      },
    },
  },
};
export default config;
