import React, { useCallback, useEffect, useState } from 'react';
import { ErrorBoundary } from 'react-error-boundary';
import {
  FaErrorComp,
  SVG,
  FaErrorLog,
  Text
} from 'Components/factorsComponents';
import { Button, Dropdown, Menu, Spin, Tooltip } from 'antd';
import { useHistory } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import { getQuickDashboardDateRange } from 'Views/Dashboard/utils';
import { selectActivePreDashboard } from 'Reducers/dashboard/selectors';
import { get } from 'lodash';
import { setItemToLocalStorage } from 'Utils/localStorage.helpers';
import { DASHBOARD_KEYS } from 'Constants/localStorage.constants';
import NoDataChart from 'Components/NoDataChart';
import useFeatureLock from 'hooks/useFeatureLock';
import { FEATURES } from 'Constants/plans.constants';
import { PathUrls } from 'Routes/pathUrls';
import { TOOLTIP_CONSTANTS } from 'Constants/tooltips.constans';
import { InfoCircleOutlined } from '@ant-design/icons';
import { fetchActiveDashboardConfig } from '../state/services';
import SubMenu from './Widget/SubMenu';
import SortableCards from './Widget/SortableCards';

const dashboardRefreshInitialState = {
  inProgress: false,
  widgetIdGettingFetched: null,
  widgetIdsLeftToBeFetched: [],
  widgetIdsAlreadyFetched: []
};

function PreBuildDashboard({}) {
  const history = useHistory();
  const dispatch = useDispatch();
  const [durationObj, setDurationObj] = useState(getQuickDashboardDateRange());
  const [oldestRefreshTime, setOldestRefreshTime] = useState(null);

  const activeDashboard = useSelector((state) =>
    selectActivePreDashboard(state)
  );
  const { active_project } = useSelector((state) => state.global);
  const config = useSelector(
    (state) => state.preBuildDashboardConfig.config.data.result
  );
  const widget = useSelector((state) => state.preBuildDashboardConfig.widget);
  const predefinedConfigData = useSelector(
    (state) => state.preBuildDashboardConfig.config
  );

  const fetchConfig = useCallback(() => {
    if (active_project.id && activeDashboard?.inter_id) {
      dispatch(
        fetchActiveDashboardConfig(
          active_project?.id,
          activeDashboard?.inter_id
        )
      );
    }
  }, [active_project.id, activeDashboard?.inter_id, dispatch]);

  const { isFeatureLocked: isWebAnalyticsLocked } = useFeatureLock(
    FEATURES.FEATURE_WEB_ANALYTICS_DASHBOARD
  );

  useEffect(() => {
    if (isWebAnalyticsLocked) {
      history.push(PathUrls.Dashboard);
    }
  }, [isWebAnalyticsLocked]);

  useEffect(() => {
    fetchConfig();
  }, [fetchConfig, activeDashboard?.id]);

  const [dashboardRefreshState, setDashboardRefreshState] = useState(
    dashboardRefreshInitialState
  );

  const resetDashboardRefreshState = useCallback(() => {
    setDashboardRefreshState(dashboardRefreshInitialState);
  }, []);

  const handleDurationChange = useCallback(
    (dates) => {
      let from;
      let to;
      const { startDate, endDate } = dates;
      setOldestRefreshTime(null);
      resetDashboardRefreshState();
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
          DASHBOARD_KEYS.QUICK_DASHBOARD_DURATION,
          JSON.stringify(newState)
        );
        return newState;
      });
    },
    [resetDashboardRefreshState]
  );

  const onDataLoadSuccess = useCallback(({ unitId }) => {
    setDashboardRefreshState((currState) => {
      if (currState.inProgress) {
        return {
          inProgress: currState.widgetIdsLeftToBeFetched.length > 0,
          widgetIdsAlreadyFetched: [
            ...currState.widgetIdsAlreadyFetched,
            unitId
          ],
          widgetIdGettingFetched:
            currState.widgetIdsLeftToBeFetched.length > 0
              ? currState.widgetIdsLeftToBeFetched[0]
              : null,
          widgetIdsLeftToBeFetched: currState.widgetIdsLeftToBeFetched.slice(1)
        };
      }
      return dashboardRefreshInitialState;
    });
  }, []);

  if (predefinedConfigData?.loading) {
    return (
      <div className='flex justify-center items-center w-full h-64'>
        <Spin size='large' />
      </div>
    );
  }

  if (predefinedConfigData?.error) {
    return (
      <div className='flex justify-center items-center w-full h-full pt-4 pb-4'>
        <NoDataChart />
      </div>
    );
  }

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
      <div className='flex items-start justify-between'>
        <div className='flex flex-col items-start'>
          <div>
            <div className='flex justify-center'>
              <Text level={4} type='title' weight='medium'>
                {activeDashboard?.name}
              </Text>
              <Tooltip
                className='mt-2 ml-1'
                title='This is a pre-made dashboard and can not be edited.'
                placement='bottom'
                color={TOOLTIP_CONSTANTS.DARK}
              >
                <InfoCircleOutlined
                  style={{ fontSize: '18px', color: '#8C8C8C' }}
                />
              </Tooltip>
            </div>
            <div className='w-3/4'>
              <Text level={7} type='title' weight='medium' color='grey'>
                {activeDashboard?.description}
              </Text>
            </div>
          </div>
        </div>
      </div>
      <div className='my-6 flex-1'>
        <SubMenu
          config={config}
          durationObj={durationObj}
          handleDurationChange={handleDurationChange}
          activeDashboard={activeDashboard}
        />
        <SortableCards
          widget={widget}
          durationObj={durationObj}
          handleDurationChange={handleDurationChange}
          setOldestRefreshTime={setOldestRefreshTime}
          dashboardRefreshState={dashboardRefreshState}
          onDataLoadSuccess={onDataLoadSuccess}
        />
      </div>
    </ErrorBoundary>
  );
}

PreBuildDashboard.propTypes = {};

export default PreBuildDashboard;
