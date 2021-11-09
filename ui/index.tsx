import GlobalFooter from 'components/GlobalFooter';
import React from 'react';
import ReactDOM from 'react-dom';
import { Provider } from 'react-redux'
import reportWebVitals from './reportWebVitals';
import Root from './root';
import './styles/styles.css';
import './styles/index.scss';
import { BrowserRouter as Router } from 'react-router-dom';
import { createTheme, CssBaseline, ThemeProvider, Typography } from '@mui/material';
import { store } from 'store';
import AdapterMoment from '@mui/lab/AdapterMoment';
import { LocalizationProvider } from '@mui/lab'

const darkMode = window.localStorage.getItem('darkMode') === 'true';

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
    }
  }
});

ReactDOM.render(
  <React.StrictMode>
    <Provider store={ store }>
      <Router>
        <ThemeProvider theme={ theme }>
          <LocalizationProvider dateAdapter={ AdapterMoment }>
            <CssBaseline/>
            <Root/>
            <GlobalFooter/>
          </LocalizationProvider>
        </ThemeProvider>
      </Router>
    </Provider>
  </React.StrictMode>,
  document.getElementById('root')
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
