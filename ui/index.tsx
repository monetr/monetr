import React from 'react';
import ReactDOM from 'react-dom';
import { Provider } from 'react-redux'
import reportWebVitals from './reportWebVitals';
import Root from './root';
import './styles/styles.css';
import './styles/index.scss';
import { BrowserRouter as Router } from 'react-router-dom';
import { createTheme, MuiThemeProvider, Typography } from '@material-ui/core';
import { store } from 'store';

// {/* eslint-disable-next-line no-undef */
// }
// if (module.hot) {
//   module.hot.accept()
// }

const theme = createTheme({
  palette: {
    primary: {
      main: '#4E1AA0',
      contrastText: '#FFFFFF',
    },
    secondary: {
      main: '#FF5798',
      contrastText: '#FFFFFF',
    }
  }
});

ReactDOM.render(
  <React.StrictMode>
    <Provider store={ store }>
      <Router>
        <MuiThemeProvider theme={ theme }>
          <Root/>
          <Typography
            className="absolute inline w-full text-center bottom-1 opacity-30"
          >
            © { new Date().getFullYear() } monetr LLC
          </Typography>
        </MuiThemeProvider>
      </Router>
    </Provider>
  </React.StrictMode>,
  document.getElementById('root')
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
