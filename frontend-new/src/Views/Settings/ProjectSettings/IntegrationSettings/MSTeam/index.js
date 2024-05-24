import React, { useState } from 'react';
import { connect } from 'react-redux';
import { Button, message, Input, Avatar, Popover, Modal } from 'antd';
import { Text, FaErrorComp, FaErrorLog, SVG } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import {
  disableTeamsIntegration,
  enableTeamsIntegration,
  fetchProjectSettingsV1
} from '../../../../../reducers/global';
import { sendSlackNotification } from '../../../../../utils/slack';

const MSTeamIntegration = ({
  activeProject,
  agent_details,
  fetchProjectSettingsV1,
  enableTeamsIntegration,
  disableTeamsIntegration,
  projectSettings,
  integrationCallback
}) => {
  const [loading, setLoading] = useState(false);

  const onDisconnect = () => {
    Modal.confirm({
      title: 'Are you sure you want to disable this?',
      content:
        'You are about to disable this integration. Factors will stop bringing in data from this source.',
      okText: 'Disconnect',
      cancelText: 'Cancel',
      onOk: () => {
        setLoading(true);
        disableTeamsIntegration(activeProject.id)
          .then(() => {
            setLoading(false);
            setTimeout(() => {
              message.success('Microsoft Teams integration disconnected!');
            }, 500);
            fetchProjectSettingsV1(activeProject.id);
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

  const enableTeams = () => {
    setLoading(true);
    enableTeamsIntegration(activeProject.id)
      .then((r) => {
        setLoading(false);
        if (r.status == 200) {
          window.location = r.data.redirectURL;
          sendSlackNotification(
            agent_details.email,
            activeProject.name,
            'Microsoft Teams'
          );
        }
        if (r.status >= 400) {
          message.error('Error fetching Microsoft Teams redirect url');
        }
      })
      .catch((err) => {
        setLoading(false);
        console.log('Microsoft Teams error-->', err);
      });
  };

  return (
    <ErrorBoundary
      fallback={
        <FaErrorComp subtitle='Facing issues with Microsoft Teams integrations' />
      }
      onError={FaErrorLog}
    >
      <div className='mt-4'>
        {!projectSettings?.int_teams ? (
          <Button
            className='mr-2'
            type='primary'
            loading={loading}
            onClick={enableTeams}
          >
            Connect Now
          </Button>
        ) : (
          <div className='flex items-center justify-between'>
            <Button
              className='mr-2'
              loading={loading}
              onClick={() => onDisconnect()}
            >
              Disconnect
            </Button>
            <div>
              <Popover
                content={
                  <Text type='title' size={10} extraClass='max-w-xs'>
                    The feature is only accessible to
                    <span className='font-bold text-slate-500'>
                      {` ${agent_details.first_name} ${agent_details.last_name}`}
                    </span>
                  </Text>
                }
                title={null}
                trigger='hover'
              >
                <div className='flex gap-2 items-center'>
                  <Text
                    type='title'
                    level={7}
                    color='character-primary'
                    extraClass='m-0 '
                  >
                    Integrated by
                  </Text>
                  <Avatar
                    src='../../../../../assets/avatar/avatar.png'
                    className='ml-2'
                    size={24}
                  />
                  <Text
                    type='title'
                    level={7}
                    color='character-primary'
                    extraClass='m-0 '
                  >
                    {`${agent_details.first_name} ${agent_details.last_name}`}
                  </Text>
                </div>
              </Popover>
            </div>
          </div>
        )}
      </div>
    </ErrorBoundary>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  agent_details: state.agent.agent_details,
  projectSettings: state.global.projectSettingsV1
});

export default connect(mapStateToProps, {
  fetchProjectSettingsV1,
  enableTeamsIntegration,
  disableTeamsIntegration
})(MSTeamIntegration);
