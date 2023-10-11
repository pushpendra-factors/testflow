import React, { useState, useEffect, useCallback } from 'react';
import { Button, Col, Divider, Row, Spin } from 'antd';
import { useSelector, useDispatch, connect } from 'react-redux';
import { ErrorBoundary } from 'react-error-boundary';
import FaSelect from 'Components/FaSelect';
import factorsai from 'factorsai';
import { getHubspotContact } from 'Reducers/global';
import {
  fetchActiveDashboardUnits,
  DeleteUnitFromDashboard,
  deleteDashboard
} from '../../reducers/dashboard/services';
import {
  WIDGET_DELETED,
  DASHBOARD_DELETED,
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
import ExistingReportsModal from './ExistingReportsModal';
import { changeActiveDashboard as changeActiveDashboardService } from 'Reducers/dashboard/services';
import NewReportButton from './NewReportButton';
import { useParams } from 'react-router-dom';
import { selectActiveDashboard, selectDashboardList } from 'Reducers/dashboard/selectors';

function ProjectDropdown({
  setaddDashboardModal,
  handleEditClick,
  durationObj,
  handleDurationChange,
  isPinned = false,
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
  const { activeDashboardUnits } = useSelector(
    (state) => state.dashboard
  );
  const activeDashboard = useSelector((state) => selectActiveDashboard(state));
  const dashboards = useSelector((state) => selectDashboardList(state));
  const [selectVisible, setSelectVisible] = useState(false);
  const [showDashboardName, setDashboardName] = useState('');
  const [showDashboardDesc, setDashboardDesc] = useState('');
  const [deleteDashboardModal, showDeleteDashboardModal] = useState(false);
  const [dashboardDeleteApi, setDashboardDeleteApi] = useState(false);
  const [showProjectModal, setShowProjectModal] = useState(false);

  const [isReportsModalOpen, setIsReportsModalOpen] = useState(false);
  const dispatch = useDispatch();

  const { dashboard_id } = useParams();

  const { agent_details } = useSelector((state) => state.agent);

  const changeActiveDashboard = useCallback(
    (val) => {
      if (val === activeDashboard?.id) {
        return false;
      }
      resetDashboardRefreshState();
      setOldestRefreshTime(null);
      const selectedDashboard = dashboards.find((d) => d.id === val);
      dispatch(changeActiveDashboardService(selectedDashboard));
    },
    [
      activeDashboard?.id,
      resetDashboardRefreshState,
      setOldestRefreshTime,
      dashboards,
      dispatch
    ]
  );

  useEffect(() => {
    setDashboardName(activeDashboard?.name);
    setDashboardDesc(activeDashboard?.description);
  }, [activeDashboard]);

  useEffect(() => {
    if (dashboard_id) {
      changeActiveDashboard(dashboard_id);
    }
  }, [dashboard_id, activeDashboard]);

  useEffect(() => {
    if (activeDashboard) {
      factorsai.track('VIEW_DASHBOARD', {
        email_id: agent_details?.email,
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
      setDashboardName(dashboards[0]?.name);
      setDashboardDesc(dashboards[0]?.description);
      changeActiveDashboard(dashboards[0]?.id);
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
            showIcon
          />
        ) : (
          false
        )}
      </div>
    </div>
  );

  const handleTour = () => {
    // userflow.start('c162ed75-0983-41f3-ae56-8aedd7dbbfbd');
  };

  const generateDBList = () => {
    const dashboardList = [
      { label: 'Pinned Dashboards', icon: 'pin', values: [] },
      { label: 'All Dashboards', icon: 'dashboard', values: [] }
    ];

    for (let i = 0; i < dashboards.length; i++) {
      if (isPinned) {
        dashboardList[0].values.push([
          dashboards[i].name,
          dashboards[i].description,
          dashboards[i].id
        ]);
        dashboardList[1].values.push([
          dashboards[i].name,
          dashboards[i].description,
          dashboards[i].id
        ]);
      } else {
        dashboardList[1].values.push([
          dashboards[i].name,
          dashboards[i].description,
          dashboards[i].id
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

  if (dashboards.length) {
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

        <div className='flex items-start justify-between'>
          <div className='flex flex-col items-start'>
            <div className='flex items-center'>
              <Text
                color='character-primary'
                level={4}
                weight='bold'
                extraClass='mb-0'
                type='title'
                id={'fa-at-text--dashboard-title'}
              >
                {showDashboardName}
              </Text>
            </div>
            {setDashboard()}
            <Text
              level={7}
              type='title'
              weight='medium'
              color='grey'
              id={'fa-at-text--dashboard-desc'}
            >
              {showDashboardDesc}
            </Text>
          </div>
          <div className='flex items-center'>
            <NewReportButton
              showSavedReport={true}
              setIsReportsModalOpen={setIsReportsModalOpen}
            />
            {additionalActions()}
          </div>
        </div>
        <div className='my-6 flex-1'>
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

export default connect(null, { getHubspotContact })(ProjectDropdown);
