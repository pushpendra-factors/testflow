import React, { useState } from 'react';
import { connect } from 'react-redux';
import {
  fetchProjectSettings,
  udpateProjectSettings,
  enableHubspotIntegration
} from 'Reducers/global';
import {
  Row,
  Col,
  Modal,
  Input,
  Form,
  Button,
  notification,
  message
} from 'antd';
import { Text, FaErrorComp, FaErrorLog } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import factorsai from 'factorsai';
import { sendSlackNotification } from '../../../../../utils/slack';

const HubspotIntegration = ({
  fetchProjectSettings,
  udpateProjectSettings,
  activeProject,
  currentProjectSettings,
  currentAgent,
  enableHubspotIntegration,
  integrationCallback
}) => {
  const [errorInfo, seterrorInfo] = useState(null);
  const [loading, setLoading] = useState(false);

  const isHubspotEnabled = () =>
    currentProjectSettings &&
    currentProjectSettings?.int_hubspot &&
    currentProjectSettings?.int_hubspot_refresh_token != '';

  const onClickEnableHubspot = () => {
    setLoading(true);

    // Factors INTEGRATION tracking
    factorsai.track('INTEGRATION', {
      name: 'hubspot',
      activeProjectID: activeProject.id
    });

    enableHubspotIntegration(activeProject.id)
      .then((r) => {
        setLoading(false);
        sendSlackNotification(
          currentAgent.email,
          activeProject.name,
          'Hubspot'
        );
        if (r.status == 307) {
          window.location = r.data.redirect_url;
        }
      })
      .catch((err) => {
        setLoading(false);
        message.error(`${err?.data?.error}`);
      });
  };

  const onDisconnect = () => {
    Modal.confirm({
      title: 'Are you sure you want to disable this?',
      content:
        'You are about to disable this integration. Factors will stop bringing in data from this source.',
      okText: 'Disconnect',
      cancelText: 'Cancel',
      onOk: () => {
        setLoading(true);
        udpateProjectSettings(activeProject.id, {
          int_hubspot_api_key: '',
          int_hubspot: false
        })
          .then(() => {
            setLoading(false);
            setTimeout(() => {
              message.success('Hubspot integration disconnected!');
            }, 500);
            integrationCallback();
          })
          .catch((err) => {
            message.error(`${err?.data?.error}`);
            setLoading(false);
          });
      },
      onCancel: () => {}
    });
  };

  const isEnabled = isHubspotEnabled();

  return (
    <ErrorBoundary
      fallback={
        <FaErrorComp subtitle='Facing issues with Hubspot integrations' />
      }
      onError={FaErrorLog}
    >
      <div className='mt-4 flex' data-tour='step-11'>
        {isEnabled ? (
          <Button loading={loading} onClick={() => onDisconnect()}>
            Disconnect
          </Button>
        ) : (
          <Button
            type='primary'
            loading={loading}
            onClick={onClickEnableHubspot}
          >
            Enable using Hubspot
          </Button>
        )}
      </div>
    </ErrorBoundary>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  currentProjectSettings: state.global.currentProjectSettings,
  currentAgent: state.agent.agent_details
});

export default connect(mapStateToProps, {
  fetchProjectSettings,
  udpateProjectSettings,
  enableHubspotIntegration
})(HubspotIntegration);
