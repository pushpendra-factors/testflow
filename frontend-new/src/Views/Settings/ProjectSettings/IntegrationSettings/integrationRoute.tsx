import React, {
  Suspense,
  useCallback,
  useEffect,
  useMemo,
  useState
} from 'react';
import { Switch, Route, useRouteMatch } from 'react-router-dom';
import { ErrorBoundary } from 'react-error-boundary';
import { FaErrorComp, FaErrorLog } from 'factorsComponents';
import PageSuspenseLoader from 'Components/SuspenseLoaders/PageSuspenseLoader';

import { fetchProjectSettings, fetchProjectSettingsV1 } from 'Reducers/global';
import { ConnectedProps, connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import logger from 'Utils/logger';
import IntegrationWithId from './IntegrationWithId';
import BaseComponent from './index';
import {
  IntegrationContext,
  defaultIntegrationContextData
} from './IntegrationContext';
import { getIntegrationStatus } from './service';
import { IntegrationStatusData } from './types';

function IntegrationMain({
  activeProject,
  fetchProjectSettings,
  fetchProjectSettingsV1
}: IntegrationRouteProps) {
  const { path } = useRouteMatch();
  const [contextData, setContextData] = useState({
    ...defaultIntegrationContextData,
    dataLoading: true,
    integrationStatusLoading: true
  });

  useEffect(() => {
    fetchProjectSettings(activeProject.id).then(() => {
      setContextData((cData) => ({ ...cData, dataLoading: false }));
    });
    fetchProjectSettingsV1(activeProject.id);
  }, []);

  // fetching integration status data

  const fetchIntegrationStatus = useCallback(async () => {
    try {
      setContextData((cData) => ({
        ...cData,
        integrationStatusLoading: true
      }));
      const res = await getIntegrationStatus(activeProject.id);
      const integrationStatusData = res?.data as IntegrationStatusData;
      setContextData((cData) => ({
        ...cData,
        integrationStatus: integrationStatusData,
        integrationStatusLoading: false
      }));
    } catch (error) {
      logger.error('Error in fetching integration stagus', error);
    }
  }, []);

  useEffect(() => {
    fetchIntegrationStatus();
  }, []);

  const ContextData = useMemo(
    () => ({
      ...contextData,
      fetchIntegrationStatus
    }),
    [contextData, fetchIntegrationStatus]
  );

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
        <IntegrationContext.Provider value={ContextData}>
          <Switch>
            <Route exact path={`${path}`} component={BaseComponent} />
            <Route
              exact
              path={`${path}/:integration_id`}
              component={IntegrationWithId}
            />
          </Switch>
        </IntegrationContext.Provider>
      </Suspense>
    </ErrorBoundary>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      fetchProjectSettings,
      fetchProjectSettingsV1
    },
    dispatch
  );

const connector = connect(mapStateToProps, mapDispatchToProps);
type IntegrationRouteProps = ConnectedProps<typeof connector>;

export default connector(IntegrationMain);
