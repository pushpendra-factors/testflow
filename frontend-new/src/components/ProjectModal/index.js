import React, { useState, useEffect, useRef, useMemo } from 'react';
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
import {
  updateAgentInfo,
  fetchAgentInfo,
  fetchProjectAgents,
  signout
} from 'Reducers/agentActions';
import { USER_LOGOUT } from 'Reducers/types';
import { getActiveProjectDetails, fetchProjectSettings } from 'Reducers/global';
import { connect, useDispatch, useSelector } from 'react-redux';
import { useHistory } from 'react-router-dom';
import factorsai from 'factorsai';
import useAutoFocus from 'hooks/useAutoFocus';
import { PathUrls } from 'Routes/pathUrls';
import logger from 'Utils/logger';
import { meetLink } from 'Utils/meetLink';
import { PLANS, PLANS_V0 } from 'Constants/plans.constants';
import VirtualList from 'rc-virtual-list';
import { RESET_GROUPBY } from 'Reducers/coreQuery/actions';
import { TOOLTIP_CONSTANTS } from '../../constants/tooltips.constans';
import UserSettings from '../../Views/Settings/UserSettings';
import styles from './index.module.scss';
import { Text, SVG } from '../factorsComponents';
import { LeftOutlined, PlusOutlined, RightOutlined } from '@ant-design/icons';
import useKeyboardNavigation from 'hooks/useKeyboardNavigation';
import ProjectsListsPopoverContent from './ProjectsListsPopoverContent';

function ProjectModal(props) {
  const [ShowPopOver, setShowPopOver] = useState(false);
  const [searchProjectName, setsearchProjectName] = useState('');
  const [ShowUserSettings, setShowUserSettings] = useState(false);
  const [changeProjectModal, setchangeProjectModal] = useState(false);
  const [selectedProject, setselectedProject] = useState(null);
  const history = useHistory();
  const variant = props?.variant === 'onboarding' ? 'onboarding' : 'app';
  const [showProjectsList, setShowProjectsList] = useState(false);
  const { plan } = useSelector((state) => state.featureConfig);

  let isFreePlan = true;
  if (plan) {
    isFreePlan =
      plan?.name === PLANS.PLAN_FREE || plan?.name === PLANS_V0?.PLAN_FREE;
  }
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
    localStorage.setItem('prevActiveProject', props?.active_project?.id || '');
    props.getActiveProjectDetails(selectedProject?.id);
    props.fetchProjectSettings(selectedProject?.id);
    history.push('/');
    notification.success({
      message: 'Project Changed!',
      description: `You are currently viewing data from ${selectedProject?.name}`
    });
    dispatch({ type: RESET_GROUPBY });
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
      // Factors identify users
      const userAndProjectDetails = {
        ...props?.currentAgent,
        project_name: props?.active_project?.name,
        project_id: props?.active_project?.id
      };
      factorsai.identify(props?.currentAgent?.email, userAndProjectDetails);
    }
  }, [props?.currentAgent, props?.active_project]);
  useEffect(() => {
    if (showProjectsList || ShowPopOver === false) setsearchProjectName('');
  }, [showProjectsList, ShowPopOver]);
  const userLogout = async () => {
    try {
      await props.signout();
      dispatch({ type: USER_LOGOUT });
    } catch (error) {
      logger.error('Error in logging out', error);
    }
  };
  const handleClosePopover = () => setShowPopOver(false);
  const popoverContent = (
    <ProjectsListsPopoverContent
      variant={variant}
      currentAgent={props.currentAgent}
      active_project={props.active_project}
      projects={props.projects}
      showProjectsList={showProjectsList}
      setShowPopOver={setShowPopOver}
      showUserSettingsModal={showUserSettingsModal}
      userLogout={userLogout}
      setShowProjectsList={setShowProjectsList}
      searchProject={searchProject}
      searchProjectName={searchProjectName}
      setchangeProjectModal={setchangeProjectModal}
      setselectedProject={setselectedProject}
    />
  );
  return (
    <>
      <Popover
        placement='bottomRight'
        overlayClassName='fa-popupcard--wrapper fa-at-popover--projects'
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
          mouseEnterDelay={1}
          title={
            variant === 'app'
              ? 'Access your projects, account settings, and more'
              : ''
          }
          color={TOOLTIP_CONSTANTS.DARK}
        >
          <Button
            className={`${styles.button} flex items-center mr-4`}
            type='text'
            size='large'
            id='fa-at-dropdown--projects'
          >
            <Avatar
              size={36}
              shape='square'
              style={{
                background:
                  props?.active_project?.profile_picture?.length > 0
                    ? '#FFFFFF'
                    : '#FF7875',
                textTransform: 'uppercase',
                fontWeight: '600',
                borderRadius: '8px'
              }}
              src={props?.active_project?.profile_picture}
            >
              {' '}
              {props?.active_project?.name?.charAt(0)
                ? `${props.active_project?.name?.charAt(0)}`
                : props?.currentAgent?.first_name?.charAt(0)}
            </Avatar>

            <div className='flex flex-col items-start ml-2'>
              <div className='flex items-center'>
                <Text
                  type='title'
                  level={7}
                  extraClass='m-0 capitalize'
                  weight='bold'
                  color={variant === 'app' ? 'white' : undefined}
                >
                  {props?.active_project?.name
                    ? props.active_project.name
                    : props?.currentAgent?.first_name}
                </Text>
                <SVG name='caretDown' size={20} color='#BFBFBF' />
              </div>
              <div
                className={`text-xs ${
                  variant === 'app' ? 'text-white' : ''
                }  opacity-80`}
              >
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

      <Modal
        visible={changeProjectModal}
        zIndex={1020}
        onCancel={() => {
          setchangeProjectModal(false);
          setselectedProject(null);
        }}
        className='fa-modal--regular fa-modal--slideInDown'
        okText='Switch'
        onOk={() => {
          setShowPopOver(false);
          setchangeProjectModal(false);
          setselectedProject(null);
          switchProject();
        }}
        centered
        transitionName=''
        maskTransitionName=''
      >
        <div className='p-4'>
          <Row>
            <Col span={24}>
              <Text type='title' level={4} weight='bold' extraClass='m-0'>
                Do you want to switch the project?
              </Text>
              <Text type='title' level={7} color='grey' extraClass='m-0 mt-2'>
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

const mapStateToProps = (state) => ({
  projects: state.global.projects,
  active_project: state.global.active_project,
  currentAgent: state.agent.agent_details,
  agents: state.agent.agents
});
export default connect(mapStateToProps, {
  fetchProjectAgents,
  getActiveProjectDetails,
  signout,
  updateAgentInfo,
  fetchAgentInfo,
  fetchProjectSettings
})(ProjectModal);
