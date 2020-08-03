import React from 'react';
import ReactDOM from 'react-dom';
import './assets/index.scss';
import 'font-awesome/css/font-awesome.min.css';
import App from './App';
import './assets/tailwind.output.css';
import 'antd/dist/antd.css';

ReactDOM.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
  document.getElementById('root')
);