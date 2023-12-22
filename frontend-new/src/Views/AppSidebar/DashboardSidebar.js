import React, { memo, useCallback, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import cx from 'classnames';
import { Button } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import {
  selectDashboardListFilteredBySearchText,
  selectDashboardList,
  selectActiveDashboard,
  selectAreDraftsSelected
} from 'Reducers/dashboard/selectors';

import { changeActiveDashboard } from 'Reducers/dashboard/services';
import { NEW_DASHBOARD_TEMPLATES_MODAL_OPEN } from 'Reducers/types';
import { makeDraftsActiveAction } from 'Reducers/dashboard/actions';
import { useHistory } from 'react-router-dom';
import { PathUrls } from 'Routes/pathUrls';
import SidebarMenuItem from './SidebarMenuItem';
import SidebarSearch from './SidebarSearch';
import styles from './index.module.scss';

function DashboardItem({ dashboard }) {
  const dispatch = useDispatch();
  const history = useHistory();
  const activeDashboard = useSelector((state) => selectActiveDashboard(state));
  const dashboards = useSelector((state) => selectDashboardList(state));
  const areDraftsSelected = useSelector((state) =>
    selectAreDraftsSelected(state)
  );

  const handleActiveDashboardChange = useCallback(() => {
    const selectedDashboard = dashboards.find((d) => d.id === dashboard.id);
    history.replace(`${PathUrls.Dashboard}/${selectedDashboard.id}`);
    dispatch(changeActiveDashboard(selectedDashboard));
  }, [dashboard, dashboards, dispatch]);

  const isActive =
    activeDashboard?.id === dashboard?.id && areDraftsSelected === false;

  return (
    <SidebarMenuItem
      text={dashboard.name}
      onClick={handleActiveDashboardChange}
      isActive={isActive}
    />
  );
}

function DashboardSidebar() {
  const [searchText, setSearchText] = useState('');
  const areDraftsSelected = useSelector((state) =>
    selectAreDraftsSelected(state)
  );
  const queries = useSelector((state) => state.queries.data);
  const dispatch = useDispatch();

  const filteredDashboardList = useSelector((state) =>
    selectDashboardListFilteredBySearchText(state, searchText)
  );

  const handleDraftsClick = () => {
    dispatch(makeDraftsActiveAction());
  };

  return (
    <div className='flex flex-col row-gap-2'>
      <div role='button' onClick={handleDraftsClick} className='px-4 w-full'>
        <div
          className={cx(
            'flex col-gap-1 cursor-pointer py-2 rounded-md items-center w-full px-2',
            styles['draft-title'],
            {
              [styles['item-active']]: areDraftsSelected
            }
          )}
        >
          <SVG name='drafts' />
          <Text
            color='character-primary'
            type='title'
            level={7}
            extraClass='mb-0'
          >
            Drafts
          </Text>
          <Text
            type='title'
            extraClass='mb-0'
            level={8}
            color='character-secondary'
          >
            {queries.length}
          </Text>
        </div>
      </div>
      <div className='flex flex-col row-gap-5 px-4'>
        <SidebarSearch
          placeholder='Search board'
          setSearchText={setSearchText}
          searchText={searchText}
        />
        <div
          className={cx(
            'flex flex-col row-gap-1 overflow-auto',
            styles['dashboard-list-container']
          )}
        >
          {filteredDashboardList.map((dashboard) => (
            <DashboardItem dashboard={dashboard} key={dashboard.id} />
          ))}
        </div>
        <Button
          className={cx(
            'flex col-gap-2 items-center w-full',
            styles.sidebar_action_button
          )}
          onClick={() => {
            dispatch({ type: NEW_DASHBOARD_TEMPLATES_MODAL_OPEN });
          }}
          id='fa-at-btn--new-dashboard'
        >
          <SVG
            name='plus'
            size={16}
            extraClass={styles.sidebar_action_button__content}
            isFill={false}
          />
          <Text
            level={6}
            type='title'
            extraClass={cx('m-0', styles.sidebar_action_button__content)}
          >
            New Dashboard
          </Text>
        </Button>
      </div>
    </div>
  );
}

export default memo(DashboardSidebar);
