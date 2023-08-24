import React, { useState, useEffect } from 'react';
import {
  Button,
  Avatar,
  Popover,
  Modal,
  Row,
  Col,
  notification,
  Tooltip
} from 'antd';
import { Text, SVG } from '../factorsComponents';
import styles from './index.module.scss';
import {
  updateAgentInfo,
  fetchAgentInfo,
  fetchProjectAgents,
  signout
} from 'Reducers/agentActions';
import { USER_LOGOUT } from 'Reducers/types';
import { setActiveProject } from 'Reducers/global';
import UserSettings from '../../Views/Settings/UserSettings';
// import NewProject from '../../Views/Settings/SetupAssist/Modals/NewProject';
import { connect, useDispatch } from 'react-redux';
import { useHistory } from 'react-router-dom';
import factorsai from 'factorsai';
import { fetchProjectSettings } from 'Reducers/global';
import { TOOLTIP_CONSTANTS } from '../../constants/tooltips.constans';
import useAutoFocus from 'hooks/useAutoFocus';
import { PathUrls } from 'Routes/pathUrls';

function ProjectModal(props) {
  const [ShowPopOver, setShowPopOver] = useState(false);
  const [searchProjectName, setsearchProjectName] = useState('');
  // const [showProjectModal, setShowProjectModal] = useState(false);
  const [ShowUserSettings, setShowUserSettings] = useState(false);
  const [changeProjectModal, setchangeProjectModal] = useState(false);
  const [selectedProject, setselectedProject] = useState(null);
  const history = useHistory();
  const inputComponentRef = useAutoFocus(ShowPopOver);

  const dispatch = useDispatch();

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
    props.fetchProjectSettings(selectedProject?.id);
    history.push('/');
    notification.success({
      message: 'Project Changed!',
      description: `You are currently viewing data from ${selectedProject?.name}`
    });
  };

  // const UpdateOnboardingSeen = () => {
  //   props.updateAgentInfo({ is_onboarding_flow_seen: true }).then(() => {
  //     props.fetchAgentInfo();
  //   });
  //   //Factors FIRST_TIME_LOGIN tracking for NON_INVITED
  //   factorsai.track('FIRST_TIME_LOGIN', { email: props?.currentAgent?.email });
  //   if (props?.currentAgent?.is_auth0_user) {
  //     factorsai.track('$form_submitted', {
  //       $email: props?.currentAgent?.email
  //     });
  //   }
  // };

  useEffect(() => {
    if (props?.currentAgent) {
      //Factors identify users
      let userAndProjectDetails = {
        ...props?.currentAgent,
        project_name: props?.active_project?.name,
        project_id: props?.active_project?.id
      };
      factorsai.identify(props?.currentAgent?.email, userAndProjectDetails);
    }
  }, [props?.currentAgent, props?.active_project]);

  const userLogout = () => {
    props.signout();
    dispatch({ type: USER_LOGOUT });
  };

  const popoverContent = () => (
    <div data-tour='step-9' className={'fa-popupcard'}>
      <div className={`${styles.popover_content__header}`}>Signed in as</div>
      <div
        className={`${styles.popover_content__settings}`}
        onClick={() => {
          setShowPopOver(false);
          showUserSettingsModal();
        }}
      >
        <div className='flex items-center'>
          <Avatar
            size={40}
            style={{
              color: '#f56a00',
              backgroundColor: '#fde3cf',
              fontSize: '12px'
            }}
          >{`${props.currentAgent?.first_name?.charAt(
            0
          )}${props.currentAgent?.last_name?.charAt(0)}`}</Avatar>
          <div className='flex flex-col ml-3'>
            <Text
              type={'title'}
              level={7}
              weight={'bold'}
              extraClass={'m-0'}
            >{`${props.currentAgent?.first_name} ${props.currentAgent?.last_name}`}</Text>
            <div className={`text-xs`}>{props.currentAgent?.email}</div>
          </div>
        </div>
        <SVG name='settings' size={24} />
      </div>
      <div className={'fa-popupcard-divider'} />
      <div className={`${styles.popover_content__projectList}`}>
        <Text
          type={'title'}
          level={7}
          weight={'bold'}
          extraClass={'m-0'}
          color='grey-2'
        >
          Your Projects
        </Text>
        <Button
          type={'text'}
          className='fa-btn--custom'
          onClick={() => {
            setShowPopOver(false);
            // setShowProjectModal(true);
            history.push(`${PathUrls.Onboarding}?setup=new`);
          }}
        >
          <SVG name='plus' />
        </Button>
      </div>

      {props.projects?.length > 6 ? (
        <input
          onChange={(e) => searchProject(e)}
          value={searchProjectName}
          placeholder={'Search Project'}
          className={'fa-project-list--search'}
          ref={inputComponentRef}
        />
      ) : null}
      <div className={'flex flex-col items-start fa-project-list--wrapper'}>
        {props.projects
          .filter((project) =>
            project?.name
              .toLowerCase()
              .includes(searchProjectName.toLowerCase())
          )
          .map((project, index) => (
            <div
              key={index}
              className={`flex justify-between items-center project-item mx-2 ${
                props.active_project?.id === project?.id ? 'active' : null
              }`}
              onClick={() => {
                if (props.active_project?.id !== project?.id) {
                  setShowPopOver(false);
                  setchangeProjectModal(true);
                  setselectedProject(project);
                }
              }}
            >
              <div className='flex items-center flex-no-wrap'>
                {project.profile_picture ? (
                  <img
                    src={project.profile_picture}
                    style={{
                      borderRadius: '4px',
                      width: '28px',
                      height: '28px'
                    }}
                  />
                ) : (
                  <Avatar
                    size={28}
                    shape='square'
                    style={{
                      background: '#83D2D2',
                      fontSize: '14px',
                      textTransform: 'uppercase',
                      fontWeight: '400',
                      borderRadius: '4px'
                    }}
                  >{`${project?.name?.charAt(0)}`}</Avatar>
                )}

                <span className='font-bold ml-3'>{project?.name}</span>
              </div>
              {props.active_project?.id === project?.id ? (
                <SVG name='check_circle' />
              ) : null}
            </div>
          ))}
      </div>
      <div className={'fa-popupcard-divider'} />
      <div className={styles.popover_content__additionalActions}>
        <a href='https://help.factors.ai' target='_blank'>
          Help
        </a>
      </div>

      <div style={{ borderTop: 'thin solid #e7e9ed' }}>
        <Button
          size={'large'}
          type={'text'}
          onClick={() => {
            setShowPopOver(false);
            userLogout();
          }}
          className={styles.popover_content__signout}
        >
          <div className='flex items-center'>
            <SVG name='signout' extraClass='mr-1' color='#EA6262' />
            <span style={{ color: '#EA6262' }}>Logout</span>
          </div>
        </Button>
      </div>
    </div>
  );
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
        <Tooltip
          title='Access your projects, account settings, and more'
          color={TOOLTIP_CONSTANTS.DARK}
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
                background: '#FF7875',
                textTransform: 'uppercase',
                fontWeight: '600',
                borderRadius: '8px'
              }}
            >{`${props.active_project?.name?.charAt(0)}`}</Avatar>

            <div className='flex flex-col items-start ml-2'>
              <div className='flex items-center'>
                <Text
                  type={'title'}
                  level={7}
                  extraClass={'m-0'}
                  weight={'bold'}
                  color='white'
                >
                  {`${props.active_project?.name}`}
                </Text>
                <SVG name='caretDown' size={20} color='#BFBFBF' />
              </div>
              <div className={`text-xs text-white opacity-80`}>
                {props.currentAgent?.email}
              </div>
            </div>
          </Button>
        </Tooltip>
      </Popover>

      <UserSettings
        visible={ShowUserSettings}
        handleCancel={closeUserSettingsModal}
      />
      {/* <NewProject
        visible={showProjectModal}
        handleCancel={() => setShowProjectModal(false)}
      /> */}
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
    agents: state.agent.agents
  };
};
export default connect(mapStateToProps, {
  fetchProjectAgents,
  setActiveProject,
  signout,
  updateAgentInfo,
  fetchAgentInfo,
  fetchProjectSettings
})(ProjectModal);
