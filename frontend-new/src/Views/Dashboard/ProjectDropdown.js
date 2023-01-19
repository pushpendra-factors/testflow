import React, { useState, useEffect, useCallback } from 'react';
import {
  Avatar,
  Button,
  Checkbox,
  Col,
  Divider,
  Dropdown,
  Input,
  List,
  Menu,
  message,
  Modal,
  Row,
  Space,
  Spin,
  Tooltip,
  Typography
} from 'antd';
import { useSelector, useDispatch, connect } from 'react-redux';
import { ErrorBoundary } from 'react-error-boundary';
import FaSelect from 'Components/FaSelect';
import factorsai from 'factorsai';
import { fetchDemoProject, getHubspotContact } from 'Reducers/global';
import userflow from 'userflow.js';
import {
  fetchActiveDashboardUnits,
  DeleteUnitFromDashboard,
  deleteDashboard,
  assignUnitsToDashboard
} from '../../reducers/dashboard/services';
import {
  ACTIVE_DASHBOARD_CHANGE,
  WIDGET_DELETED,
  DASHBOARD_DELETED,
  NEW_DASHBOARD_TEMPLATES_MODAL_OPEN,
  ADD_DASHBOARD_MODAL_OPEN
} from '../../reducers/types';
import SortableCards from './SortableCards';
import DashboardSubMenu from './DashboardSubMenu';
import ExpandableView from './ExpandableView';
import ConfirmationModal from '../../components/ConfirmationModal';
import styles from './index.module.scss';
import NoDataChart from '../../components/NoDataChart';
import {
  SVG,
  FaErrorComp,
  FaErrorLog,
  Text
} from '../../components/factorsComponents';
import GroupSelect2 from '../../components/QueryComposer/GroupSelect2';
import NewProject from '../Settings/SetupAssist/Modals/NewProject';
import { setItemToLocalStorage } from '../../utils/localStorage.helpers';
import { DASHBOARD_KEYS } from '../../constants/localStorage.constants';
import {
  CaretDownOutlined,
  DownCircleFilled,
  DownOutlined,
  LockFilled,
  LockOutlined,
  PlusOutlined,
  UnlockFilled,
  UnlockOutlined,
  UserOutlined
} from '@ant-design/icons';

import { useHistory } from 'react-router-dom';
import { INITIALIZE_GROUPBY } from 'Reducers/coreQuery/actions';
import {
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_KPI,
  QUERY_TYPE_PROFILE
} from 'Utils/constants';
import ExistingReportsModal from './ExistingReportsModal';

