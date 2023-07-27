import React, { useState, useEffect } from 'react';
import { Row, Col, Skeleton, message } from 'antd';
import { Text, FaErrorComp, FaErrorLog } from 'factorsComponents';
import { connect } from 'react-redux';
import { fetchProjectSettings, fetchProjectSettingsV1 } from 'Reducers/global';

import { ErrorBoundary } from 'react-error-boundary';

import { ADWORDS_INTERNAL_REDIRECT_URI } from './util';
import { featureLock } from '../../../../routes/feature';
import { IntegrationProviderData } from './integrations.constants';
import IntegrationCard from './IntegrationCard';

function IntegrationSettings({
  currentProjectSettings,
  activeProject,
  fetchProjectSettings,
  currentAgent,
  fetchProjectSettingsV1
}) {
  const [dataLoading, setDataLoading] = useState(true);

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
          <Col span={18}>
            <div className='mb-10 pl-4'>
              <Row>
                <Col span={12}>
                  <Text type='title' level={3} weight='bold' extraClass='m-0'>
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
  currentAgent: state.agent.agent_details
});

export default connect(mapStateToProps, {
  fetchProjectSettings,
  fetchProjectSettingsV1
})(IntegrationSettings);
