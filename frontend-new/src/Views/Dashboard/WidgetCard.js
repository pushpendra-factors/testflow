import React, { useRef, useEffect, useCallback, useState } from 'react';
import _ from 'lodash';
import { connect, useDispatch, useSelector } from 'react-redux';
import { Button, Dropdown, Menu, Tooltip } from 'antd';
import { Text, SVG } from '../../components/factorsComponents';
import { RightOutlined, LeftOutlined } from '@ant-design/icons';
import CardContent from './CardContent';
import {
  initialState,
  formatApiData,
  calculateActiveUsersData,
  calculateFrequencyData,
  getStateQueryFromRequestQuery,
} from '../CoreQuery/utils';
import { cardClassNames } from '../../reducers/dashboard/utils';
import { getDataFromServer, getSavedAttributionMetrics } from './utils';
import {
  QUERY_TYPE_EVENT,
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_CAMPAIGN,
  QUERY_TYPE_WEB,
  ATTRIBUTION_METRICS,
} from '../../utils/constants';
import { DashboardContext } from '../../contexts/DashboardContext';
import { useHistory, useLocation } from 'react-router-dom';
import { shouldDataFetch } from '../../utils/dataFormatter';
import { fetchWeeklyIngishts } from '../../reducers/insights';
import styles from './index.module.scss';

