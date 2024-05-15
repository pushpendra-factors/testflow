import React, { useContext, useEffect, useRef } from 'react';
import { Alert, Divider, Skeleton, message } from 'antd';
import { ErrorBoundary } from 'react-error-boundary';
import { FaErrorComp, FaErrorLog } from 'Components/factorsComponents';
import { useHistory, useParams } from 'react-router-dom';
import { ConnectedProps, connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { fetchDashboards } from 'Reducers/dashboard/services';
import { fetchQueries } from 'Reducers/coreQuery/services';
import logger from 'Utils/logger';
import moment from 'moment';
import useIntegrationCheck from 'hooks/useIntegrationCheck';
import useAgentInfo from 'hooks/useAgentInfo';
import { IntegrationProviderData } from './integrations.constants';
import IntegrationHeader from './IntegrationHeader';
import IntegrationInstruction from './IntegrationInstruction';
import {
  createDashboardsFromTemplatesForRequiredIntegration,
  getIntegrationStatus,
  showIntegrationStatus
} from './util';
import { IntegrationContext } from './IntegrationContext';

const IntegrationWithId = ({
  currentProjectSettingsLoading,
  dashboardTemplates,
  activeProject,
  currentProjectSettings,
  fetchDashboards,
  fetchQueries
}: IntegrationWithIdProps) => {
  const { integration_id: integrationId } = useParams();
  const history = useHistory();
  const { email: userEmail } = useAgentInfo();
  const Integration = IntegrationProviderData.find(
    (integration) => integration.id === integrationId
  );
  const templateDashboardStatusRef = useRef(false);
  const {
    integrationStatus,
    fetchIntegrationStatus,
    integrationStatusLoading
  } = useContext(IntegrationContext);
  const integrationStatusValue = getIntegrationStatus(
    integrationStatus?.[integrationId]
  );
  const integrationInfo = useIntegrationCheck();
  const isIntegrated = integrationInfo?.[integrationId];
  const showIntegrationStatusFlag = showIntegrationStatus(userEmail);

  const integrationStatusMessage = integrationStatus?.[integrationId]?.message;
  const lastSyncDetail =
    showIntegrationStatusFlag &&
    integrationStatus?.[integrationId]?.last_synced_at &&
    isIntegrated
      ? `Last sync: ${moment
          .unix(integrationStatus?.[integrationId]?.last_synced_at)
          .fromNow()}`
      : '';

  const isErrorState = integrationStatusValue === 'error';

  const handleBackClick = () => {
    sessionStorage.setItem('integration-card', integrationId);
    history.goBack();
  };

  const integrationCallback = () => {
    if (fetchIntegrationStatus) {
      fetchIntegrationStatus();
    }
  };

  // effect for creating dashboards from templates based on the integrations
  useEffect(() => {
    let timeout = false;

    const initializeTimeout = () => {
      if (currentProjectSettingsLoading) return;
      if (
        dashboardTemplates?.loading ||
        !dashboardTemplates?.data ||
        !dashboardTemplates?.data?.length
      )
        return;
      // do nothing if dashboard creation is in process
      if (templateDashboardStatusRef.current) return;
      // setting up a timer so the latest values can be used
      timeout = setTimeout(async () => {
        try {
          // do nothing if dashboard creation is in process
          if (templateDashboardStatusRef.current) return;
          templateDashboardStatusRef.current = true;

          const res = await createDashboardsFromTemplatesForRequiredIntegration(
            activeProject.id,
            dashboardTemplates?.data,
            currentProjectSettings
          );
          if (res) {
            fetchDashboards(activeProject.id);
            fetchQueries(activeProject.id);
          }
          // setting template dashboard status back to false
          setTimeout(() => {
            templateDashboardStatusRef.current = false;
          }, 0);
        } catch (error) {
          logger.error('Error in creating dashboard from template', error);
          templateDashboardStatusRef.current = false;
        }
      }, 2000);
    };

    initializeTimeout();

    return () => {
      if (timeout) clearTimeout(timeout);
    };
  }, [
    activeProject?.id,
    currentProjectSettingsLoading,
    currentProjectSettings
  ]);

  useEffect(() => {
    if (window.location.href.indexOf('?error=') > -1) {
      const searchParams = new URLSearchParams(window.location.search);
      if (searchParams) {
        const error = searchParams.get('error');
        const str = error.replace('_', ' ');
        const finalmsg = str.toLocaleLowerCase();
        if (finalmsg) {
          message.error(finalmsg);
        }
      }
    }

    if (window.location.href.indexOf('status=') > -1) {
      const searchParams = new URLSearchParams(window.location.search);
      if (searchParams) {
        const error = searchParams.get('status');
        const str = error.replace('_', ' ');
        const finalmsg = str.toLocaleLowerCase();
        if (finalmsg) {
          message.error(
            `Error: ${finalmsg}. Sorry! That doesn’t seem right. Please try again`
          );
        }
      }
    }
  }, []);
  if (integrationStatusLoading) {
    return (
      <>
        <Skeleton />
        <Skeleton />
        <Skeleton />
        <Skeleton />
      </>
    );
  }

  if (Integration?.id && Integration?.Component) {
    return (
      <ErrorBoundary
        fallback={
          <FaErrorComp
            size='medium'
            title='Integrations Error'
            subtitle='We are facing some issues with the integrations. Drop us a message on the in-app chat.'
          />
        }
        onError={FaErrorLog}
      >
        <div>
          <IntegrationHeader
            handleBackClick={handleBackClick}
            title={Integration.name}
            description={Integration.desc}
            iconText={Integration.icon}
            lastSyncDetail={lastSyncDetail}
          />
          <Divider style={{ margin: '16px 0px' }} />
          {showIntegrationStatusFlag && isIntegrated && isErrorState && (
            <Alert message={integrationStatusMessage} type='error' showIcon />
          )}
          {Integration.showInstructionMenu && (
            <IntegrationInstruction
              title={Integration.instructionTitle}
              description={Integration.instructionDescription}
              kbLink={Integration.kbLink}
            />
          )}

          <Integration.Component integrationCallback={integrationCallback} />
          <Divider style={{ margin: '16px 0px' }} />
        </div>
      </ErrorBoundary>
    );
  }
  return (
    <div className='flex h-full w-full items-center justify-center'>
      <p> Integration not Found!</p>
    </div>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  currentProjectSettings: state.global.currentProjectSettings,
  currentProjectSettingsLoading: state.global.currentProjectSettingsLoading,
  currentAgent: state.agent.agent_details,
  dashboardTemplates: state.dashboardTemplates.templates
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      fetchDashboards,
      fetchQueries
    },
    dispatch
  );

const connector = connect(mapStateToProps, mapDispatchToProps);
type IntegrationWithIdProps = ConnectedProps<typeof connector>;

export default connector(IntegrationWithId);
