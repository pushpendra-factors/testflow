import React from 'react';

import lazyWithRetry from 'Utils/lazyWithRetry';
import { Redirect, Route, Switch } from 'react-router-dom';
import { renderRoutes } from './utils';

const Login = lazyWithRetry(
  () => import(/* webpackChunkName: "login" */ '../Views/Pages/Login')
);
const SingleSignOn = lazyWithRetry(
  () => import(/* webpackChunkName: "login" */ '../Views/Pages/SingleSignOn')
);
const ForgotPassword = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "forgot-password" */ '../Views/Pages/ForgotPassword'
    )
);
const ResetPassword = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "reset-password" */ '../Views/Pages/ResetPassword'
    )
);
const SignUp = lazyWithRetry(
  () => import(/* webpackChunkName: "signup" */ '../Views/Pages/SignUp')
);
const Activate = lazyWithRetry(
  () => import(/* webpackChunkName: "activate" */ '../Views/Pages/Activate')
);
const Templates = lazyWithRetry(
  () =>
    import(
      /* webpackChunkName: "templates" */ '../Views/CoreQuery/Templates/ResultsPage'
    )
);

const AppLayout = lazyWithRetry(
  () => import(/* webpackChunkName: "main-app" */ '../Views/AppLayout')
);

export const APP_ROUTES = {
  Signup: {
    path: '/signup',
    Component: SignUp,
    exact: true
  },
  Activate: {
    path: '/activate',
    Component: Activate,
    exact: true
  },
  SetPassword: {
    path: '/setpassword',
    Component: ResetPassword,
    exact: true
  },
  ForgotPassword: {
    path: '/forgotpassword',
    Component: ForgotPassword,
    exact: true
  },
  Login: {
    title: 'Login',
    path: '/login',
    Component: Login,
    exact: true
  },
  SingleSignOn: {
    title: 'Single Sign-On',
    path: '/sso',
    Component: SingleSignOn,
    exact: true
  },

  Templates: {
    path: '/templates',
    Component: Templates,
    Private: true,
    exact: true
  },
  APPLayout: {
    path: '/',
    Component: AppLayout
  }
};

export function AppRoutes() {
  return (
    <Switch>
      {renderRoutes(APP_ROUTES)}

      {/* If no routes match */}
      <Route>
        <Redirect to='/' />
      </Route>
    </Switch>
  );
}