function WidgetCard({
  unit,
  onDrop,
  showDeleteWidgetModal,
  durationObj,
  refreshClicked,
  setRefreshClicked,
  fetchWeeklyIngishts,
}) {
  const hasComponentUnmounted = useRef(false);
  const cardRef = useRef(null);
  const history = useHistory();
  const location = useLocation();
  const [resultState, setResultState] = useState(initialState);
  const { active_project } = useSelector((state) => state.global);
  const { activeDashboardUnits } = useSelector((state) => state.dashboard);
  const { metadata } = useSelector((state) => state.insights);
  const dispatch = useDispatch();
  const [attributionMetrics, setAttributionMetrics] = useState([
    ...ATTRIBUTION_METRICS,
  ]);

  useEffect(() => {
    if (
      location.state &&
      location.state.dashboardWidgetId &&
      location.state.dashboardWidgetId === unit.id
    ) {
      window.scrollTo({
        top: cardRef.current.getBoundingClientRect().top - 180,
        behavior: 'smooth',
      });
      location.state = undefined;
      window.history.replaceState(null, '');
    }
  }, [location.state, unit.id]);

  const getData = useCallback(
    async (refresh = false) => {
      try {
        hasComponentUnmounted.current = false;
        setResultState({
          ...initialState,
          loading: true,
        });

        let queryType,
          apiCallStatus = {
            required: true,
            message: null,
          };

        if (unit.query.query.query_group) {
          if (
            unit.query.query.cl &&
            unit.query.query.cl === QUERY_TYPE_CAMPAIGN
          ) {
            queryType = QUERY_TYPE_CAMPAIGN;
          } else {
            queryType = QUERY_TYPE_EVENT;
          }
        } else if (
          unit.query.query.cl &&
          unit.query.query.cl === QUERY_TYPE_ATTRIBUTION
        ) {
          apiCallStatus = shouldDataFetch(durationObj);
          queryType = QUERY_TYPE_ATTRIBUTION;
        } else if (
          unit.query.query.cl &&
          unit.query.query.cl === QUERY_TYPE_WEB
        ) {
          queryType = QUERY_TYPE_WEB;
        } else {
          queryType = QUERY_TYPE_FUNNEL;
        }

        if (apiCallStatus.required) {
          const res = await getDataFromServer(
            unit.query,
            unit.id,
            unit.dashboard_id,
            durationObj,
            refresh,
            active_project.id
          );
          if (
            queryType === QUERY_TYPE_FUNNEL &&
            !hasComponentUnmounted.current
          ) {
            setResultState({
              ...initialState,
              data: res.data.result,
            });
          } else if (
            queryType === QUERY_TYPE_ATTRIBUTION &&
            !hasComponentUnmounted.current
          ) {
            setResultState({
              ...initialState,
              data: res.data.result,
              apiCallStatus,
            });
          } else if (
            queryType === QUERY_TYPE_CAMPAIGN &&
            !hasComponentUnmounted.current
          ) {
            setResultState({
              ...initialState,
              data: res.data.result,
            });
          } else {
            if (!hasComponentUnmounted.current) {
              const result_group = res.data.result.result_group;
              const equivalentQuery = getStateQueryFromRequestQuery(
                unit.query.query.query_group[0]
              );
              const appliedBreakdown = [
                ...equivalentQuery.breakdown.event,
                ...equivalentQuery.breakdown.global,
              ];

              if (unit.query.query.query_group.length === 1) {
                setResultState({
                  ...initialState,
                  data: result_group[0],
                });
              } else if (unit.query.query.query_group.length === 3) {
                const userData = formatApiData(
                  result_group[0],
                  result_group[1]
                );
                const sessionsData = result_group[2];
                const activeUsersData = calculateActiveUsersData(
                  userData,
                  sessionsData,
                  appliedBreakdown
                );
                setResultState({
                  ...initialState,
                  data: activeUsersData,
                });
              } else if (unit.query.query.query_group.length === 4) {
                const eventsData = formatApiData(
                  result_group[0],
                  result_group[1]
                );
                const userData = formatApiData(
                  result_group[2],
                  result_group[3]
                );
                const frequencyData = calculateFrequencyData(
                  eventsData,
                  userData,
                  appliedBreakdown
                );
                setResultState({
                  ...initialState,
                  data: frequencyData,
                });
              } else {
                setResultState({
                  ...initialState,
                  data: formatApiData(result_group[0], result_group[1]),
                });
              }
            }
          }
        } else {
          setResultState({
            ...initialState,
            apiCallStatus,
          });
        }
        setRefreshClicked(false);
      } catch (err) {
        console.log(err);
        console.log(err.response);
        setRefreshClicked(false);
        setResultState({
          ...initialState,
          error: true,
        });
      }
    },
    [
      active_project.id,
      unit.query,
      unit.id,
      unit.dashboard_id,
      durationObj,
      setRefreshClicked,
    ]
  );

  useEffect(() => {
    getData();
    return () => {
      hasComponentUnmounted.current = true;
    };
  }, [getData, durationObj]);

  useEffect(() => {
    if (refreshClicked) {
      getData(true);
    }
  }, [refreshClicked, getData]);

  useEffect(() => {
    if (unit.settings && unit.settings.attributionMetrics) {
      setAttributionMetrics(
        getSavedAttributionMetrics(JSON.parse(unit.settings.attributionMetrics))
      );
    }
  }, [unit.settings]);

  const handleDelete = useCallback(() => {
    showDeleteWidgetModal(unit);
  }, [unit, showDeleteWidgetModal]);

  const getMenu = () => {
    return (
      <Menu>
        <Menu.Item key='0'>
          <a onClick={handleDelete} href='#!'>
            Delete Widget
          </a>
        </Menu.Item>
      </Menu>
    );
  };

  const changeCardSize = useCallback(
    (cardSize) => {
      const unitIndex = activeDashboardUnits.data.findIndex(
        (au) => au.id === unit.id
      );
      const updatedUnit = {
        ...unit,
        className: cardClassNames[cardSize],
        cardSize,
      };
      const newState = [
        ...activeDashboardUnits.data.slice(0, unitIndex),
        updatedUnit,
        ...activeDashboardUnits.data.slice(unitIndex + 1),
      ];
      onDrop(newState);
    },
    [unit, activeDashboardUnits.data, onDrop]
  );

  const handleEditQuery = useCallback(() => {
    // console.log('dashboard unit id-->>', unit);
    // console.log('metadata',metadata);
    // console.log('metadata',metadata.DashboardUnitWiseResult[unit.id]);

    if (metadata?.DashboardUnitWiseResult) {
      const insightsItem = metadata?.DashboardUnitWiseResult[unit.id];
      if (insightsItem) {
        dispatch({ type: 'SET_ACTIVE_INSIGHT', payload: { 
          id: unit?.id,
          isDashboard: true,
          ...insightsItem
        } 
      });
      } else {
        dispatch({ type: 'SET_ACTIVE_INSIGHT', payload: false });
      }

      if (insightsItem?.Enabled) {
        if (!_.isEmpty(insightsItem?.InsightsRange)) {
          let insightsLen =  Object.keys(insightsItem?.InsightsRange)?.length || 0; 
          fetchWeeklyIngishts(
            active_project.id,
            unit.id,
            Object.keys(insightsItem.InsightsRange)[insightsLen-1],
            insightsItem.InsightsRange[
              Object.keys(insightsItem.InsightsRange)[insightsLen-1]
            ][0]
          ).catch((e) => {
            console.log('weekly-ingishts fetch error', e);
          });
        } else {
          dispatch({ type: 'SET_ACTIVE_INSIGHT', payload: insightsItem });
        }
      } else {
        dispatch({ type: 'RESET_WEEKLY_INSIGHTS', payload: false });
      }
    }

    history.push({
      pathname: '/analyse',
      state: {
        query: { ...unit.query, settings: unit.settings },
        global_search: true,
        navigatedFromDashboard: unit,
      },
    });
  }, [
    history,
    unit,
    active_project.id,
    dispatch,
    fetchWeeklyIngishts,
    metadata?.DashboardUnitWiseResult,
  ]);

  return (
    <div
      className={`${unit.title.split(' ').join('-')} ${
        unit.className
      } py-3 flex widget-card-top-div`}
    >
      <div
        id={`card-${unit.id}`}
        ref={cardRef}
        className={'fa-dashboard--widget-card h-full w-full flex'}
      >
        <div className={'py-5 flex justify-between items-start w-full'}>
          <div className={'w-full flex flex-1 flex-col h-full'}>
            <div
              className={`${styles.widgetCard} flex items-center justify-between px-6 pb-4`}
            >
              <Tooltip title={unit.title} mouseEnterDelay={0.2}>
                <div className='flex flex-col truncate'>
                  <div
                    className='flex cursor-pointer items-center'
                    onClick={handleEditQuery}
                  >
                    <Text
                      ellipsis
                      type={'title'}
                      level={6}
                      weight={'bold'}
                      extraClass={`${styles.widgetCard_text} m-0 mr-1 flex`}
                    >
                      {unit.title} 
                      <SVG extraClass={`${styles.expand_icon} ml-1`} size={20} name="expand" />
                    </Text>
                  </div>
                  {/* <div className="description">
                  <Text
                    ellipsis
                    type={"paragraph"}
                    mini
                    color={"grey"}
                    extraClass={"m-0"}
                  >
                    {unit.description}
                  </Text>
                </div> */}
                </div>
              </Tooltip>
              <div className='flex items-center'>
                {resultState.apiCallStatus &&
                resultState.apiCallStatus.required &&
                resultState.apiCallStatus.message ? (
                  <Tooltip
                    mouseEnterDelay={0.2}
                    title={resultState.apiCallStatus.message}
                  >
                    <div className='cursor-pointer'>
                      <SVG color='#dea069' name={'warning'} />
                    </div>
                  </Tooltip>
                ) : null}
                <Dropdown overlay={getMenu()} trigger={['hover']}>
                  <Button
                    type='text'
                    icon={<SVG size={20} name={'threedot'} color='#8692A3' />}
                  />
                </Dropdown>
              </div>
            </div>
            <DashboardContext.Provider
              value={{
                attributionMetrics,
                setAttributionMetrics,
                handleEditQuery,
              }}
            >
              <CardContent
                durationObj={durationObj}
                unit={unit}
                resultState={resultState}
              />
            </DashboardContext.Provider>
          </div>
        </div>
      </div>
      <div
        id={`resize-${unit.id}`}
        className={'fa-widget-card--resize-container'}
      >
        <span className={'fa-widget-card--resize-contents'}>
          {unit.cardSize === 0 ? (
            <>
              <a href='#!' onClick={changeCardSize.bind(this, 1)}>
                <RightOutlined />
              </a>
              <a href='#!' onClick={changeCardSize.bind(this, 2)}>
                <LeftOutlined />
              </a>
            </>
          ) : null}
          {unit.cardSize === 1 ? (
            <a href='#!' onClick={changeCardSize.bind(this, 0)}>
              <LeftOutlined />
            </a>
          ) : null}
          {unit.cardSize === 2 ? (
            <a href='#!' onClick={changeCardSize.bind(this, 0)}>
              <RightOutlined />
            </a>
          ) : null}
        </span>
      </div>
    </div>
  );
}

export default connect(null, { fetchWeeklyIngishts })(React.memo(WidgetCard));
