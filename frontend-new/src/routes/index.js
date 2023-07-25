import React, { useEffect } from 'react';
import { Route, Switch, Redirect } from 'react-router-dom';
import lazyWithRetry from 'Utils/lazyWithRetry';
import PrivateRoute from 'Components/PrivateRoute';
import { APP_LAYOUT_ROUTES, APP_ROUTES } from './constants';
import { AdminLock, featureLock } from './feature';
import { ATTRIBUTION_ROUTES } from 'Attribution/utils/constants';
import SetupAssist from 'Views/Settings/SetupAssist';
import { useDispatch } from 'react-redux';
import { UPDATE_ALL_ROUTES } from 'Reducers/types';
import OnBoard from 'Views/Settings/SetupAssist/Welcome/OnboardFlow';
import withFeatureLockHOC from 'HOC/withFeatureLock';
import { FEATURES } from 'Constants/plans.constants';
import LockedStateComponent from 'Components/GenericComponents/LockedStateVideoComponent';
import { PathUrls } from './pathUrls';
import ConfigurePlans from 'Views/Settings/ProjectSettings/ConfigurePlans';

const Attribution = lazyWithRetry(() => import('../features/attribution/ui'));
const FeatureLockedAttributionComponent = withFeatureLockHOC(Attribution, {
  featureName: FEATURES.FEATURE_ATTRIBUTION,
  LockedComponent: () => (
    <LockedStateComponent
      title={'Attribution'}
      description='Attribute revenue and conversions to the right marketing channels, campaigns, and touchpoints to gain a clear understanding of what drives success. Identify the most effective marketing strategies, optimize your budget allocation, and make data-driven decisions to maximize ROI and achieve your business goals.'
    />
  )
});

const renderRoutes = (routesObj) => {
  return Object.keys(routesObj)
    .map((routeName) => {
      const route = routesObj[routeName];

      if (!route) return null;
      const { Component, exact = false, path, Private } = route;

      if (!Component || !path) return null;

      if (Private) {
        return (
          <PrivateRoute
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
};

export const AppRoutes = () => (
  <Switch>
    {renderRoutes(APP_ROUTES)}

    {/* If no routes match */}
    <Route>
      <Redirect to='/' />
    </Route>
  </Switch>
);

export const AppLayoutRoutes = ({
  activeAgent,
  demoProjectId,
  active_project,
  currentProjectSettings
}) => {
  const dispatch = useDispatch();

  useEffect(() => {
    if (featureLock(activeAgent)) {
      let allRoutes = [];
      allRoutes.push(ATTRIBUTION_ROUTES.base);

      dispatch({ type: UPDATE_ALL_ROUTES, payload: allRoutes });
    }
  }, [activeAgent]);
  useEffect(() => {
    let allRoutes = [];
    for (let obj of Object.keys(APP_LAYOUT_ROUTES)) {
      allRoutes.push(APP_LAYOUT_ROUTES[obj].path);
    }

    dispatch({ type: UPDATE_ALL_ROUTES, payload: allRoutes });
  }, []);
  return (
    <Switch>
      {renderRoutes(APP_LAYOUT_ROUTES)}
      {/* Additional Conditional routes  */}

      <PrivateRoute
        path={ATTRIBUTION_ROUTES.base}
        name='attribution'
        component={FeatureLockedAttributionComponent}
      />

      {AdminLock(activeAgent) ? (
        <PrivateRoute
          path={PathUrls.ConfigurePlans}
          name='Configure Plans'
          component={ConfigurePlans}
        />
      ) : null}

      {!demoProjectId.includes(active_project?.id) ? (
        <PrivateRoute path='/project-setup' component={SetupAssist} />
      ) : (
        <Redirect to='/' />
      )}
    </Switch>
  );
};
