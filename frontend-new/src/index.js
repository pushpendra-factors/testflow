import React from 'react';
import ReactDOM from 'react-dom';
import 'antd/dist/antd.css';
import './assets/tailwind.output.css';
import './assets/index.scss';
import './styles/factors-ai.main.scss';
import 'c3/c3.css';

import App from './App';

ReactDOM.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
  document.getElementById('root')
);