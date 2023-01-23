import React from 'react';
import { Route, Switch, Redirect } from 'react-router-dom';
import lazyWithRetry from 'Utils/lazyWithRetry';
import PrivateRoute from 'Components/PrivateRoute';
import { APP_LAYOUT_ROUTES, APP_ROUTES } from './constants';
import { WhiteListedAccounts } from 'Routes/constants';
import { featureLock } from './feature';
import { ATTRIBUTION_ROUTES } from 'Attribution/utils/constants';
import SetupAssist from 'Views/Settings/SetupAssist';
const PathAnalysis = lazyWithRetry(() => import('../Views/PathAnalysis'));
const PathAnalysisReport = lazyWithRetry(() =>
  import('../Views/PathAnalysis/PathAnalysisReport')
);
const Attribution = lazyWithRetry(() => import('../features/attribution/ui'));

const renderRoutes = (routesObj) => {
  return Object.keys(routesObj)
    .map((routeName) => {
      const route = routesObj[routeName];
      if (!route) return null;
      const { Component, exact = false, path, Private } = route;
      if (!Component || !path) return null;

      if (Private) {
        return <PrivateRoute exact={exact} path={path} component={Component} />;
      }
      return <Route exact={exact} path={path} component={Component} />;
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
  return (
    <Switch>
      {renderRoutes(APP_LAYOUT_ROUTES)}

      {/* Additional Conditional routes  */}
      {featureLock(activeAgent) ? (
        <Route
          path={ATTRIBUTION_ROUTES.base}
          name='attribution'
          component={Attribution}
        />
      ) : null}

      {currentProjectSettings?.is_path_analysis_enabled && (
        <>
          <Route
            path='/path-analysis'
            name='PathAnalysis'
            exact
            component={PathAnalysis}
          />
          <Route
            path='/path-analysis/insights'
            name='PathAnalysisInsights'
            exact
            component={PathAnalysisReport}
          />
        </>
      )}

      {!demoProjectId.includes(active_project?.id) ? (
        <Route path='/project-setup' component={SetupAssist} />
      ) : (
        <Redirect to='/' />
      )}
    </Switch>
  );
};
