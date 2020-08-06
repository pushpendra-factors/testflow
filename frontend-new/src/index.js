import React from 'react';
import ReactDOM from 'react-dom';
import 'antd/dist/antd.css';
import './assets/index.scss';
import App from './App';
import './assets/tailwind.output.css';

ReactDOM.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
  document.getElementById('root')
);