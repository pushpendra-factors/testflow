import React, { useState, useEffect, useCallback } from 'react';
import { Button, Spin } from 'antd';
import { useSelector, useDispatch } from 'react-redux';
import {
  fetchActiveDashboardUnits,
  DeleteUnitFromDashboard,
} from '../../reducers/dashboard/services';
import { ACTIVE_DASHBOARD_CHANGE, WIDGET_DELETED } from '../../reducers/types';
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
  Text,
} from '../../components/factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import FaSelect from 'Components/FaSelect';
import GroupSelect2 from '../../components/QueryComposer/GroupSelect2';
import { deleteDashboard } from '../../reducers/dashboard/services';
import { DASHBOARD_DELETED } from '../../reducers/types';
import factorsai from 'factorsai';

function ProjectDropdown({
  setaddDashboardModal,
  handleEditClick,
  durationObj,
  handleDurationChange,
  refreshClicked,
  setRefreshClicked,
  isPinned = false,
}) {
  const [moreOptions, setMoreOptions] = useState(false);
  const [widgetModal, setwidgetModal] = useState(false);
  const [deleteWidgetModal, showDeleteWidgetModal] = useState(false);
  const [deleteApiCalled, setDeleteApiCalled] = useState(false);
  const [widgetModalLoading, setwidgetModalLoading] = useState(false);
  const { active_project } = useSelector((state) => state.global);
  const { dashboards, activeDashboard, activeDashboardUnits } = useSelector(
    (state) => state.dashboard
  );
  const [selectVisible, setSelectVisible] = useState(false);
  const [showDashboardName, setDashboardName] = useState('');
  const [showDashboardDesc, setDashboardDesc] = useState('');
  const [deleteDashboardModal, showdeleteDashboardModal] = useState(false);
  const [dashboardDeleteApi, setDashboardDeleteApi] = useState(false);

  const dispatch = useDispatch();

  const changeActiveDashboard = useCallback(
    (val) => {
      if (parseInt(val) === activeDashboard?.id) {
        return false;
      }
      dispatch({
        type: ACTIVE_DASHBOARD_CHANGE,
        payload: dashboards.data.find((d) => d.id === parseInt(val)),
      });
    },
    [dashboards, dispatch, activeDashboard?.id]
  );

  useEffect(() => {
    setDashboardName(activeDashboard?.name);
    setDashboardDesc(activeDashboard?.description);
  }, [activeDashboard]);



  useEffect(()=>{ 
    if(activeDashboard){ 
      //Factors VIEW_DASHBOARD tracking
    factorsai.track('VIEW_DASHBOARD',{'dashboard_name': activeDashboard?.name, 'dashboard_type': activeDashboard?.type, 'dashboard_id': activeDashboard?.id});
    }
  },[activeDashboard]);


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
    setwidgetModalLoading(true);
    setwidgetModal(val);
    // for canvas to load properly before rendering the charts
    setTimeout(() => {
      window.scrollTo(0, 0);
      setwidgetModalLoading(false);
    }, 1000);
  };

  const confirmDeleteDashboard = useCallback(async () => {
    try {
      setDashboardDeleteApi(true);
      await deleteDashboard(active_project.id, activeDashboard?.id);
      setDashboardDeleteApi(false);
      dispatch({ type: DASHBOARD_DELETED, payload: activeDashboard });
      showdeleteDashboardModal(false);
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
    dispatch,
  ]);

  const toggleDashboardSelect = () => {
    setSelectVisible(!selectVisible);
  };

  const setDashboard = () => {
    return (
      <div className={styles.event_selector}>
        {selectVisible ? (
          <GroupSelect2
            groupedProperties={generateDBList()}
            placeholder='Search Dashboard'
            iconColor='#3E516C'
            optionClick={handleOptChange}
            onClickOutside={() => setSelectVisible(false)}
            additionalActions={
              <Button
                type='text'
                size='large'
                className={`w-full`}
                icon={<SVG name='plus' />}
                onClick={() => {
                  setaddDashboardModal(true);
                  setSelectVisible(false);
                }}
              >
                New Dashboard
              </Button>
            }
          ></GroupSelect2>
        ) : null}
      </div>
    );
  };

  const setAdditionalactions = (opt) => {
    if (opt[1] === 'edit') {
      handleEditClick(activeDashboard);
    } else if (opt[1] === 'trash') {
      showdeleteDashboardModal(true);
    }
    setMoreOptions(false);
  };

  const additionalActions = () => {
    return (
      <div className={'fa--query_block--actions-cols flex'}>
        <div className={`relative`}>
          <Button
            type='text'
            size={'large'}
            onClick={() => setMoreOptions(true)}
            className={`btn-custom ml-1`}
          >
            <SVG name='more' />
          </Button>

          {moreOptions ? (
            <FaSelect
              extraClass={styles.additionalops}
              options={[
                ['Edit Details', 'edit'],
                // ['Pin Dashboard', 'pin'],
                ['Delete Dashboard', 'trash'],
              ]}
              optionClick={(val) => setAdditionalactions(val)}
              onClickOutside={() => setMoreOptions(false)}
              posRight={true}
            ></FaSelect>
          ) : (
            false
          )}
        </div>
      </div>
    );
  };

  const generateDBList = () => {
    const dashboardList = [
      { label: 'Pinned Dashboards', icon: 'pin', values: [] },
      { label: 'All Dashboards', icon: 'dashboard', values: [] },
    ];

    for (let i = 0; i < dashboards.data.length; i++) {
      if (isPinned) {
        dashboardList[0].values.push([
          dashboards.data[i].name,
          dashboards.data[i].description,
          dashboards.data[i].id,
        ]);
        dashboardList[1].values.push([
          dashboards.data[i].name,
          dashboards.data[i].description,
          dashboards.data[i].id,
        ]);
      } else
        dashboardList[1].values.push([
          dashboards.data[i].name,
          dashboards.data[i].description,
          dashboards.data[i].id,
        ]);
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
      <div className='flex justify-center items-center w-full h-64'>
        <NoDataChart />
      </div>
    );
  } 

  if (dashboards.data.length) {
    return (
      <>
        <ErrorBoundary
          fallback={
            <FaErrorComp
              size={'medium'}
              title={'Dashboard Error'}
              subtitle={
                'We are facing trouble loading dashboards. Drop us a message on the in-app chat.'
              }
            />
          }
          onError={FaErrorLog}
        >
          <div className={`flex items-center justify-between mx-10 mb-4`}>
            <div className={'flex flex-col items-start ml-4'}>
              <Button
                className={`${styles.dropdownbtn}`}
                type='text'
                size={'large'}
                onClick={toggleDashboardSelect}
              >
                {showDashboardName}
                <SVG name='caretDown' size={20} />
              </Button>
              {setDashboard()}
              <Text
                level={7}
                type={'title'}
                weight={'medium'}
                color={'grey'}
              >
                {showDashboardDesc}
              </Text>
            </div>
            <div className='flex items-center'>
              <Button
                type='primary'
                size={'large'}
                onClick={() => setaddDashboardModal(true)}
                icon={<SVG name='plus' size={16} color={'white'} />}
              >
                New Dashboard
              </Button>
              {additionalActions()}
            </div>
          </div>
          <div
            className={'pl-10 pr-6 py-6 flex-1'}
            style={{ backgroundColor: '#f6f6f8' }}
          >
            <DashboardSubMenu
              durationObj={durationObj}
              handleDurationChange={handleDurationChange}
              dashboard={activeDashboard}
              handleEditClick={handleEditClick}
              refreshClicked={refreshClicked}
              setRefreshClicked={setRefreshClicked}
            />
            <SortableCards
              durationObj={durationObj}
              setwidgetModal={handleToggleWidgetModal}
              showDeleteWidgetModal={showDeleteWidgetModal}
              refreshClicked={refreshClicked}
              setRefreshClicked={setRefreshClicked}
            />
          </div>

          <ExpandableView
            widgetModalLoading={widgetModalLoading}
            widgetModal={widgetModal}
            setwidgetModal={setwidgetModal}
            durationObj={durationObj}
          />

          <ConfirmationModal
            visible={deleteWidgetModal ? true : false}
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
            onCancel={showdeleteDashboardModal.bind(this, false)}
            title={`Delete Dashboard - ${activeDashboard?.name}`}
            okText='Confirm'
            cancelText='Cancel'
            confirmLoading={dashboardDeleteApi}
          />
        </ErrorBoundary>
      </>
    );
  }

  return null;
}

export default ProjectDropdown;
