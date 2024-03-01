import React, { useState, useEffect, useRef } from 'react';
import { Row, Col, Skeleton, message } from 'antd';
import { Text, FaErrorComp, FaErrorLog } from 'factorsComponents';
import { connect } from 'react-redux';
import { fetchProjectSettings, fetchProjectSettingsV1 } from 'Reducers/global';

import { ErrorBoundary } from 'react-error-boundary';

import logger from 'Utils/logger';
import { fetchDashboards } from 'Reducers/dashboard/services';
import { fetchQueries } from 'Reducers/coreQuery/services';
import {
  ADWORDS_INTERNAL_REDIRECT_URI,
  createDashboardsFromTemplatesForRequiredIntegration
} from './util';
import { IntegrationProviderData } from './integrations.constants';
import IntegrationCard from './IntegrationCard';

function IntegrationSettings({
  activeProject,
  currentProjectSettings,
  currentProjectSettingsLoading,
  fetchProjectSettings,
  fetchProjectSettingsV1,
  fetchDashboards,
  fetchQueries,
  dashboards,
  dashboardTemplates,
  sdkCheck,
  bingAds,
  marketo
}) {
  const [dataLoading, setDataLoading] = useState(true);
  const templateDashboardStatusRef = useRef(false);

  // effect for creating dashboards from templates based on the integrations
  useEffect(() => {
    let timeout = false;

    const initializeTimeout = () => {
      // returning for unavailable values or loading states
      if (dashboards?.loading || !dashboards?.data) return;
      if (dashboardTemplates?.loading || !dashboardTemplates?.data) return;
      if (!activeProject?.id) return;
      if (currentProjectSettingsLoading) return;
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
            dashboards.data,
            dashboardTemplates.data,
            sdkCheck,
            currentProjectSettings,
            bingAds,
            marketo
          );
          if (res) {
            await fetchDashboards(activeProject.id);
            await fetchQueries(activeProject.id);
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
    dashboards,
    activeProject?.id,
    dashboardTemplates,
    sdkCheck,
    currentProjectSettings,
    bingAds,
    marketo,
    currentProjectSettingsLoading
  ]);

  useEffect(() => {
    fetchProjectSettings(activeProject.id).then(() => {
      setDataLoading(false);
    });
    fetchProjectSettingsV1(activeProject.id);
  }, [activeProject]);

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
      <div className='fa-container'>
        <Row gutter={[24, 24]} justify='center'>
          <Col span={22}>
            <div className='mb-10'>
              <Row>
                <Col span={12}>
                  <Text
                    type='title'
                    level={3}
                    weight='bold'
                    extraClass='m-0'
                    id='fa-at-text--page-title'
                  >
                    Integrations
                  </Text>
                </Col>
              </Row>
              <Row className='mt-4'>
                <Col span={24}>
                  {dataLoading ? (
                    <Skeleton active paragraph={{ rows: 4 }} />
                  ) : (
                    IntegrationProviderData.map((item, index) => {
                      let defaultOpen = false;

                      if (
                        window.location.href.indexOf(
                          ADWORDS_INTERNAL_REDIRECT_URI
                        ) > -1
                      ) {
                        defaultOpen = true;
                      }

                      return (
                        <IntegrationCard
                          integrationConfig={item}
                          key={item.name}
                          defaultOpen={defaultOpen}
                        />
                      );
                    })
                  )}
                </Col>
              </Row>
            </div>
          </Col>
        </Row>
      </div>
    </ErrorBoundary>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  currentProjectSettings: state.global.currentProjectSettings,
  currentProjectSettingsLoading: state.global.currentProjectSettingsLoading,
  currentAgent: state.agent.agent_details,
  dashboards: state.dashboard.dashboards,
  dashboardTemplates: state.dashboardTemplates.templates,
  sdkCheck: state?.global?.projectSettingsV1?.int_completed,
  bingAds: state?.global?.bingAds,
  marketo: state?.global?.marketo
});

export default connect(mapStateToProps, {
  fetchProjectSettings,
  fetchProjectSettingsV1,
  fetchDashboards,
  fetchQueries
})(IntegrationSettings);
