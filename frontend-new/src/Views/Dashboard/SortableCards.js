import React, { useCallback, useRef, useMemo } from 'react';
import { ReactSortable } from 'react-sortablejs';
import { useSelector, useDispatch } from 'react-redux';
import WidgetCard from './WidgetCard';
import { UNITS_ORDER_CHANGED } from '../../reducers/types';
import { updateDashboard } from '../../reducers/dashboard/services';
import { getRequestForNewState } from '../../reducers/dashboard/utils';
import { QUERY_TYPE_WEB } from '../../utils/constants';
import WebsiteAnalytics from './WebsiteAnalytics';
import { Text } from '../../components/factorsComponents';
import SkeletonCard from 'Components/SkeletonCard';
import SkeletonGrid from 'Components/SkeletonCard/SkeletonGrid';
import HowToCreateNewReport from './../../assets/images/illustrations/HowToCreateNewReport.webm';
function NoDataDashboard() {
  return (
    <div
      style={{
        overflow: 'hidden',
        height: 'calc(100vh - 245px)',
        position: 'relative'
      }}
    >
      <SkeletonGrid />
      <div className='flex justify-center absolute top-0 w-full h-full text-center'>
        <div>
          <video
            autoPlay
            style={{
              width: '300px',
              objectFit: 'cover',
              clipPath: 'inset(1px 1px)',
              borderRadius: '8px',
              margin: '40px auto 0 auto'
            }}
            loop
          >
            <source src={HowToCreateNewReport} type='video/mp4' />
          </video>
          <Text type='title' level={5} weight='bold' extraClass='m-0'>
            What kind of reports do you want to store in this dashboard?
          </Text>
          <Text type='title' level={7} color='grey' extraClass='m-0'>
            You can create a new report or pick from one of your draft reports
          </Text>
        </div>
      </div>
      {/* <img
        alt='no-data'
        src='https://s3.amazonaws.com/www.factors.ai/assets/img/product/no-data.png'
        className='mb-8'
      />
      <Text type='title' level={5} weight='bold' extraClass='m-0'>
        Add widgets to start monitoring. sgd
      </Text>
      <Text type='title' level={7} color='grey' extraClass='m-0'>
        You can select any of the saved reports and add them to dashboard as
        widgets to monitor your metrics.
      </Text> */}
    </div>
  );
}

function SortableCards({
  setwidgetModal,
  durationObj,
  showDeleteWidgetModal,
  setOldestRefreshTime,
  dashboardRefreshState,
  onDataLoadSuccess,
  handleWidgetRefresh,
  resetDashboardRefreshState
}) {
  const dispatch = useDispatch();
  const timerRef = useRef(null);

  const { active_project: activeProject } = useSelector(
    (state) => state.global
  );
  const { data: savedQueries } = useSelector((state) => state.queries);
  const { activeDashboardUnits, activeDashboard } = useSelector(
    (state) => state.dashboard
  );

  const onDrop = useCallback(
    async (newState) => {
      const tmpState = newState.filter((e) => e.id !== 'addnewreport');
      const body = getRequestForNewState(tmpState);
      dispatch({
        type: UNITS_ORDER_CHANGED,
        payload: tmpState,
        units_position: body
      });
      clearTimeout(timerRef.current);
      timerRef.current = setTimeout(() => {
        updateDashboard(activeProject.id, activeDashboard.id, {
          units_position: body
        });
      }, 300);
    },
    [activeDashboard?.id, activeProject.id, dispatch]
  );

  const activeUnits = useMemo(() => {
    return activeDashboardUnits.data.filter(
      (elem) =>
        savedQueries.findIndex(
          (sq) => sq.id === elem.query_id && sq.query.cl !== QUERY_TYPE_WEB
        ) > -1
    );
  }, [activeDashboardUnits, savedQueries]);

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
    let activeUnitsWithAddNewReport = [
      ...activeUnits,
      {
        id: 'addnewreport',
        addNewReport: true,
        durationObj: { from: 0, to: 0 },
        unit: {
          query: {
            query: { ewp: [] },
            title: 'addnewreport',
            id: 'addnewreport'
          },
          query_id: '',
          cardSize: 0,
          className: 'w-1/2'
        },
        dashboardRefreshState: {},
        onDataLoadSuccess: () => {}
      }
    ];
    return (
      <ReactSortable
        className='flex flex-wrap'
        list={activeUnitsWithAddNewReport}
        setList={onDrop}
      >
        {activeUnitsWithAddNewReport.map((item) => {
          const savedQuery = savedQueries.find((sq) => sq.id === item.query_id);
          if (item.addNewReport) {
            // This is only got addNewReport Widget
            return <WidgetCard key={item.id} {...item} />;
          }
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
              handleWidgetRefresh={handleWidgetRefresh}
            />
          );
        })}
      </ReactSortable>
    );
  }
  if (webAnalyticsUnits.length) {
    return (
      <WebsiteAnalytics
        durationObj={durationObj}
        webAnalyticsUnits={webAnalyticsUnits}
        setwidgetModal={setwidgetModal}
        dashboardRefreshState={dashboardRefreshState}
        resetDashboardRefreshState={resetDashboardRefreshState}
      />
    );
  }
  return <NoDataDashboard />;
}

export default SortableCards;
