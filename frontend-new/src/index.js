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
import App from './App';
import { persistStore, persistReducer } from 'redux-persist';
import storage from 'redux-persist/lib/storage'; // defaults to localStorage for web
import { PersistGate } from 'redux-persist/integration/react';
import ErrorBoundary from './ErrorBoundary';
import * as Sentry from "@sentry/react";
import { Integrations } from "@sentry/tracing";
import { TourProvider } from '@reactour/tour';
import steps from './steps';



if (!process.env.NODE_ENV || process.env.NODE_ENV === 'development') {
  // dev env
} else {
  // production env
  Sentry.init({
    dsn: "https://edc572f4f8bb4c8094acbc8df35389cf@o435495.ingest.sentry.io/5567060",
    autoSessionTracking: true,
    integrations: [
      new Integrations.BrowserTracing(),
    ],
  
    // We recommend adjusting this value in production, or using tracesSampler
    // for finer control
    tracesSampleRate: 1.0,
  }); 
}



const persistConfig = {
  key: 'root',
  storage,
  whitelist: ['agent','factors']
};
const persistedReducer = persistReducer(persistConfig, reducers);

const composeEnhancer = window.__REDUX_DEVTOOLS_EXTENSION_COMPOSE__ || compose;

const middlewares = [
  createPromise(),
  thunk
];

const store = createStore(persistedReducer, composeEnhancer(applyMiddleware(...middlewares)));
const persistor = persistStore(store);

ReactDOM.render(
  <React.StrictMode>
    <Provider store={store}>
      <PersistGate loading={null} persistor={persistor}>
        <ErrorBoundary>
          <TourProvider steps={steps}>
            <App /> 
          </TourProvider>
        </ErrorBoundary>
      </PersistGate>
    </Provider>
  </React.StrictMode>,
  document.getElementById('root')
);
