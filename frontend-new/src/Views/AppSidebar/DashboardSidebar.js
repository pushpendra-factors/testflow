import React, { memo, useCallback, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import cx from 'classnames';
import { Button } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import {
  selectDashboardListFilteredBySearchText,
  selectDashboardList,
  selectActiveDashboard
} from 'Reducers/dashboard/selectors';
import styles from './index.module.scss';

import SidebarSearch from './SidebarSearch';
import { changeActiveDashboard } from 'Reducers/dashboard/services';
import { NEW_DASHBOARD_TEMPLATES_MODAL_OPEN } from 'Reducers/types';
import SidebarMenuItem from './SidebarMenuItem';

const DashboardItem = ({ dashboard }) => {
  const dispatch = useDispatch();
  const activeDashboard = useSelector((state) => selectActiveDashboard(state));
  const dashboards = useSelector((state) => selectDashboardList(state));

  const handleActiveDashboardChange = useCallback(() => {
    const selectedDashboard = dashboards.find((d) => d.id === dashboard.id);
    dispatch(changeActiveDashboard(selectedDashboard));
  }, [dashboard, dashboards, dispatch]);

  const isActive = activeDashboard.id === dashboard.id;

  return (
    <div
      role='button'
      onClick={handleActiveDashboardChange}
      className={cx(
        'py-2 cursor-pointer rounded-md pl-8 pr-2 flex justify-between col-gap-2 items-center',
        {
          [styles['active']]: isActive
        }
      )}
    >
      <SidebarMenuItem text={dashboard.name} />
      {isActive && <SVG size={16} color='#595959' name='arrowright' />}
    </div>
  );
};

const DashboardSidebar = () => {
  const [searchText, setSearchText] = useState('');
  const dispatch = useDispatch();

  const filteredDashboardList = useSelector((state) =>
    selectDashboardListFilteredBySearchText(state, searchText)
  );

  return (
    <div className='flex flex-col row-gap-5 px-2'>
      <SidebarSearch setSearchText={setSearchText} searchText={searchText} />
      <div
        className={cx(
          'flex flex-col row-gap-3 overflow-auto',
          styles['dashboard-list-container']
        )}
      >
        {filteredDashboardList.map((dashboard) => {
          return <DashboardItem dashboard={dashboard} key={dashboard.id} />;
        })}
      </div>
      <Button
        className={cx(
          'flex col-gap-2 items-center',
          styles['sidebar-action-button']
        )}
        type='secondary'
        onClick={() => {
          dispatch({ type: NEW_DASHBOARD_TEMPLATES_MODAL_OPEN });
        }}
      >
        <SVG name={'plus'} size={16} color='#1890FF' />
        <Text level={7} type='title' color='brand-color-6' extraClass='mb-0'>
          New Dashboard
        </Text>
      </Button>
    </div>
  );
};

export default memo(DashboardSidebar);
