import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { Button, Spin } from 'antd';
import { connect, useDispatch } from 'react-redux';
import { bindActionCreators } from 'redux';
import { useHistory } from 'react-router-dom';
import { isEmpty } from 'lodash';

import { SVG, Text } from 'Components/factorsComponents';
import FaDatepicker from 'Components/FaDatepicker';
import { fetchAttributionActiveUnits } from 'Attribution/state/services';
import { QUERY_TYPE_ATTRIBUTION } from 'Utils/constants';
import NoReports from './NoReports';
import SortableCards from './SortableCards';
import { ATTRIBUTION_ROUTES } from 'Attribution/utils/constants';
import { setItemToLocalStorage } from 'Utils/localStorage.helpers';
import { getDashboardDateRange } from 'Views/Dashboard/utils';
import { DASHBOARD_KEYS } from 'Constants/localStorage.constants';
import ConfirmationModal from 'Components/ConfirmationModal';
import { DeleteUnitFromDashboard } from 'Reducers/dashboard/services';
import { deleteReport } from 'Reducers/coreQuery/services';
import {
  ATTRIBUTION_QUERY_DELETED,
  ATTRIBUTION_WIDGET_DELETED
} from 'Attribution/state/action.constants';
import { PathUrls } from '../../../../routes/pathUrls';

function Reports({
  attributionDashboardUnits,
  savedQueries,
  savedQueriesLoading,
  currentProjectSettingsLoading,
  currentProjectSettings,
  activeProject
}) {
  const history = useHistory();
  const dispatch = useDispatch();
  const [deleteWidgetModal, showDeleteWidgetModal] = useState(false);
  const [deleteApiCalled, setDeleteApiCalled] = useState(false);
  const [durationObj, setDurationObj] = useState(getDashboardDateRange());

  const handleDurationChange = (dates) => {
    let from;
    let to;
    const { startDate, endDate } = dates;
    // setOldestRefreshTime(null);
    if (Array.isArray(dates.startDate)) {
      from = get(startDate, 0);
      to = get(startDate, 1);
    } else {
      from = startDate;
      to = endDate;
    }

    setDurationObj((currState) => {
      const newState = {
        ...currState,
        from,
        to,
        dateType: dates.dateType
      };
      setItemToLocalStorage(
        DASHBOARD_KEYS.DASHBOARD_DURATION,
        JSON.stringify(newState)
      );
      return newState;
    });
  };

  useEffect(() => {
    if (
      !currentProjectSettingsLoading &&
      currentProjectSettings &&
      !isEmpty(currentProjectSettings)
    ) {
      if (!currentProjectSettings?.attribution_config) {
        history.replace(ATTRIBUTION_ROUTES.base);
      }
    }
  }, [currentProjectSettings, currentProjectSettingsLoading]);

  const activeUnits = useMemo(() => {
    return attributionDashboardUnits.data.filter(
      (elem) =>
        savedQueries.findIndex(
          (sq) =>
            sq.id === elem.query_id && sq.query.cl === QUERY_TYPE_ATTRIBUTION
        ) > -1
    );
  }, [attributionDashboardUnits, savedQueries]);

  const deleteWidget = useCallback(async () => {
    try {
      setDeleteApiCalled(true);
      await DeleteUnitFromDashboard(
        activeProject.id,
        deleteWidgetModal.dashboard_id,
        deleteWidgetModal.id
      );

      await deleteReport({
        project_id: activeProject.id,
        queryId: deleteWidgetModal.id
      });
      dispatch({
        type: ATTRIBUTION_WIDGET_DELETED,
        payload: deleteWidgetModal.id
      });
      dispatch({ type: ATTRIBUTION_QUERY_DELETED, id: deleteWidgetModal.id });
      setDeleteApiCalled(false);
      showDeleteWidgetModal(false);
    } catch (err) {
      console.log(err);
      console.log(err.response);
    }
  }, [
    deleteWidgetModal.dashboard_id,
    deleteWidgetModal.id,
    activeProject.id,
    dispatch
  ]);

  if (
    attributionDashboardUnits?.loading ||
    savedQueriesLoading ||
    currentProjectSettingsLoading
  ) {
    return (
      <div className='flex items-center justify-center h-full w-full'>
        <div className='w-full h-64 flex items-center justify-center'>
          <Spin size='large' />
        </div>
      </div>
    );
  }

  if (!activeUnits || activeUnits?.length <= 0) {
    return <NoReports />;
  }

  return (
    <div className='flex flex-col items-center'>
      <div className='flex w-full justify-between items-center px-8'>
        <div className='flex items-center gap-4'>
          <Text
            type='title'
            level={6}
            weight='bold'
            color='black'
            extraClass='m-0'
          >
            Attribution Reports
          </Text>
          <FaDatepicker
            customPicker
            presetRange
            range={{
              startDate: durationObj.from,
              endDate: durationObj.to
            }}
            quarterPicker
            monthPicker
            buttonSize='large'
            placement='bottomRight'
            className='mr-2'
            onSelect={handleDurationChange}
          />
        </div>
        <div className='flex items-center gap-2'>
          <Button
            type='link'
            size='large'
            onClick={() => history.push(PathUrls.ConfigureAttribution)}
          >
            Configuration
          </Button>
          <Button
            type='primary'
            size='large'
            onClick={() => history.push(ATTRIBUTION_ROUTES.report)}
          >
            <SVG name='plus' color='white' className='w-full' /> Add Report
          </Button>
          {/* <Button
            type='text'
            size='large'
            className='ml-1'
            style={{ padding: '4px 6px' }}
          >
            <SVG name='more' size={24} />
          </Button> */}
        </div>
      </div>
      <div className='w-full px-8 mt-2 flex flex-col'>
        {/* sortable cards */}
        <SortableCards
          activeUnits={activeUnits}
          durationObj={durationObj}
          showDeleteWidgetModal={showDeleteWidgetModal}
        />
      </div>
      <ConfirmationModal
        visible={!!deleteWidgetModal}
        confirmationText='Are you sure you want to delete this widget?'
        onOk={deleteWidget}
        onCancel={showDeleteWidgetModal.bind(this, false)}
        title='Delete Widget'
        okText='Confirm'
        cancelText='Cancel'
        confirmLoading={deleteApiCalled}
      />
    </div>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  activeDashboard: state.dashboard.activeDashboard,
  attributionDashboardUnits:
    state.attributionDashboard.attributionDashboardUnits,
  savedQueries: state.attributionDashboard.attributionQueries.data,
  savedQueriesLoading: state.attributionDashboard.attributionQueries.loading,
  attributionDashboard: state.attributionDashboard,
  currentProjectSettings: state.global.currentProjectSettings,
  currentProjectSettingsLoading: state.global.currentProjectSettingsLoading
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      fetchAttributionActiveUnits
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(Reports);
