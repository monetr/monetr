import React from 'react';
import ReactDOM from 'react-dom';
import { Provider } from 'react-redux'
import reportWebVitals from './reportWebVitals';
import Root from "./root";
import configureStore from './store';
import "./styles/styles.css";
import './styles/index.scss';

const store = configureStore();

ReactDOM.render(
  <React.StrictMode>
    <Provider store={ store }>
      <Root/>
    </Provider>
  </React.StrictMode>,
  document.getElementById('root')
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
