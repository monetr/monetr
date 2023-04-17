import { createTheme, darken } from '@mui/material';

import tailwindConfig from '../tailwind.config.cjs';

import resolveConfig from 'tailwindcss/resolveConfig';

const fullConfig = resolveConfig(tailwindConfig);
const darkMode = true; // window.localStorage.getItem('darkMode') === 'true';
const inputHeight = 56; // Default is 56
const defaultPrimary = fullConfig.theme.colors['purple']['500']; // '#4E1AA0';
const defaultSecondary = '#FF5798';

const theme = createTheme({
  shape: {
    borderRadius: 10,
  },
  components: {
    MuiAppBar: {
      styleOverrides: {
        root: {
          backgroundColor: defaultPrimary,
          backgroundImage: 'none',
        },
      },
    },
    MuiInputBase: {
      styleOverrides: {
        root: {
          height: inputHeight,
        },
      },
    },
    MuiTextField: {
      styleOverrides: {
        root: {
          height: inputHeight,
        },
      },
    },
    MuiInputLabel: {
      styleOverrides: {
        root: {
          // transform: `translate(14px, ${(inputHeight / 3.5).toFixed(0)}px) scale(1)`,
        },
      },
    },
  },
  palette: {
    mode: darkMode ? 'dark' : 'light',
    text: darkMode ? {
      primary:  fullConfig.theme.colors['purple']['200'],
    } : {},
    primary: {
      main: darkMode ?
        fullConfig.theme.colors['purple']['100'] :
        fullConfig.theme.colors['purple']['500'],
      dark: fullConfig.theme.colors['purple']['100'],
      light: fullConfig.theme.colors['purple']['500'],
      contrastText: '#FFFFFF',
    },
    secondary: {
      main: darkMode ? darken(defaultSecondary, 0.2) : defaultSecondary,
      contrastText: '#FFFFFF',
    },
    background: {
      default: darkMode ? fullConfig.theme.colors['purple']['800'] :
        fullConfig.theme.colors['purple']['50'],
    },
  },
});

export default theme;
