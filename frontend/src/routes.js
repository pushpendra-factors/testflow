import React from 'react';
import Loadable from 'react-loadable'

import DefaultLayout from './containers/DefaultLayout';

function Loading() {
  return <div>Loading...</div>;
}

const Dashboard = Loadable({
  loader: () => import('./views/Dashboard'),
  loading: Loading,
});

const Query = Loadable({
  loader: () => import('./views/Query'),
  loading: Loading,
});

const Factor = Loadable({
  loader: () => import('./views/Factor'),
  loading: Loading,
});

const AutoTrack = Loadable({
  loader: () => import('./views/AutoTrack'),
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


// https://github.com/ReactTraining/react-router/tree/master/packages/react-router-config
const routes = [
  { path: '/', exact: true, name: 'Home', component: DefaultLayout },
  { path: '/dashboard', name: 'Dashboard', component: Dashboard },
  { path: '/query', name: 'Query', component: Query },
  { path: '/factor', name: 'Factor', component: Factor },
  { path: '/auto-track', name: 'AutoTrack', component: AutoTrack },
  { path: '/users', exact: true,  name: 'Users', component: Users },
  { path: '/users/:id', exact: true, name: 'User Details', component: User },
];

export default routes;
