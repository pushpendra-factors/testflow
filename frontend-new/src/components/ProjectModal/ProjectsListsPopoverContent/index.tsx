import { LeftOutlined, PlusOutlined, RightOutlined } from '@ant-design/icons';
import { SVG, Text } from 'Components/factorsComponents';
import { PathUrls } from 'Routes/pathUrls';
import { Avatar, Button, Input } from 'antd';
import React, { useMemo, useRef, useState } from 'react';
import { Link, useHistory } from 'react-router-dom';
import VirtualList from 'rc-virtual-list';
import useKeyboardNavigation from 'hooks/useKeyboardNavigation';
import useAutoFocus from 'hooks/useAutoFocus';
import { useProductFruitsApi } from 'react-product-fruits';
import { useSelector } from 'react-redux';
import { PLANS, PLANS_V0 } from 'Constants/plans.constants';
import { meetLink } from 'Utils/meetLink';
import { ProjectListsPopoverContentType } from './types';
import styles from '../index.module.scss';

const renderProjectImage = (project: any) =>
  project.profile_picture ? (
    <img
      alt='profile_picture'
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
  );

export function SearchProjectsList(props: {
  projects: Array<any>;
  active_project: any;
  handleProjectListItemClick: any;
}) {
  const {
    projects,
    active_project,
    handleProjectListItemClick = () => {}
  } = props;
  const inputComponentRef = useAutoFocus(true);
  const [searchProjectName, setsearchProjectName] = useState('');
  const searchProject = (e) => {
    setsearchProjectName(e.target.value);
  };
  const projectsList = useMemo(
    () =>
      projects
        .filter(
          (project) =>
            project?.name
              .toLowerCase()
              .includes(searchProjectName.toLowerCase())
        )
        .sort((a, b) =>
          active_project?.id === a?.id
            ? -1
            : active_project?.id === b?.id
              ? 1
              : 0
        ),
    [projects, searchProjectName, active_project]
  );
  return (
    <>
      {projects?.length > 0 ? (
        <Input
          onChange={(e) => searchProject(e)}
          value={searchProjectName}
          placeholder='Search Project'
          className='fa-project-list--search border-black w-full fa-input'
          ref={inputComponentRef}
          tabIndex={0}
        />
      ) : null}
      <div className='flex flex-col items-start fa-project-list--wrapper w-full'>
        <VirtualList
          data={projectsList}
          height={240}
          style={{ width: '100%', padding: '5px' }}
          itemHeight={44}
          itemKey='id'
          fullHeight
        >
          {(project, index) => (
            <div
              tabIndex={0}
              key={index}
              className={`flex justify-between items-center mx-2 ${
                active_project?.id === project?.id ? 'active' : ''
              } ${styles.project_item}`}
              style={{ margin: 0 }}
              onClick={() => handleProjectListItemClick(project)}
              onKeyUp={(e) =>
                e.key === 'Enter' ? handleProjectListItemClick(project) : ''
              }
            >
              <div className='flex items-center flex-nowrap'>
                {renderProjectImage(project)}

                <span className='font-bold ml-3'>{project?.name}</span>
              </div>
              {active_project?.id === project?.id ? (
                <SVG name='check_circle' color='#1890FF' />
              ) : null}
            </div>
          )}
        </VirtualList>
      </div>
    </>
  );
}
function ProjectsListsPopoverContent(props: ProjectListsPopoverContentType) {
  const {
    variant,
    currentAgent,
    showProjectsList,
    setShowPopOver,
    showUserSettingsModal,
    userLogout,
    setShowProjectsList,
    setchangeProjectModal,
    setselectedProject,
    active_project,
    projects
  } = props;
  const history = useHistory();

  const { plan } = useSelector((state: any) => state.featureConfig);
  let isFreePlan = true;
  if (plan) {
    isFreePlan =
      plan?.name === PLANS.PLAN_FREE || plan?.name === PLANS_V0?.PLAN_FREE;
  }

  const containerListRef = useRef<any>();
  const onKeydownEvent = (e: any) => useKeyboardNavigation(containerListRef, e);
  const handleClosePopover = () => setShowPopOver(false);

  const [openRaiseIssue, setOpenRaiseIssue] = useState(false);
  const handleProjectListItemClick = (project: any) => {
    if (active_project?.id !== project?.id) {
      setShowPopOver(false);
      setchangeProjectModal(true);
      setselectedProject(project);
    }
  };
  const renderProjectsList = (
    <div className={styles.projects_list_container}>
      <div style={{ overflow: 'hidden' }}>
        {showProjectsList === false ? (
          <div
            className={`${styles.active_project_div}`}
            onClick={() => setShowProjectsList(true)}
          >
            <div className='flex items-center gap-2'>
              {renderProjectImage(active_project)}

              <Text
                type='title'
                level={7}
                weight='bold'
                extraClass='m-0'
                color='grey-2'
              >
                {active_project?.name}{' '}
              </Text>
            </div>
            <div>
              <RightOutlined />
            </div>
          </div>
        ) : (
          <div ref={containerListRef} onKeyDown={onKeydownEvent}>
            {' '}
            <div className={`${styles.active_project_div_active}`}>
              <div onClick={() => setShowProjectsList(false)}>
                <LeftOutlined />
              </div>
              <div>
                <Text
                  type='title'
                  level={7}
                  weight='bold'
                  extraClass='m-0'
                  color='grey-2'
                >
                  My Projects
                </Text>
              </div>
            </div>
            <SearchProjectsList
              handleProjectListItemClick={handleProjectListItemClick}
              active_project={active_project}
              projects={projects}
            />
          </div>
        )}
      </div>
    </div>
  );
  const handleRaiseIssue = (api: {
    feedback: {
      showModal: (position: any, onClose?: () => void) => void;
    };
  }) => {
    if (openRaiseIssue) {
      api.feedback.showModal({ bottom: '0px', right: '100px' }, () => {
        setOpenRaiseIssue(false);
      });
    }
    return () => {};
  };

  useProductFruitsApi(handleRaiseIssue, [openRaiseIssue]);

  const actionsList = [
    {
      id: 'group-1',
      group: 'Shortcuts',
      items: [
        {
          id: 'item-1',
          text: 'Plans & Billing',
          props: {
            to: `${PathUrls.SettingsPricing}?activeTab=billing`,
            onClick: handleClosePopover
          }
        },
        {
          id: 'item-2',
          text: 'Invite users',
          props: {
            to: PathUrls.SettingsMembers,
            onClick: handleClosePopover
          }
        },
        {
          id: 'item-3',
          text: 'Enrichment Rules',
          props: {
            to: `${PathUrls.SettingsIntegration}/factors_deanonymisation?activeTab=enrichmentRules`,
            onClick: handleClosePopover
          }
        },
        {
          id: 'item-4',
          text: 'Setup Assist',
          props: {
            to: PathUrls.Checklist,
            onClick: handleClosePopover
          }
        }
      ]
    },
    {
      id: 'group-2',
      group: 'Help and Support',
      items: [
        {
          id: 'item-1',
          text: 'Schedule a call',
          props: {
            onClick: () => {
              window.open(meetLink(isFreePlan), '_blank');
              setShowPopOver(false);
            }
          }
        },
        {
          id: 'item-2',
          props: {
            href: 'https://help.factors.ai/en/',
            target: '_blank',
            rel: 'noreferrer',
            onClick: handleClosePopover
          },
          text: 'Product Documentation'
        },
        {
          id: 'item-3',
          text: 'Raise an issue',
          props: {
            onClick: () => setOpenRaiseIssue(true)
          }
        },

        {
          id: 'item-4',
          text: 'Privacy and Security',
          props: {
            target: '_blank',
            rel: 'noreferrer',
            href: 'https://www.factors.ai/privacy-policy',
            onClick: handleClosePopover
          }
        }
      ]
    }
  ];
  return (
    <div data-tour='step-9' className='fa-popupcard'>
      <div className={`${styles.popover_content__header}`}>Signed in as</div>
      <div
        className={`${styles.popover_content__settings} ${
          variant === 'app' ? 'cursor-pointer' : ''
        }`}
        onClick={() => {
          if (variant === 'app') {
            setShowPopOver(false);
            // showUserSettingsModal();
            history.push(PathUrls.SettingsPersonalUser);
          }
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
          >{`${currentAgent?.first_name?.charAt(
            0
          )}${currentAgent?.last_name?.charAt(0)}`}</Avatar>
          <div className='flex flex-col ml-3'>
            <Text
              type='title'
              level={7}
              weight='bold'
              extraClass='m-0'
            >{`${currentAgent?.first_name} ${currentAgent?.last_name}`}</Text>
            <div className='text-xs'>{currentAgent?.email}</div>
          </div>
        </div>
        {variant === 'app' && <SVG name='settings' size={24} />}
      </div>
      <div className='fa-popupcard-divider' style={{ margin: 0 }} />
      <div>{renderProjectsList}</div>
      {projects?.length > 0 && (
        <div className='fa-popupcard-divider' style={{ margin: 0 }} />
      )}

      {variant === 'app' && showProjectsList === false && (
        <>
          {actionsList.map((eachGroup, eachGroupIndex) => (
            <React.Fragment key={eachGroup.id}>
              <div className='px-4 py-2 text-xs'>{eachGroup.group}</div>
              {eachGroup.items.map((eachItem) => (
                <div
                  key={eachGroup.id + eachItem.id}
                  className={` ${styles.popover_content__additionalActions}`}
                >
                  {eachItem?.props?.href ? (
                    <a {...eachItem?.props}>{eachItem.text}</a>
                  ) : (
                    <Link {...eachItem?.props}>{eachItem.text}</Link>
                  )}
                </div>
              ))}

              <div
                className='fa-popupcard-divider'
                style={{ margin: '4px 0' }}
              />
            </React.Fragment>
          ))}
        </>
      )}

      <div>
        {showProjectsList === true ? (
          <Button
            size='large'
            type='text'
            icon={<PlusOutlined />}
            className={styles.popover_content__signout}
            onClick={() => {
              setShowPopOver(false);
              history.push(`${PathUrls.Onboarding}?setup=new`);
            }}
          >
            New Project
          </Button>
        ) : (
          <Button
            icon={<SVG name='signout' extraClass='mr-1' color='#EA6262' />}
            size='large'
            type='text'
            onClick={() => {
              setShowPopOver(false);
              userLogout();
            }}
            className={styles.popover_content__signout}
          >
            <span style={{ color: '#EA6262' }}>Logout</span>
          </Button>
        )}
      </div>
    </div>
  );
}

export default ProjectsListsPopoverContent;
