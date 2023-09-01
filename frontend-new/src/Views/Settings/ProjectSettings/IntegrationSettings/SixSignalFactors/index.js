import React, { useState } from 'react';
import { useEffect } from 'react';
import { connect } from 'react-redux';
import { fetchProjectSettings, udpateProjectSettings } from 'Reducers/global';
import { Modal, Button, message } from 'antd';
import { FaErrorComp, FaErrorLog } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import ConnectedScreen from './ConnectedScreen';
import useAgentInfo from 'hooks/useAgentInfo';
import { SolutionsAccountId } from 'Routes/constants';

function SixSignalFactorsIntegration({
  fetchProjectSettings,
  udpateProjectSettings,
  activeProject,
  currentProjectSettings,
  setIsActive,
  kbLink = false,
  currentAgent
}) {
  const [loading, setLoading] = useState(false);
  const { email: userEmail } = useAgentInfo();

  useEffect(() => {
    if (currentProjectSettings?.int_factors_six_signal_key) {
      setIsActive(true);
    }
  }, [currentProjectSettings, setIsActive]);

  const onConnect = () => {
    Modal.confirm({
      title:
        'Are you sure you want to connect Factors de-anonymization for this project?',
      content:
        'You are about to enable this integration. Factors will start bringing in data from this source.',
      okText: 'Connect',
      cancelText: 'Cancel',
      onOk: () => {
        setLoading(true);
        udpateProjectSettings(activeProject.id, {
          int_factors_six_signal_key: true,
          six_signal_config: {}
        })
          .then(() => {
            setLoading(false);
            setTimeout(() => {
              message.success('6Signal integration connected!');
            }, 500);
            setIsActive(false);
          })
          .catch((err) => {
            message.error(`${err?.data?.error}`);
            setLoading(false);
          });
      },
      onCancel: () => {}
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
          int_factors_six_signal_key: false,
          six_signal_config: {}
        })
          .then(() => {
            setLoading(false);
            setTimeout(() => {
              message.success('6Signal integration disconnected!');
            }, 500);
            setIsActive(false);
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
        <FaErrorComp subtitle='Facing issues with 6Signal Factors integrations' />
      }
      onError={FaErrorLog}
    >
      {currentProjectSettings?.int_factors_six_signal_key && (
        <ConnectedScreen />
      )}

      <div className='mt-4 flex' data-tour='step-11'>
        {userEmail === SolutionsAccountId && (
          <>
            {currentProjectSettings?.int_factors_six_signal_key ? (
              <Button loading={loading} onClick={() => onDisconnect()}>
                Disconnect
              </Button>
            ) : (
              <Button
                type='primary'
                loading={loading}
                onClick={() => onConnect()}
              >
                Connect Now
              </Button>
            )}
          </>
        )}

        {kbLink && (
          <a
            className='ant-btn ml-2 '
            target='_blank'
            href={kbLink}
            rel='noreferrer'
          >
            View documentation
          </a>
        )}
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
  udpateProjectSettings
})(SixSignalFactorsIntegration);
