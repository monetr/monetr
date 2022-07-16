import { createTheme } from '@mui/material';

const darkMode = false; // window.localStorage.getItem('darkMode') === 'true';
const theme = createTheme({
  shape: {
    borderRadius: 10,
  },
  palette: {
    mode: darkMode ? 'dark' : 'light',
    primary: {
      main: darkMode ? '#712ddd' : '#4E1AA0',
      contrastText: '#FFFFFF',
    },
    secondary: {
      main: '#FF5798',
      contrastText: '#FFFFFF',
    },
    background: {
      default: darkMode ? '#2f2f2f' : '#FFFFFF',
    },
  },
});

export default theme;
