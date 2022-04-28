import React, { useState, useEffect } from 'react';
import { Button, Avatar, Popover, Modal, Row, Col, notification } from 'antd';
import { Text, SVG } from '../factorsComponents';
import { PlusOutlined, PoweroffOutlined } from '@ant-design/icons';
import styles from './index.module.scss';
import {
  updateAgentInfo,
  fetchAgentInfo,
  fetchProjectAgents,
  signout,
} from 'Reducers/agentActions';
import { setActiveProject } from 'Reducers/global';
import UserSettings from '../../Views/Settings/UserSettings';
import NewProject from '../../Views/Settings/SetupAssist/Modals/NewProject';
import { connect } from 'react-redux';
import { useHistory } from 'react-router-dom';
import factorsai from 'factorsai';

function ProjectModal(props) {
  const [ShowPopOver, setShowPopOver] = useState(false);
  const [searchProjectName, setsearchProjectName] = useState('');
  const [showProjectModal, setShowProjectModal] = useState(false);
  const [ShowUserSettings, setShowUserSettings] = useState(false);
  const [changeProjectModal, setchangeProjectModal] = useState(false);
  const [selectedProject, setselectedProject] = useState(null);
  const history = useHistory();

  const searchProject = (e) => {
    setsearchProjectName(e.target.value);
  };
  const showUserSettingsModal = () => {
    setShowUserSettings(true);
  };
  const closeUserSettingsModal = () => {
    setShowUserSettings(false);
  };

  const switchProject = () => {
    localStorage.setItem('activeProject', selectedProject?.id);
    props.setActiveProject(selectedProject);
    history.push('/');
    notification.success({
      message: 'Project Changed!',
      description: `You are currently viewing data from ${selectedProject.name}`,
    });
  };

  const UpdateOnboardingSeen = () => {
    props.updateAgentInfo({ is_onboarding_flow_seen: true }).then(() => {
      props.fetchAgentInfo();
    });
    //Factors FIRST_TIME_LOGIN tracking for NON_INVITED
    factorsai.track('FIRST_TIME_LOGIN', { email: props?.currentAgent?.email });
  };

  useEffect(() => {
    if (!props.agents) {
      props.fetchProjectAgents(props.active_project?.id);
    }
    if (!props.currentAgent) {
      props.fetchAgentInfo();
    }
    if (props.agents && props.currentAgent) {
      let agent = props.agents?.filter(
        (agent) => agent.email === props.currentAgent.email
      );
      if (agent[0]?.invited_by) {
        setShowProjectModal(false);
      } else if (!props.currentAgent?.is_onboarding_flow_seen) {
        setShowProjectModal(true);
        UpdateOnboardingSeen();
      }
    } else if (
      props.currentAgent &&
      !props.currentAgent?.is_onboarding_flow_seen
    ) {
      setShowProjectModal(true);
      UpdateOnboardingSeen();
    }
  }, [props.active_project, props.agents, props.currentAgent]);

  useEffect(() => {
    if (props?.currentAgent) {
      //Factors identify users
      let userAndProjectDetails = {
        ...props?.currentAgent,
        project_name: props?.active_project?.name,
        project_id: props?.active_project?.id,
      };
      factorsai.identify(props?.currentAgent?.email, userAndProjectDetails);
    }
  }, [props?.currentAgent, props?.active_project]);

  const popoverContent = () => {
    return (
      <div data-tour='step-9' className={'fa-popupcard'}>
        <Text type={'title'} level={7} weight={'bold'} extraClass={'m-0'}>
          Projects
        </Text>
        {props.projects.length > 6 ? (
          <input
            onChange={(e) => searchProject(e)}
            value={searchProjectName}
            placeholder={'Search Project'}
            className={'fa-project-list--search'}
          />
        ) : null}
        <div className={'flex flex-col items-start fa-project-list--wrapper'}>
          {props.projects
            .filter((project) =>
              project.name
                .toLowerCase()
                .includes(searchProjectName.toLowerCase())
            )
            .map((project, index) => {
              return (
                <div
                  key={index}
                  className={`flex justify-start items-center project-item ${
                    props.active_project.id === project.id ? 'active' : null
                  }`}
                  onClick={() => {
                    if (props.active_project.id !== project.id) {
                      setShowPopOver(false);
                      setchangeProjectModal(true);
                      setselectedProject(project);
                    }
                  }}
                >
                  <Avatar
                    size={28}
                    style={{
                      color: '#fff',
                      backgroundColor: '#52BE95',
                      fontSize: '14px',
                      textTransform: 'uppercase',
                      fontWeight: '400',
                    }}
                  >{`${project?.name?.charAt(0)}`}</Avatar>
                  <Text
                    type={'title'}
                    level={7}
                    weight={'bold'}
                    extraClass={'m-0 ml-2'}
                  >
                    {project.name}
                  </Text>
                </div>
              );
            })}
        </div>
        <div className={'fa-popupcard-divider'} />
        <Button
          size={'large'}
          type={'text'}
          onClick={() => {
            setShowPopOver(false);
            setShowProjectModal(true);
          }}
        >
          <span className={'mr-4'}>
            <PlusOutlined />
          </span>
          New Project
        </Button>
        <div className={'fa-popupcard-divider'} />
        <div
          className={'flex justify-start items-center project-item'}
          onClick={() => {
            setShowPopOver(false);
            showUserSettingsModal();
          }}
        >
          <Avatar
            size={28}
            style={{
              color: '#f56a00',
              backgroundColor: '#fde3cf',
              fontSize: '12px',
            }}
          >{`${props.currentAgent?.first_name?.charAt(
            0
          )}${props.currentAgent?.last_name?.charAt(0)}`}</Avatar>
          <Text type={'title'} level={7} extraClass={'m-0 ml-2'}>
            Account Settings
          </Text>
        </div>
        <Button
          size={'large'}
          type={'text'}
          onClick={() => {
            setShowPopOver(false);
            props.signout();
          }}
        >
          <span className={'mr-4'}>
            <PoweroffOutlined />
          </span>
          Logout
        </Button>
      </div>
    );
  };
  return (
    <>
      <Popover
        placement='bottomRight'
        overlayClassName={'fa-popupcard--wrapper'}
        title={false}
        content={popoverContent}
        visible={ShowPopOver}
        onVisibleChange={(visible) => {
          setShowPopOver(visible);
        }}
        onClick={() => {
          setsearchProjectName('');
          setShowPopOver(true);
        }}
        trigger='click'
      >
        <Button
          className={`${styles.button} flex items-center mr-4`}
          type='text'
          size='large'
        >
          <Avatar
            size={36}
            shape='square'
            style={{
              background: '#ff0000',
              opacity: '0.6',
              textTransform: 'uppercase',
              fontWeight: '400',
              borderRadius: '4px',
            }}
          >{`${props.active_project?.name?.charAt(0)}`}</Avatar>

          <div className='flex flex-col items-start ml-2'>
            <div className='flex items-center'>
              <Text type={'title'} level={7} extraClass={'m-0'} weight={'bold'}>
                {`${props.active_project?.name}`}
              </Text>
              <SVG name='caretDown' size={20} />
            </div>
            <div className={`text-xs`}>{props.currentAgent?.email}</div>
          </div>
        </Button>
      </Popover>

      <UserSettings
        visible={ShowUserSettings}
        handleCancel={closeUserSettingsModal}
      />
      <NewProject
        visible={showProjectModal}
        handleCancel={() => setShowProjectModal(false)}
      />
      <Modal
        visible={changeProjectModal}
        zIndex={1020}
        onCancel={() => {
          setchangeProjectModal(false);
          setselectedProject(null);
        }}
        className={'fa-modal--regular fa-modal--slideInDown'}
        okText={'Switch'}
        onOk={() => {
          setShowPopOver(false);
          setchangeProjectModal(false);
          setselectedProject(null);
          switchProject();
        }}
        centered={true}
        transitionName=''
        maskTransitionName=''
      >
        <div className={'p-4'}>
          <Row>
            <Col span={24}>
              <Text type={'title'} level={4} weight={'bold'} extraClass={'m-0'}>
                Do you want to switch the project?
              </Text>
              <Text
                type={'title'}
                level={7}
                color={'grey'}
                extraClass={'m-0 mt-2'}
              >
                You can easily switch between projects. You will be redirected a
                different dataset.
              </Text>
            </Col>
          </Row>
        </div>
      </Modal>
    </>
  );
}

const mapStateToProps = (state) => {
  return {
    projects: state.global.projects,
    active_project: state.global.active_project,
    currentAgent: state.agent.agent_details,
    agents: state.agent.agents,
  };
};
export default connect(mapStateToProps, {
  fetchProjectAgents,
  setActiveProject,
  signout,
  updateAgentInfo,
  fetchAgentInfo,
})(ProjectModal);
