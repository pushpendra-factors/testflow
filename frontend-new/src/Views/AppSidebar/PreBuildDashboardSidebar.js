import React, { memo, useCallback, useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import cx from 'classnames';
import { Button } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import {
  selectPreDashboardListFilteredBySearchText,
  selectPreDashboardList,
  selectActivePreDashboard
} from 'Reducers/dashboard/selectors';
import styles from './index.module.scss';

import SidebarMenuItem from './SidebarMenuItem';
import { changeActivePreDashboard } from 'Views/PreBuildDashboard/state/services';

const DashboardItem = ({ dashboard }) => {
  const dispatch = useDispatch();
  const activeDashboard = useSelector((state) => selectActivePreDashboard(state));
  const dashboards = useSelector((state) => selectPreDashboardList(state));

  const handleActiveDashboardChange = useCallback(() => {
    const selectedDashboard = dashboards.find((d) => d.id === dashboard.id);
    dispatch(changeActivePreDashboard(selectedDashboard));
  }, [dashboard, dashboards, dispatch]);

  const isActive = activeDashboard.id === dashboard.id;

  useEffect(() => {
    if(!isActive) {
      dispatch(changeActivePreDashboard(dashboards?.[0]));
    }
  },[dashboards, dispatch, isActive])

  return (
    <SidebarMenuItem
      text={dashboard.name}
      onClick={handleActiveDashboardChange}
      isActive={isActive}
    />
  );
};

const DashboardSidebar = () => {
  const [searchText, setSearchText] = useState('');
  const dispatch = useDispatch();

  const filteredDashboardList = useSelector((state) =>
    selectPreDashboardListFilteredBySearchText(state, searchText)
  );

  return (
    <div className='flex flex-col row-gap-5 px-4'>
      <div
        className={cx(
          'flex flex-col row-gap-1 overflow-auto',
          styles['dashboard-list-container']
        )}
      >
        {filteredDashboardList.map((dashboard) => {
          return <DashboardItem dashboard={dashboard} key={dashboard.id} />;
        })}
      </div>
    </div>
  );
};

export default memo(DashboardSidebar);
