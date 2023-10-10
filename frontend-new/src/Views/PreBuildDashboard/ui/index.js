import React, { useCallback, useEffect, useState } from 'react';
import { ErrorBoundary } from 'react-error-boundary';
import {
  FaErrorComp,
  SVG,
  FaErrorLog,
  Text
} from 'Components/factorsComponents';
import { Button, Dropdown, Menu } from 'antd';
import SortableCards from './Widget/SortableCards';
import SubMenu from './Widget/SubMenu';
import { useHistory } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import { getQuickDashboardDateRange } from 'Views/Dashboard/utils';
import { selectActivePreDashboard } from 'Reducers/dashboard/selectors';
import { fetchActiveDashboardConfig } from '../state/services';
import { get } from 'lodash';
import { setItemToLocalStorage } from 'Utils/localStorage.helpers';
import { DASHBOARD_KEYS } from 'Constants/localStorage.constants';

const dashboardRefreshInitialState = {
  inProgress: false,
  widgetIdGettingFetched: null,
  widgetIdsLeftToBeFetched: [],
  widgetIdsAlreadyFetched: []
};



const PreBuildDashboard = ({}) => {
  const history = useHistory();
  const dispatch = useDispatch();
  const [deleteWidgetModal, showDeleteWidgetModal] = useState(false);
  const [deleteApiCalled, setDeleteApiCalled] = useState(false);
  const [durationObj, setDurationObj] = useState(getQuickDashboardDateRange());
  const [oldestRefreshTime, setOldestRefreshTime] = useState(null);


  const activeDashboard = useSelector((state) => selectActivePreDashboard(state));
  const { active_project } = useSelector((state) => state.global);
  const config = useSelector((state) => state.preBuildDashboardConfig.config.data.result);
  const widget = useSelector((state) => state.preBuildDashboardConfig.widget);

  const fetchConfig = useCallback(() => {
    if (active_project.id && activeDashboard?.inter_id) {
      dispatch(
        fetchActiveDashboardConfig(active_project?.id, activeDashboard?.inter_id)
      );
    }
  }, [active_project.id, activeDashboard?.inter_id, dispatch]);

  useEffect(() => {
    fetchConfig();
  }, [fetchConfig]);

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
  

  const menu = (
    <Menu
      // onClick={HandleMenuItemClick}
      // style={{ borderRadius: '5px', paddingTop: '8px' }}
    >
      
        <Menu.Item
        >
          <div>{'label1'}</div>
        </Menu.Item>
        <Menu.Item
        >
          <div>{'label2'}</div>
        </Menu.Item>
        <Menu.Item
        >
          <div>{'label3'}</div>
        </Menu.Item>
    </Menu>
  );
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
            <Text level={4} type='title' weight='medium'>
              {activeDashboard?.name}
            </Text>
            <div className='w-3/4'>
              <Text level={7} type='title' weight='medium' color='grey'>
               {activeDashboard?.description}
              </Text>
            </div>
          </div>
        </div>
        {/* <div>
          <Dropdown overlay={menu} placement='bottomRight'>
            <Button type='text' icon={<SVG name={'threedot'} size={25} />} />
          </Dropdown>
        </div> */}
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
};

PreBuildDashboard.propTypes = {};

export default PreBuildDashboard;
