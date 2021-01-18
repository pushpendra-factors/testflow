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
import { persistStore, persistReducer } from 'redux-persist';
import storage from 'redux-persist/lib/storage'; // defaults to localStorage for web
import { PersistGate } from 'redux-persist/integration/react';
import ErrorBoundary from './ErrorBoundary';
import * as Sentry from "@sentry/react";
import { Integrations } from "@sentry/tracing";



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

  // <!-- Begin Inspectlet Asynchronous Code -->
    window.__insp = window.__insp || [];
    __insp.push(['wid', 1994835818]);
    var ldinsp = function(){
    if(typeof window.__inspld != "undefined") return; window.__inspld = 1; var insp = document.createElement('script'); insp.type = 'text/javascript'; insp.async = true; insp.id = "inspsync"; insp.src = ('https:' == document.location.protocol ? 'https' : 'http') + '://cdn.inspectlet.com/inspectlet.js?wid=1994835818&r=' + Math.floor(new Date().getTime()/3600000); var x = document.getElementsByTagName('script')[0]; x.parentNode.insertBefore(insp, x); };
    setTimeout(ldinsp, 0);
  // <!-- End Inspectlet Asynchronous Code -->
}



const persistConfig = {
  key: 'root',
  storage,
  whitelist: ['global', 'agent', 'factors']
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
            <App /> 
        </ErrorBoundary>
      </PersistGate>
    </Provider>
  </React.StrictMode>,
  document.getElementById('root')
);
