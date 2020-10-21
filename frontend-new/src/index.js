import React from 'react';
import ReactDOM from 'react-dom';
import { Provider } from 'react-redux';
import { createStore, compose, applyMiddleware } from 'redux';
import thunk from 'redux-thunk';
import { createPromise } from 'redux-promise-middleware';
import reducers from './reducers';
import './assets/tailwind.output.css';
import './assets/index.scss';
import './styles/factors-ai.main.scss';
import 'c3/c3.css';

import App from './App';

const composeEnhancer = window.__REDUX_DEVTOOLS_EXTENSION_COMPOSE__ || compose;

const middlewares = [
  createPromise(),
  thunk
];

const store = createStore(reducers, composeEnhancer(applyMiddleware(...middlewares)));

ReactDOM.render(
  <React.StrictMode>
    <Provider store={store}>
      <App />
    </Provider>
  </React.StrictMode>,
  document.getElementById('root')
);
