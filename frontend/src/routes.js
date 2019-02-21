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

const Users = Loadable({
  loader: () => import('./views/Users/Users'),
  loading: Loading,
});

const User = Loadable({
  loader: () => import('./views/Users/User'),
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
  { path: '/settings', name: 'Settings', component: Settings },
  { path: '/users', exact: true,  name: 'Users', component: Users },
  { path: '/users/:id', exact: true, name: 'User Details', component: User },
  { path: '/refresh', exact: true, name: 'Refresh', component: ReloadComponent },
];

export default routes;
