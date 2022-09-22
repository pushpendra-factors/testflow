import React, { useCallback, useRef, useMemo } from 'react';
import { ReactSortable } from 'react-sortablejs';
import WidgetCard from './WidgetCard';
import { useSelector, useDispatch } from 'react-redux';
import { UNITS_ORDER_CHANGED } from '../../reducers/types';
import { updateDashboard } from '../../reducers/dashboard/services';
import { getRequestForNewState } from '../../reducers/dashboard/utils';
import { QUERY_TYPE_WEB } from '../../utils/constants';
import WebsiteAnalytics from './WebsiteAnalytics';
import { Text } from '../../components/factorsComponents';

function SortableCards({
  setwidgetModal,
  durationObj,
  showDeleteWidgetModal,
  setOldestRefreshTime,
  dashboardRefreshState,
  onDataLoadSuccess
}) {
  const dispatch = useDispatch();
  const timerRef = useRef(null);

  const { active_project } = useSelector((state) => state.global);
  const { data: savedQueries } = useSelector((state) => state.queries);
  const { activeDashboardUnits, activeDashboard } = useSelector(
    (state) => state.dashboard
  );

  const NoDataDashboard = () => {
    return (
      <div
        className={
          'flex flex-col justify-center fa-dashboard--no-data-container items-center'
        }
      >
        <img
          alt="no-data"
          src="https://s3.amazonaws.com/www.factors.ai/assets/img/product/no-data.png"
          className={'mb-8'}
        />
        <Text type={'title'} level={5} weight={'bold'} extraClass={'m-0'}>
          Add widgets to start monitoring.
        </Text>
        <Text type={'title'} level={7} color={'grey'} extraClass={'m-0'}>
          You can select any of the saved reports and add them to dashboard as
          widgets to monitor your metrics.
        </Text>
      </div>
    );
  };
  const onDrop = useCallback(
    async (newState) => {
      const body = getRequestForNewState(newState);
      dispatch({
        type: UNITS_ORDER_CHANGED,
        payload: newState,
        units_position: body
      });
      clearTimeout(timerRef.current);
      timerRef.current = setTimeout(() => {
        updateDashboard(active_project.id, activeDashboard.id, {
          units_position: body
        });
      }, 300);
    },
    [activeDashboard?.id, active_project.id, dispatch]
  );

  const activeUnits = useMemo(
    () =>
      activeDashboardUnits.data.filter(
        (elem) =>
          savedQueries.findIndex(
            (sq) => sq.id === elem.query_id && sq.query.cl !== QUERY_TYPE_WEB
          ) > -1
      ),
    [activeDashboardUnits, savedQueries]
  );

  const webAnalyticsUnits = useMemo(
    () =>
      activeDashboardUnits.data
        .filter(
          (elem) =>
            savedQueries.findIndex(
              (sq) => sq.id === elem.query_id && sq.query.cl === QUERY_TYPE_WEB
            ) > -1
        )
        .map((elem) => {
          const query = savedQueries.find((sq) => sq.id === elem.query_id);
          return {
            ...elem,
            title: query.title,
            query: { ...query.query }
          };
        }),
    [activeDashboardUnits, savedQueries]
  );

  if (activeUnits.length) {
    return (
      <ReactSortable
        className="flex flex-wrap"
        list={activeUnits}
        setList={onDrop}
      >
        {activeUnits.map((item) => {
          const savedQuery = savedQueries.find((sq) => sq.id === item.query_id);

          return (
            <WidgetCard
              durationObj={durationObj}
              key={item.id}
              unit={{ ...item, query: savedQuery }}
              onDrop={onDrop}
              showDeleteWidgetModal={showDeleteWidgetModal}
              setOldestRefreshTime={setOldestRefreshTime}
              dashboardRefreshState={dashboardRefreshState}
              onDataLoadSuccess={onDataLoadSuccess}
            />
          );
        })}
      </ReactSortable>
    );
  } else if (webAnalyticsUnits.length) {
    return (
      <WebsiteAnalytics
        durationObj={durationObj}
        webAnalyticsUnits={webAnalyticsUnits}
        setwidgetModal={setwidgetModal}
        dashboardRefreshState={dashboardRefreshState}
      />
    );
  } else {
    return <NoDataDashboard />;
  }
}

export default SortableCards;
