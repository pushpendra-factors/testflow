import React, { useState, useEffect, useRef, useMemo } from 'react';
import {
  Button,
  Avatar,
  Popover,
  Modal,
  Row,
  Col,
  notification,
  Tooltip,
  message
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
import { haveRestrictionForSelectedProject } from 'Utils/global';

function ProjectModal(props) {
  const { getActiveProjectDetails, fetchProjectSettings } = props;
  const [ShowPopOver, setShowPopOver] = useState(false);
  const [ShowUserSettings, setShowUserSettings] = useState(false);
  const [changeProjectModal, setchangeProjectModal] = useState(false);
  const [selectedProject, setselectedProject] = useState(null);
  const history = useHistory();

  const variant = props?.variant === 'onboarding' ? 'onboarding' : 'app';
  const [showProjectsList, setShowProjectsList] = useState(false);
  const { plan } = useSelector((state) => state.featureConfig);
  const { projects } = useSelector((state) => state.global);
  const { loginMethod } = useSelector((state) => state.agent);
  const { currentProjectSettings } = useSelector((state) => state.global);

  let isFreePlan = true;
  if (plan) {
    isFreePlan =
      plan?.name === PLANS.PLAN_FREE || plan?.name === PLANS_V0?.PLAN_FREE;
  }
  const dispatch = useDispatch();

  const showUserSettingsModal = () => {
    setShowUserSettings(true);
  };
  const closeUserSettingsModal = () => {
    setShowUserSettings(false);
  };

  const switchProject = async () => {
    const can = haveRestrictionForSelectedProject(
      loginMethod,
      selectedProject?.login_method
    );

    if (can) {
      localStorage.setItem('activeProject', selectedProject?.id);
      await getActiveProjectDetails(selectedProject?.id);
      await fetchProjectSettings(selectedProject?.id);
      history.push('/');
      dispatch({ type: RESET_GROUPBY });
    } else {
      history.push(PathUrls.ProjectChangeAuthentication, {
        selectedProject: selectedProject,
        currentActiveProject: props?.active_project,
        currentAgent: props.currentAgent,
        projects: projects,
        login_method: currentProjectSettings?.sso_state
      });
    }
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

  const userLogout = async () => {
    try {
      await props.signout();
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
