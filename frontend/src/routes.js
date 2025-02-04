import Loadable from 'react-loadable';
import React from 'react';
import { Redirect } from 'react-router-dom';

import DefaultLayout from './containers/DefaultLayout';
import Loading from './loading';

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

const Settings = Loadable({
  loader: () => import('./views/Settings'),
  loading: Loading,
})

const SettingsSegment =  Loadable({
  loader: () => import('./views/Settings/Segment'),
  loading: Loading,
});

const SettingsAdwords =  Loadable({
  loader: () => import('./views/Settings/Adwords'),
  loading: Loading,
});

const SettingsHubspot =  Loadable({
  loader: () => import('./views/Settings/Hubspot'),
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

const SettingsFacebook = Loadable({
  loader: () => import('./views/Settings/Facebook'),
  loading: Loading,
});

const SettingsSalesforce = Loadable({
  loader: () => import('./views/Settings/Salesforce'),
  loading: Loading,
});

const SettingsLinkedin = Loadable({
  loader: () => import('./views/Settings/Linkedin'),
  loading: Loading,
}); 

const Profile = Loadable({
  loader: () => import('./views/Profile'),
  loading: Loading,
});

const SettingsAgents = Loadable({
  loader: () => import('./views/Settings/Agents'),
  loading: Loading,
});

const ReloadComponent = (props) => {
  // Not using props history to avoid multiple backs.
  // props.history.go(-1);
  // return "";
  
  return <Redirect to='/dashboard' />
}

const AccountSettings = Loadable({
  loader: () => import('./views/AccountSettings'),
  loading: Loading,
})

// https://github.com/ReactTraining/react-router/tree/master/packages/react-router-config
const routes = [
  { path: '/', exact: true, name: 'Home', component: DefaultLayout },
  { path: '/dashboard', name: 'Dashboard', component: Dashboard },
  { path: '/core', name: 'Query', component: Query },
  { path: '/factor', name: 'Factor', component: Factor },
  { path: '/settings/segment', exact: true, name: 'Segment', component: SettingsSegment },
  { path: '/settings/adwords', exact: true, name: 'Adwords', component: SettingsAdwords },
  { path: '/settings/hubspot', exact: true, name: 'Hubspot', component: SettingsHubspot },
  { path: '/settings/jssdk', exact: true, name: 'JsSdk', component: SettingsJsSdk },
  { path: '/settings/androidsdk', exact: true, name: 'AndroidSdk', component: SettingsAndroidSdk },
  { path: '/settings/iossdk', exact: true, name: 'IosSdk', component: SettingsIosSdk },
  { path: '/settings/agents', exact: true, name: 'Agents', component: SettingsAgents },
  { path: '/settings/facebook', exact: true, name: 'Facebook', component: SettingsFacebook },
  { path: '/settings/Salesforce', exact: true, name: 'Salesforce', component: SettingsSalesforce },
  { path: '/settings/linkedin', exact: true, name: 'Linkedin', component: SettingsLinkedin },
  { path: '/settings', name: 'Settings', component: Settings },
  { path: '/account_settings', name: 'AccountSettings', component: AccountSettings },
  { path: '/profile', name: 'Profile', component: Profile },
  { path: '/refresh', exact: true, name: 'Refresh', component: ReloadComponent },
];

// routes only for email@factors.ai.
const internalRoutes = [];

export  {
  routes,
  internalRoutes
};
