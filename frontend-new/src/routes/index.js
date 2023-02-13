import React, { useEffect } from 'react';
import { Route, Switch, Redirect } from 'react-router-dom';
import lazyWithRetry from 'Utils/lazyWithRetry';
import PrivateRoute from 'Components/PrivateRoute';
import { APP_LAYOUT_ROUTES, APP_ROUTES } from './constants';
import { WhiteListedAccounts } from 'Routes/constants';
import { featureLock } from './feature';
import { ATTRIBUTION_ROUTES } from 'Attribution/utils/constants';
import SetupAssist from 'Views/Settings/SetupAssist';
import { useDispatch } from 'react-redux';
import { UPDATE_ALL_ROUTES } from 'Reducers/types';
const PathAnalysis = lazyWithRetry(() => import('../Views/PathAnalysis'));
const PathAnalysisReport = lazyWithRetry(() =>
  import('../Views/PathAnalysis/PathAnalysisReport')
);
const Attribution = lazyWithRetry(() => import('../features/attribution/ui'));
const SixSignalReport = lazyWithRetry(() =>
  import('../features/6signal-report/ui')
);

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
    if (currentProjectSettings.is_path_analysis_enabled) {
      let allRoutes = [];
      allRoutes.push('/path-analysis');
      allRoutes.push('/path-analysis/insights');
      dispatch({ type: UPDATE_ALL_ROUTES, payload: allRoutes });
    }
  }, [currentProjectSettings.is_path_analysis_enabled]);

  useEffect(() => {
    if (featureLock(activeAgent)) {
      let allRoutes = [];
      allRoutes.push('/reports/6_signal');
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
      {featureLock(activeAgent) ? (
        <PrivateRoute
          path={ATTRIBUTION_ROUTES.base}
          name='attribution'
          component={Attribution}
        />
      ) : null}

      {featureLock(activeAgent) ? (
        <Route
          path='/reports/6_signal'
          name='6-signal-report'
          component={SixSignalReport}
        />
      ) : null}

      {currentProjectSettings?.is_path_analysis_enabled && (
        <>
          <PrivateRoute
            path='/path-analysis'
            name='PathAnalysis'
            exact
            component={PathAnalysis}
          />
          <PrivateRoute
            path='/path-analysis/insights'
            name='PathAnalysisInsights'
            exact
            component={PathAnalysisReport}
          />
        </>
      )}

      {!demoProjectId.includes(active_project?.id) ? (
        <PrivateRoute path='/project-setup' component={SetupAssist} />
      ) : (
        <Redirect to='/' />
      )}
    </Switch>
  );
};
