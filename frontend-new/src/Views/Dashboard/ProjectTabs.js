import React, { useState, useEffect, useCallback } from 'react';
import { Tabs, Button, Spin, Select } from 'antd';
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
} from '../../components/factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
// import InfoCard from '../../components/InfoCard';
// import { LockOutlined, UnlockOutlined } from '@ant-design/icons';

const { TabPane } = Tabs;

function ProjectTabs({
  setaddDashboardModal,
  handleEditClick,
  durationObj,
  handleDurationChange,
  refreshClicked,
  setRefreshClicked,
}) {
  const [widgetModal, setwidgetModal] = useState(false);
  const [deleteWidgetModal, showDeleteWidgetModal] = useState(false);
  const [deleteApiCalled, setDeleteApiCalled] = useState(false);
  const [widgetModalLoading, setwidgetModalLoading] = useState(false);
  const { active_project } = useSelector((state) => state.global);
  const { dashboards, activeDashboard, activeDashboardUnits } = useSelector(
    (state) => state.dashboard
  );
  const MAX_DASHBOARD_TABS = 5;
  const dispatch = useDispatch();

  const changeActiveDashboard = useCallback(
    (value) => {
      if (parseInt(value) === activeDashboard.id) {
        return false;
      }
      dispatch({
        type: ACTIVE_DASHBOARD_CHANGE,
        payload: dashboards.data.find((d) => d.id === parseInt(value)),
      });
    },
    [dashboards, dispatch, activeDashboard?.id]
  );

  const handleTabChange = useCallback(
    (value) => {
      if (dashboards.data.length > MAX_DASHBOARD_TABS) {
        const dbIndex = dashboards.data.findIndex(
          (elem) => parseInt(elem.id) === parseInt(value)
        );
        if (dbIndex <= MAX_DASHBOARD_TABS - 2) {
          changeActiveDashboard(value);
        } else {
          return false;
        }
      } else {
        changeActiveDashboard(value);
      }
    },
    [dashboards, changeActiveDashboard]
  );

  const fetchUnits = useCallback(() => {
    if (active_project.id && activeDashboard?.id) {
      dispatch(
        fetchActiveDashboardUnits(active_project.id, activeDashboard.id)
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

  const confirmDelete = useCallback(async () => {
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

  const getActiveKey = useCallback(() => {
    if (dashboards.data.length > MAX_DASHBOARD_TABS) {
      const dbIndex = dashboards.data.findIndex(
        (elem) => parseInt(elem.id) === parseInt(activeDashboard.id)
      );
      if (dbIndex <= MAX_DASHBOARD_TABS - 2) {
        return activeDashboard?.id?.toString();
      } else {
        return dashboards.data[MAX_DASHBOARD_TABS - 1].id.toString();
      }
    } else {
      return activeDashboard?.id?.toString();
    }
  }, [activeDashboard, dashboards]);

  const getTabName = useCallback(
    (d, index) => {
      if (dashboards.data.length <= MAX_DASHBOARD_TABS) {
        return d.name;
      } else {
        if (index <= MAX_DASHBOARD_TABS - 2) {
          return d.name;
        } else if (index === MAX_DASHBOARD_TABS - 1) {
          let val = null;
          const isDashboardInDD = dashboards.data
            .slice(index)
            .find((elem) => elem.id === activeDashboard.id);
          if (isDashboardInDD) {
            val = isDashboardInDD.id;
          }
          return (
            <Select
              onSelect={changeActiveDashboard}
              value={val || dashboards.data.slice(index)[0].id}
              className={styles.dashboardDropdown}
            >
              {dashboards.data.slice(index).map((option) => {
                return (
                  <Select.Option value={option.id} key={option.id}>
                    {option.name}
                  </Select.Option>
                );
              })}
            </Select>
          );
        } else {
          return null;
        }
      }
      // return d.name;
    },
    [dashboards?.data, activeDashboard?.id, changeActiveDashboard]
  );

  const operations = (
    <>
      <Button
        data-tour = 'step-4'
        className={styles.operations}
        type='text'
        onClick={() => setaddDashboardModal(true)}
        icon={<SVG name='plus' size={16} color={'grey'} />}
      />
      {/* <Button type="text" icon={<SVG name="edit" size={16} color={'grey'} />} /> */}
    </>
  );

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

  // const Footer = (type) => {
  //   if (type === 'pr') {
  //     return (<div className={''}> <LockOutlined/> Private </div>);
  //   } if (type === 'pv') {
  //     return (<div className={''}> <UnlockOutlined/> Public </div>);
  //   }
  // };

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
          <Tabs
            onChange={handleTabChange}
            activeKey={getActiveKey()}
            className={'fa-tabs--dashboard'}
            tabBarExtraContent={operations}
          >
            {dashboards.data.map((d, index) => {
              return (
                <TabPane tab={getTabName(d, index)} key={d.id}>
                  <div className={'fa-container mt-4 min-h-screen'}>
                    <ErrorBoundary
                      fallback={
                        <FaErrorComp
                          size={'medium'}
                          title={'Dashboard Widget Error'}
                          subtitle={
                            'We are facing trouble loading dashboard widgets. Drop us a message on the in-app chat.'
                          }
                        />
                      }
                      onError={FaErrorLog}
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
                    </ErrorBoundary>
                  </div>
                </TabPane>
              );
            })}
          </Tabs>

          <ExpandableView
            widgetModalLoading={widgetModalLoading}
            widgetModal={widgetModal}
            setwidgetModal={setwidgetModal}
            durationObj={durationObj}
          />

          <ConfirmationModal
            visible={deleteWidgetModal ? true : false}
            confirmationText='Are you sure you want to delete this widget?'
            onOk={confirmDelete}
            onCancel={showDeleteWidgetModal.bind(this, false)}
            title='Delete Widget'
            okText='Confirm'
            cancelText='Cancel'
            confirmLoading={deleteApiCalled}
          />
        </ErrorBoundary>
      </>
    );
  }

  return null;
}

export default ProjectTabs;
