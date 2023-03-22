import React, { useEffect, useState } from 'react';
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
  setIsStatus,
  kbLink = false,
  fetchProjectSettingsV1,
  enableTeamsIntegration,
  disableTeamsIntegration,
  projectSettings
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
            setIsStatus('');
            fetchProjectSettingsV1(activeProject.id);
          })
          .catch((err) => {
            message.error(`${err?.data?.error}`);
            setLoading(false);
          });
      },
      onCancel: () => {}
    });
  };

  const isSlackEnabled = () => {
    fetchProjectSettingsV1(activeProject.id);
  };

  useEffect(() => {
    isSlackEnabled();
    if (projectSettings?.int_teams) {
      setIsStatus('Active');
    } else {
      setIsStatus('');
    }
  }, [activeProject, projectSettings?.int_teams]);

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
        setIsStatus('');
      });
  };

  return (
    <>
      <ErrorBoundary
        fallback={
          <FaErrorComp subtitle={'Facing issues with Microsoft Teams integrations'} />
        }
        onError={FaErrorLog}
      >
        <div className={'mt-4 flex w-full'}>
          {projectSettings?.int_teams && (
            <>
              <div
                className={
                  'mt-4 flex flex-col border-top--thin py-4 mt-2 w-full'
                }
              >
                <Text
                  type={'title'}
                  level={6}
                  weight={'bold'}
                  extraClass={'m-0'}
                >
                  Integration Details
                </Text>
                <Text
                  type={'title'}
                  level={7}
                  color={'grey'}
                  extraClass={'m-0 mt-2'}
                >
                  Integrated by{' '}
                  <Avatar
                    src='../../../../../assets/avatar/avatar.png'
                    className={'mr-2'}
                    size={24}
                  />{' '}
                  <span className={'font-bold text-gray-700'}>
                    {agent_details.first_name + ' ' + agent_details.last_name}
                  </span>
                  <Popover
                    content={
                      <Text type={'title'} size={10} extraClass={'max-w-xs'}>
                        The feature is only accessable to
                        <span className={'font-bold text-slate-500'}>
                          {' ' +
                            agent_details.first_name +
                            ' ' +
                            agent_details.last_name}
                          .
                        </span>
                      </Text>
                    }
                    title={null}
                    trigger='hover'
                  >
                    <Button
                      type={'text'}
                      className={'m-0'}
                      style={{ backgroundColor: 'white' }}
                    >
                      <SVG name={'infoCircle'} size={18} color='gray' />
                    </Button>
                  </Popover>
                </Text>
              </div>
            </>
          )}
        </div>

        <div className={'mt-4 flex'}>
          {!projectSettings?.int_teams ? (
            <Button
              className={'mr-2'}
              type={'primary'}
              loading={loading}
              onClick={enableTeams}
            >
              Connect Now
            </Button>
          ) : (
            <Button
              className={'mr-2'}
              loading={loading}
              onClick={() => onDisconnect()}
            >
              Disconnect
            </Button>
          )}
          {kbLink && (
            <a className={'ant-btn'} target={'_blank'} href={kbLink}>
              View documentation
            </a>
          )}
        </div>
      </ErrorBoundary>
    </>
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
  disableTeamsIntegration,
})(MSTeamIntegration);
