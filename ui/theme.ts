import { createTheme, darken } from '@mui/material';

import tailwindConfig from '../tailwind.config.cjs';

import resolveConfig from 'tailwindcss/resolveConfig';

const fullConfig = resolveConfig(tailwindConfig);
const darkMode = false; // window.localStorage.getItem('darkMode') === 'true';
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
    primary: {
      main: darkMode ? '#712ddd' : defaultPrimary,
      light: defaultPrimary,
      dark: '#712ddd',
      contrastText: '#FFFFFF',
    },
    secondary: {
      main: darkMode ? darken(defaultSecondary, 0.2) : defaultSecondary,
      contrastText: '#FFFFFF',
    },
    background: {
      default: darkMode ? fullConfig.theme.colors['neutral']['800'] : '#FFFFFF',
      // default: darkMode ? '#2f2f2f' : '#FFFFFF',
    },
  },
});

export default theme;