function ProjectDropdown({
  setaddDashboardModal,
  handleEditClick,
  durationObj,
  handleDurationChange,
  isPinned = false,
  fetchDemoProject,
  oldestRefreshTime,
  setOldestRefreshTime,
  handleRefreshClick,
  dashboardRefreshState,
  onDataLoadSuccess,
  resetDashboardRefreshState,
  handleWidgetRefresh
}) {
  const [moreOptions, setMoreOptions] = useState(false);
  const [widgetModal, setwidgetModal] = useState(false);
  const [deleteWidgetModal, showDeleteWidgetModal] = useState(false);
  const [deleteApiCalled, setDeleteApiCalled] = useState(false);
  const [widgetModalLoading, setWidgetModalLoading] = useState(false);
  const { active_project } = useSelector((state) => state.global);
  const { dashboards, activeDashboard, activeDashboardUnits } = useSelector(
    (state) => state.dashboard
  );
  const { projects } = useSelector((state) => state.global);
  let reports = useSelector((state) => state?.queries?.data);
  const [selectVisible, setSelectVisible] = useState(false);
  const [showDashboardName, setDashboardName] = useState('');
  const [showDashboardDesc, setDashboardDesc] = useState('');
  const [deleteDashboardModal, showDeleteDashboardModal] = useState(false);
  const [dashboardDeleteApi, setDashboardDeleteApi] = useState(false);
  const [demoProjectId, setDemoProjectId] = useState(null);
  const [showProjectModal, setShowProjectModal] = useState(false);

  const [isReportsModalOpen, setIsReportsModalOpen] = useState(false);
  const dispatch = useDispatch();
  const history = useHistory();

  const queries = useSelector((state) => state.queries);

  useEffect(() => {
    fetchDemoProject()
      .then((res) => {
        setDemoProjectId(res.data[0]);
      })
      .catch((err) => {
        console.log(err.data.error);
      });
  }, [active_project]);

  const changeActiveDashboard = useCallback(
    (val) => {
      if (val === activeDashboard?.id) {
        return false;
      }
      resetDashboardRefreshState();
      setOldestRefreshTime(null);
      const selectedDashboard = dashboards.data.find((d) => d.id === val);
      dispatch({
        type: ACTIVE_DASHBOARD_CHANGE,
        payload: selectedDashboard
      });
      setItemToLocalStorage(
        DASHBOARD_KEYS.ACTIVE_DASHBOARD_ID,
        selectedDashboard.id
      );
    },
    [dashboards, dispatch, activeDashboard?.id]
  );

  useEffect(() => {
    setDashboardName(activeDashboard?.name);
    setDashboardDesc(activeDashboard?.description);
  }, [activeDashboard]);

  useEffect(() => {
    if (activeDashboard) {
      factorsai.track('VIEW_DASHBOARD', {
        dashboard_name: activeDashboard?.name,
        dashboard_type: activeDashboard?.type,
        dashboard_id: activeDashboard?.id,
        project_id: active_project?.id,
        project_name: active_project?.name
      });
    }
  }, [activeDashboard?.id]);

  const handleOptChange = useCallback(
    (group, data) => {
      setDashboardName(data[0]);
      setDashboardDesc(data[1]);
      changeActiveDashboard(data[2]);
      setSelectVisible(false);
    },
    [dashboards, changeActiveDashboard]
  );

  const fetchUnits = useCallback(() => {
    if (active_project.id && activeDashboard?.id) {
      dispatch(
        fetchActiveDashboardUnits(active_project.id, activeDashboard?.id)
      );
    }
  }, [active_project.id, activeDashboard?.id, dispatch]);

  useEffect(() => {
    fetchUnits();
  }, [fetchUnits]);

  const handleToggleWidgetModal = (val) => {
    setWidgetModalLoading(true);
    setwidgetModal(val);
    // for canvas to load properly before rendering the charts
    setTimeout(() => {
      window.scrollTo(0, 0);
      setWidgetModalLoading(false);
    }, 1000);
  };

  const confirmDeleteDashboard = useCallback(async () => {
    try {
      setDashboardDeleteApi(true);
      await deleteDashboard(active_project.id, activeDashboard?.id);
      setDashboardDeleteApi(false);
      dispatch({ type: DASHBOARD_DELETED, payload: activeDashboard });
      showDeleteDashboardModal(false);
      setDashboardName(dashboards.data[0]?.name);
      setDashboardDesc(dashboards.data[0]?.description);
      changeActiveDashboard(dashboards.data[0]?.id);
    } catch (err) {
      console.log(err);
      setDashboardDeleteApi(false);
    }
  }, [activeDashboard, dispatch, active_project.id]);

  const confirmDeleteWidget = useCallback(async () => {
    try {
      setDeleteApiCalled(true);
      await DeleteUnitFromDashboard(
        active_project.id,
        deleteWidgetModal.dashboard_id,
        deleteWidgetModal.id
      );
      dispatch({ type: WIDGET_DELETED, payload: deleteWidgetModal.id });
      setDeleteApiCalled(false);
      showDeleteWidgetModal(false);
    } catch (err) {
      console.log(err);
      console.log(err.response);
    }
  }, [
    deleteWidgetModal.dashboard_id,
    deleteWidgetModal.id,
    active_project.id,
    dispatch
  ]);

  const toggleDashboardSelect = () => {
    setSelectVisible(!selectVisible);
  };

  const setDashboard = () => (
    <div className={styles.event_selector}>
      {selectVisible ? (
        <GroupSelect2
          groupedProperties={generateDBList()}
          placeholder='Search Dashboard'
          iconColor='#3E516C'
          optionClick={handleOptChange}
          onClickOutside={() => setSelectVisible(false)}
          additionalActions={
            <>
              <Divider className={styles.divider_newdashboard_btn} />
              <Button
                type='text'
                size='large'
                className='w-full'
                icon={<SVG name='plus' />}
                onClick={() => {
                  dispatch({ type: ADD_DASHBOARD_MODAL_OPEN });
                  // setaddDashboardModal(true);
                  // setSelectVisible(false);
                }}
              >
                New Dashboard
              </Button>
            </>
          }
        />
      ) : null}
    </div>
  );

  const setAdditionalactions = (opt) => {
    if (opt[1] === 'edit') {
      handleEditClick(activeDashboard);
    } else if (opt[1] === 'trash') {
      showDeleteDashboardModal(true);
    }
    setMoreOptions(false);
  };

  const additionalActions = () => (
    <div className='fa--query_block--actions-cols flex'>
      <div className='relative'>
        <Button
          type='text'
          size='large'
          onClick={() => setMoreOptions(true)}
          className='ml-1'
          style={{ padding: '0px', height: '32px', width: '32px' }}
        >
          <SVG name='more' size={24} />
        </Button>

        {moreOptions ? (
          <FaSelect
            extraClass={styles.additionalops}
            options={[
              ['Edit Dashboard', 'edit'],
              // ['Pin Dashboard', 'pin'],
              ['Delete Dashboard', 'trash']
            ]}
            optionClick={(val) => setAdditionalactions(val)}
            onClickOutside={() => setMoreOptions(false)}
            posRight
          />
        ) : (
          false
        )}
      </div>
    </div>
  );

  const handleTour = () => {
    userflow.start('c162ed75-0983-41f3-ae56-8aedd7dbbfbd');
  };

  const generateDBList = () => {
    const dashboardList = [
      { label: 'Pinned Dashboards', icon: 'pin', values: [] },
      { label: 'All Dashboards', icon: 'dashboard', values: [] }
    ];

    for (let i = 0; i < dashboards.data.length; i++) {
      if (isPinned) {
        dashboardList[0].values.push([
          dashboards.data[i].name,
          dashboards.data[i].description,
          dashboards.data[i].id
        ]);
        dashboardList[1].values.push([
          dashboards.data[i].name,
          dashboards.data[i].description,
          dashboards.data[i].id
        ]);
      } else {
        dashboardList[1].values.push([
          dashboards.data[i].name,
          dashboards.data[i].description,
          dashboards.data[i].id
        ]);
      }
    }
    return dashboardList;
  };

  if (dashboards.loading || activeDashboardUnits.loading) {
    return (
      <div className='flex justify-center items-center w-full h-64'>
        <Spin size='large' />
      </div>
    );
  }

  if (dashboards.error || activeDashboardUnits.error) {
    return (
      <div className='flex justify-center items-center w-full h-full pt-4 pb-4'>
        <NoDataChart />
      </div>
    );
  }

  const items = [
    {
      label: 'New KPI Report',
      key: 1,
      icon: (
        <div style={{ padding: '0 10px 0 0px' }}>
          <SVG name={`KPI_cq`} size={24} color={'blue'} />
        </div>
      ),
      description: 'Measure performance over time'
    },
    {
      label: 'New Funnel Report',
      key: 2,
      icon: (
        <div style={{ padding: '0 10px 0 0px' }}>
          <SVG name={`funnels_cq`} size={24} color={'blue'} />
        </div>
      ),
      description: 'Track how users navigate'
    },
    {
      label: 'New Attribution Report',
      key: 3,
      icon: (
        <div style={{ padding: '0 10px 0 0px' }}>
          <SVG name={`attributions_cq`} size={24} color={'blue'} />
        </div>
      ),
      description: 'Identify the channels that contribute'
    },
    {
      label: 'New Event Report',
      key: 4,
      icon: (
        <div style={{ padding: '0 10px 0 0px' }}>
          <SVG name={`events_cq`} size={24} color={'blue'} />
        </div>
      ),
      description: 'Track and Chart Events'
    },
    {
      label: 'Saved Report',
      key: 5,
      icon: (
        <div style={{ padding: '0 10px 0 0px' }}>
          {' '}
          <SVG name={'FileSignature'} size={24} color={'blue'} />
        </div>
      ),
      description: 'Select from saved Reports'
    }
  ];
  let HandleMenuItemClick = ({ item, key, keyPath, domEvent }) => {
    /*
      navigatedFromDashboardExistingReports helps to nagivate back to dashboard if closed AnalyseHeader, if and only if That was opened from New Query Menu from Dashboard
    */
    if (key === '0') {
      history.push({
        pathname: '/analyse/' + QUERY_TYPE_KPI,
        state: {
          navigatedFromDashboardExistingReports: true
        }
      });
    } else if (key === '1') {
      history.push({
        pathname: '/analyse/' + QUERY_TYPE_FUNNEL,
        state: {
          navigatedFromDashboardExistingReports: true
        }
      });
    } else if (key === '2') {
      history.push({
        pathname: '/analyse/' + QUERY_TYPE_ATTRIBUTION,
        state: {
          navigatedFromDashboardExistingReports: true
        }
      });
    } else if (key === '3') {
      history.push({
        pathname: '/analyse/' + QUERY_TYPE_EVENT,
        state: {
          navigatedFromDashboardExistingReports: true
        }
      });
    } else if (key === '4') {
      setIsReportsModalOpen((prev) => !prev);
    }
  };
  const menu = (
    <Menu
      onClick={HandleMenuItemClick}
      style={{ borderRadius: '5px', paddingTop: '8px' }}
    >
      {items.map((eachItem, eachKey) => {
        return (
          <>
            {eachKey === items.length - 1 ? (
              <Divider style={{ margin: 0 }} />
            ) : (
              ''
            )}
            <Menu.Item
              icon={eachItem.icon}
              key={eachKey}
              style={{
                margin: '2px 6px 2px 6px',
                display: 'flex',
                flexWrap: 'nowrap',
                borderRadius: '5px'
              }}
            >
              <div style={{ display: 'block' }}>
                {' '}
                <div>{eachItem.label}</div>
                <div style={{ fontSize: '12px', color: '#8692A3' }}>
                  {eachItem.description}
                </div>
              </div>
            </Menu.Item>
          </>
        );
      })}
    </Menu>
  );

  if (dashboards.data.length) {
    return (
      <ErrorBoundary
        fallback={
          <FaErrorComp
            size='medium'
            title='Dashboard Error'
            subtitle='We are facing trouble loading dashboards. Drop us a message on the in-app chat.'
          />
        }
        onError={FaErrorLog}
      >
        {isReportsModalOpen === true ? (
          <ExistingReportsModal
            isReportsModalOpen={isReportsModalOpen}
            setIsReportsModalOpen={setIsReportsModalOpen}
          />
        ) : (
          ''
        )}
        {active_project.id === demoProjectId ? (
          <div className='rounded-lg border-2 h-20 mb-3 mx-10'>
            <Row justify='space-between' className='m-0 p-3'>
              <Col span={projects.length === 1 ? 12 : 18}>
                <img
                  src='assets/icons/welcome.svg'
                  style={{ float: 'left', marginRight: '20px' }}
                />
                <Text type='title' level={6} weight='bold' extraClass='m-0'>
                  Welcome! You just entered a Factors demo project
                </Text>
                {projects.length === 1 ? (
                  <Text type='title' level={7} extraClass='m-0'>
                    These reports have been built with a sample dataset. Use
                    this to start exploring!
                  </Text>
                ) : (
                  <Text type='title' level={7} extraClass='m-0'>
                    To jump back into your Factors project, click on your
                    account card on the{' '}
                    <span className='font-bold'>top right</span> of the screen.
                  </Text>
                )}
              </Col>
              <Col className='mr-2 mt-2'>
                {projects.length === 1 ? (
                  <Button
                    type='default'
                    style={{
                      background: 'white',
                      border: '1px solid #E7E9ED',
                      height: '40px'
                    }}
                    className='m-0 mr-2'
                    onClick={() => setShowProjectModal(true)}
                  >
                    Set up my own Factors project
                  </Button>
                ) : null}

                <Button
                  type='link'
                  style={{
                    background: 'white',
                    // border: '1px solid #E7E9ED',
                    height: '40px'
                  }}
                  className='m-0 mr-2'
                  onClick={() => handleTour()}
                >
                  Take the tour{' '}
                  <SVG
                    name='Arrowright'
                    size={16}
                    extraClass='ml-1'
                    color='blue'
                  />
                </Button>
              </Col>
            </Row>
          </div>
        ) : null}
        <div className='flex items-start justify-between mx-10 my-2'>
          <div className='flex flex-col items-start'>
            <div className='flex items-center'>
              <Button
                className={`${styles.dropdownbtn}`}
                type='text'
                size='large'
                onClick={toggleDashboardSelect}
              >
                <div className={styles.dropdownbtntext + '  text-sm'}>
                  {showDashboardName}
                </div>
                <div className={styles.dropdownbtnicon}>
                  <SVG name='caretDown' size={18} />
                </div>
              </Button>
              <Button
                shape='circle'
                icon={<PlusOutlined style={{ fontSize: '18px' }} />}
                onClick={() => {
                  // setaddDashboardModal(true);
                  // setSelectVisible(false);
                  {
                    dispatch({ type: NEW_DASHBOARD_TEMPLATES_MODAL_OPEN });
                  }
                }}
                className={styles.addNewDashboardButtonProjectDropdown}
              />
            </div>
            {setDashboard()}
            <Text level={7} type='title' weight='medium' color='grey'>
              {showDashboardDesc}
            </Text>
          </div>
          <div className='flex items-center'>
            <Dropdown overlay={menu} placement='bottomRight' trigger={'click'}>
              <Button type='primary'>
                <Space>
                  <SVG name={'plus'} size={16} color='white' />
                  Report
                  {/* <CaretDownOutlined /> */}
                </Space>
              </Button>
            </Dropdown>
            {/* <Dropdown.Button
                menu={menuProps} 
                onClick={(e)=>console.log(e)} 
              >
               
                 <React.Fragment> New Dashboard </React.Fragment>
               
              </Dropdown.Button> */}
            {additionalActions()}
          </div>
        </div>
        <div className='ml-10 mr-4 my-6 flex-1'>
          <DashboardSubMenu
            durationObj={durationObj}
            handleDurationChange={handleDurationChange}
            dashboard={activeDashboard}
            handleEditClick={handleEditClick}
            refreshInProgress={dashboardRefreshState.inProgress}
            oldestRefreshTime={oldestRefreshTime}
            handleRefreshClick={handleRefreshClick}
          />

          <SortableCards
            durationObj={durationObj}
            setwidgetModal={handleToggleWidgetModal}
            showDeleteWidgetModal={showDeleteWidgetModal}
            dashboardRefreshState={dashboardRefreshState}
            setOldestRefreshTime={setOldestRefreshTime}
            onDataLoadSuccess={onDataLoadSuccess}
            handleWidgetRefresh={handleWidgetRefresh}
          />
        </div>

        <ExpandableView
          widgetModalLoading={widgetModalLoading}
          widgetModal={widgetModal}
          setwidgetModal={setwidgetModal}
          durationObj={durationObj}
        />

        <ConfirmationModal
          visible={!!deleteWidgetModal}
          confirmationText='Are you sure you want to delete this widget?'
          onOk={confirmDeleteWidget}
          onCancel={showDeleteWidgetModal.bind(this, false)}
          title='Delete Widget'
          okText='Confirm'
          cancelText='Cancel'
          confirmLoading={deleteApiCalled}
        />
        <ConfirmationModal
          visible={deleteDashboardModal}
          confirmationText='Are you sure you want to delete this Dashboard?'
          onOk={confirmDeleteDashboard}
          onCancel={showDeleteDashboardModal.bind(this, false)}
          title={`Delete Dashboard - ${activeDashboard?.name}`}
          okText='Confirm'
          cancelText='Cancel'
          confirmLoading={dashboardDeleteApi}
        />
        {/* create project modal */}
        <NewProject
          visible={showProjectModal}
          handleCancel={() => setShowProjectModal(false)}
        />
      </ErrorBoundary>
    );
  }

  return null;
}

export default connect(null, { fetchDemoProject, getHubspotContact })(
  ProjectDropdown
);
