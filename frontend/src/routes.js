import React from 'react';
import Loadable from 'react-loadable';
import { Redirect } from 'react-router-dom';

import DefaultLayout from './containers/DefaultLayout';
import Loading from './loading';

const Query = Loadable({
  loader: () => import('./views/Query'),
  loading: Loading,
});

const Factor = Loadable({
  loader: () => import('./views/Factor'),
  loading: Loading,
});

const Settings = Loadable({
  loader: () => import('./views/Settings'),
  loading: Loading,
})

const SettingsSegment =  Loadable({
  loader: () => import('./views/Settings/Segment'),
  loading: Loading,
});

const SettingsJsSdk = Loadable({
  loader: () => import('./views/Settings/JsSdk'),
  loading: Loading,
});

const SettingsAndroidSdk = Loadable({
  loader: () => import('./views/Settings/AndroidSdk'),
  loading: Loading,
});

const SettingsIosSdk = Loadable({
  loader: () => import('./views/Settings/IosSdk'),
  loading: Loading,
});

const SettingsAutoTrack = Loadable({
  loader: () => import('./views/Settings/AutoTrack'),
  loading: Loading,
});

const ReloadComponent = (props) => {
  // Todo(Dinesh): Fix browser forward after go(-1).
  props.history.go(-1);
  return "";
}

// https://github.com/ReactTraining/react-router/tree/master/packages/react-router-config
const routes = [
  { path: '/', exact: true, name: 'Home', component: DefaultLayout },
  { path: '/core', name: 'Query', component: Query },
  { path: '/factor', name: 'Factor', component: Factor },
  { path: '/settings/segment', exact: true, name: 'Segment', component: SettingsSegment },
  { path: '/settings/autotrack', exact: true, name: 'AutoTrack', component: SettingsAutoTrack },
  { path: '/settings/jssdk', exact: true, name: 'JsSdk', component: SettingsJsSdk },
  { path: '/settings/androidsdk', exact: true, name: 'AndroidSdk', component: SettingsAndroidSdk },
  { path: '/settings/iossdk', exact: true, name: 'IosSdk', component: SettingsIosSdk },
  { path: '/settings', name: 'Settings', component: Settings },
  { path: '/refresh', exact: true, name: 'Refresh', component: ReloadComponent },
];

export default routes;
