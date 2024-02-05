import React from 'react';

import lazyWithRetry from 'Utils/lazyWithRetry';
import { Redirect, Route, Switch } from 'react-router-dom';
import { renderRoutes } from './utils';

const Login = lazyWithRetry(() => import('../Views/Pages/Login'));
const ForgotPassword = lazyWithRetry(
  () => import('../Views/Pages/ForgotPassword')
);
const ResetPassword = lazyWithRetry(
  () => import('../Views/Pages/ResetPassword')
);
const SignUp = lazyWithRetry(() => import('../Views/Pages/SignUp'));
const Activate = lazyWithRetry(() => import('../Views/Pages/Activate'));
const Templates = lazyWithRetry(
  () => import('../Views/CoreQuery/Templates/ResultsPage')
);

const AppLayout = lazyWithRetry(() => import('../Views/AppLayout'));

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
