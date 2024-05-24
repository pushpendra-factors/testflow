import React, { useState, useRef } from 'react';
import { connect } from 'react-redux';
import { fetchProjectSettings, udpateProjectSettings } from 'Reducers/global';
import { Input, Button, message, Modal, notification } from 'antd';
import { Text, FaErrorComp, FaErrorLog } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import factorsai from 'factorsai';
import logger from 'Utils/logger';
import { sendSlackNotification } from '../../../../../utils/slack';

function SegmentIntegration({
  fetchProjectSettings,
  udpateProjectSettings,
  activeProject,
  currentProjectSettings,
  currentAgent,
  integrationCallback
}) {
  const [loading, setLoading] = useState(false);
  const textAreaRef = useRef(null);

  currentProjectSettings =
    currentProjectSettings?.project_settings || currentProjectSettings;

  const copyToClipboard = async () => {
    textAreaRef.current.select();
    try {
      await navigator.clipboard.writeText(activeProject?.private_token);
      notification.success({
        message: 'Success',
        description: 'Successfully copied!',
        duration: 3
      });
    } catch (err) {
      notification.error({
        message: 'Failed!',
        description: 'Failed to copy!',
        duration: 3
      });
    }
  };

  const enableSegment = () => {
    setLoading(true);

    // Factors INTEGRATION tracking
    factorsai.track('INTEGRATION', {
      name: 'segment',
      activeProjectID: activeProject.id
    });

    udpateProjectSettings(activeProject.id, {
      int_segment: true
    })
      .then(() => {
        fetchProjectSettings(activeProject.id);
        setLoading(false);
        setTimeout(() => {
          message.success('Segment integration enabled!');
        }, 500);
        sendSlackNotification(
          currentAgent.email,
          activeProject.name,
          'Segment'
        );
        integrationCallback();
      })
      .catch((err) => {
        setLoading(false);
        message.error(`${err?.data?.error}`);
        logger.error('change password failed-->', err);
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
          int_segment: false
        })
          .then(() => {
            fetchProjectSettings(activeProject.id);
            setLoading(false);
            setTimeout(() => {
              message.success('Segment integration disabled!');
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

  return (
    <ErrorBoundary
      fallback={
        <FaErrorComp subtitle='Facing issues with Segment integrations' />
      }
      onError={FaErrorLog}
    >
      {currentProjectSettings?.int_segment && (
        <div className='mt-4 flex flex-col  w-full'>
          <Text type='title' level={7} color='character-primary' weight='bold'>
            API Key
          </Text>
          <div className='mt-2 flex items-center gap-4'>
            <Input
              ref={textAreaRef}
              placeholder='API Key'
              value={activeProject?.private_token}
              style={{
                width: '320px',
                color: '#B7BEC8',
                borderRadius: '1px solid #B7BEC8'
              }}
            />
            <Button type='primary' onClick={copyToClipboard}>
              Copy Code
            </Button>
          </div>
        </div>
      )}
      <div className='mt-4 flex'>
        {currentProjectSettings?.int_segment ? (
          <Button loading={loading} onClick={() => onDisconnect()}>
            Disable
          </Button>
        ) : (
          <Button type='primary' loading={loading} onClick={enableSegment}>
            Get API Key
          </Button>
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
})(SegmentIntegration);
