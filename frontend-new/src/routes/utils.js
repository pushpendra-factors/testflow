import React from 'react';

import PrivateRoute from 'Components/PrivateRoute';
import { Route } from 'react-router-dom';

export const renderRoutes = (routesObj) =>
  Object.keys(routesObj)
    .map((routeName) => {
      const route = routesObj[routeName];

      if (!route) return null;
      const { Component, exact = false, path, Private, title } = route;

      if (!Component || !path) return null;

      if (Private) {
        return (
          <PrivateRoute
            title={title}
            exact={exact}
            path={path}
            component={Component}
            key={path}
          />
        );
      }
      return (
        <Route exact={exact} path={path} component={Component} key={path} />
      );
    })
    .filter((route) => !!route);
