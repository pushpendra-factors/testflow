import React, { Suspense, useEffect } from 'react';
import { Switch, Route, useRouteMatch } from 'react-router-dom';
import { ErrorBoundary } from 'react-error-boundary';
import { FaErrorComp, FaErrorLog } from 'factorsComponents';
import PageSuspenseLoader from 'Components/SuspenseLoaders/PageSuspenseLoader';

import lazyWithRetry from 'Utils/lazyWithRetry';
import { useDispatch, useSelector } from 'react-redux';
import {
  fetchAttributionActiveUnits,
  fetchAttributionDashboard
} from 'Attribution/state/services';

const BaseComponent = lazyWithRetry(() => import('./baseComponent'));

const Report = lazyWithRetry(() => import('./report'));
const Reports = lazyWithRetry(() => import('./reports'));

function Attribution() {
  const { path } = useRouteMatch();
  const dispatch = useDispatch();

  const { dashboard } = useSelector((state) => state.attributionDashboard);
  const { active_project: activeProject } = useSelector(
    (state) => state.global
  );
  useEffect(() => {
    if (activeProject?.id) {
      dispatch(fetchAttributionDashboard(activeProject.id));
    }
  }, [activeProject?.id]);

  useEffect(() => {
    if (dashboard?.id && activeProject?.id) {
      dispatch(fetchAttributionActiveUnits(activeProject.id, dashboard.id));
    }
  }, [dashboard, activeProject?.id]);
  return (
    <ErrorBoundary
      fallback={
        <FaErrorComp
          size='medium'
          title='Bundle Error'
          subtitle='We are facing trouble loading App Bundles. Drop us a message on the in-app chat.'
        />
      }
      onError={FaErrorLog}
    >
      <Suspense fallback={<PageSuspenseLoader />}>
        <Switch>
          <Route
            exact
            path={`${path}`}
            name='root-attribution'
            component={BaseComponent}
          />
          <Route
            name='attribution-report'
            exact
            path={`${path}/report`}
            component={Report}
          />
          <Route
            name='attribution-reports'
            exact
            path={`${path}/reports`}
            component={Reports}
          />
        </Switch>
      </Suspense>
    </ErrorBoundary>
  );
}

export default Attribution;
