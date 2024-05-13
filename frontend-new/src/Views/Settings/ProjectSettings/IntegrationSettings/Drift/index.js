import React, { useState } from 'react';
import { connect } from 'react-redux';
import { fetchProjectSettings, udpateProjectSettings } from 'Reducers/global';
import { Button, message, Modal } from 'antd';
import { FaErrorComp, FaErrorLog } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import factorsai from 'factorsai';
import { sendSlackNotification } from '../../../../../utils/slack';

const DriftIntegration = ({
  fetchProjectSettings,
  udpateProjectSettings,
  activeProject,
  currentProjectSettings,
  currentAgent
}) => {
  const [loading, setLoading] = useState(false);

  const enableDrift = () => {
    setLoading(true);

    // Factors INTEGRATION tracking
    factorsai.track('INTEGRATION', {
      name: 'drift',
      activeProjectID: activeProject.id
    });

    udpateProjectSettings(activeProject.id, { int_drift: true })
      .then(() => {
        setLoading(false);
        setTimeout(() => {
          message.success('Drift integration enabled!');
        }, 500);
        sendSlackNotification(currentAgent.email, activeProject.name, 'Drift');
      })
      .catch((err) => {
        setLoading(false);
        console.log('change password failed-->', err);
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
        udpateProjectSettings(activeProject.id, { int_drift: false })
          .then(() => {
            setLoading(false);
            setTimeout(() => {
              message.success('Drift integration disabled!');
            }, 500);
          })
          .catch((err) => {
            message.error(`${err?.data?.error}`);
            setLoading(false);
          });
      },
      onCancel: () => {}
    });
  };

  return (
    <ErrorBoundary
      fallback={
        <FaErrorComp subtitle='Facing issues with Facebook integrations' />
      }
      onError={FaErrorLog}
    >
      <div className='mt-4 flex'>
        {currentProjectSettings?.int_drift ? (
          <Button loading={loading} onClick={() => onDisconnect()}>
            Disable
          </Button>
        ) : (
          <Button type='primary' loading={loading} onClick={enableDrift}>
            Enable Now
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
  udpateProjectSettings
})(DriftIntegration);
