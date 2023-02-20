import React, {
  useRef,
  useEffect,
  useCallback,
  useState,
  useMemo
} from 'react';
import _ from 'lodash';
import { connect, useDispatch, useSelector } from 'react-redux';
import { Button, Dropdown, Menu, Tooltip } from 'antd';
import { RightOutlined, LeftOutlined } from '@ant-design/icons';
import { useHistory, useLocation } from 'react-router-dom';
import { Text, SVG } from '../../components/factorsComponents';
import CardContent from './CardContent';
import {
  initialState,
  formatApiData,
  calculateActiveUsersData,
  calculateFrequencyData,
  getStateQueryFromRequestQuery
} from '../CoreQuery/utils';
import { cardClassNames } from '../../reducers/dashboard/utils';
import {
  getDataFromServer,
  getSavedAttributionMetrics,
  getValidGranularityForSavedQueryWithSavedGranularity
} from './utils';
import {
  QUERY_TYPE_EVENT,
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_CAMPAIGN,
  QUERY_TYPE_WEB,
  ATTRIBUTION_METRICS,
  QUERY_TYPE_PROFILE,
  QUERY_TYPE_KPI
} from '../../utils/constants';
import { DashboardContext } from '../../contexts/DashboardContext';
import { shouldDataFetch } from '../../utils/dataFormatter';
import { fetchWeeklyIngishts as fetchWeeklyInsightsAction } from '../../reducers/insights';
import styles from './index.module.scss';

function WidgetCard({
  unit,
  onDrop,
  showDeleteWidgetModal,
  durationObj,
  fetchWeeklyInsights,
  setOldestRefreshTime,
  dashboardRefreshState,
  onDataLoadSuccess,
  handleWidgetRefresh
}) {
  const hasComponentUnmounted = useRef(false);
  const cardRef = useRef(null);
  const history = useHistory();
  const location = useLocation();
  const [resultState, setResultState] = useState(initialState);
  const { active_project: activeProject } = useSelector(
    (state) => state.global
  );
  const { activeDashboardUnits } = useSelector((state) => state.dashboard);
  const { metadata } = useSelector((state) => state.insights);
  const { data: savedQueries } = useSelector((state) => state.queries);
  const dispatch = useDispatch();
  const [attributionMetrics, setAttributionMetrics] = useState([
    ...ATTRIBUTION_METRICS
  ]);
  const [tableFilters, setTableFilters] = useState({});

  const savedQuery = useMemo(
    () => _.find(savedQueries, (sq) => sq.id === unit.query_id),
    [savedQueries, unit?.query_id]
  );

  const durationWithSavedFrequency = useMemo(() => {
    if (_.get(savedQuery, 'query.query_group', null)) {
      const savedFrequency = _.get(
        savedQuery,
        'query.query_group.0.gbt',
        'date'
      );
      const frequency = getValidGranularityForSavedQueryWithSavedGranularity({
        durationObj,
        savedFrequency
      });
      return {
        ...durationObj,
        frequency
      };
    }
    if (_.get(savedQuery, 'query.cl', null) === QUERY_TYPE_KPI) {
      const savedFrequency = _.get(savedQuery, 'query.qG.1.gbt', 'date');
      const frequency = getValidGranularityForSavedQueryWithSavedGranularity({
        durationObj,
        savedFrequency
      });
      return {
        ...durationObj,
        frequency
      };
    }
    return durationObj;
  }, [durationObj, savedQuery]);

  useEffect(() => {
    if (
      location.state &&
      location.state.dashboardWidgetId &&
      location.state.dashboardWidgetId === unit.id
    ) {
      window.scrollTo({
        top: cardRef.current.getBoundingClientRect().top - 180,
        behavior: 'smooth'
      });
      location.state = undefined;
      window.history.replaceState(null, '');
    }
  }, [location, unit.id]);

  const getData = useCallback(
    async (refresh = false) => {
      try {
        hasComponentUnmounted.current = false;
        setResultState({
          ...initialState,
          loading: true
        });

        let queryType;
        let apiCallStatus = {
          required: true,
          message: null
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
        } else if (unit.query.query.cl === QUERY_TYPE_KPI) {
          queryType = QUERY_TYPE_KPI;
        } else if (
          unit.query.query.cl &&
          unit.query.query.cl === QUERY_TYPE_ATTRIBUTION
        ) {
          apiCallStatus = shouldDataFetch(durationWithSavedFrequency);
          queryType = QUERY_TYPE_ATTRIBUTION;
        } else if (
          unit.query.query.cl &&
          unit.query.query.cl === QUERY_TYPE_WEB
        ) {
          queryType = QUERY_TYPE_WEB;
        } else if (
          unit.query.query.cl &&
          unit.query.query.cl === QUERY_TYPE_PROFILE
        ) {
          queryType = QUERY_TYPE_PROFILE;
        } else {
          queryType = QUERY_TYPE_FUNNEL;
        }

        let lastRefreshedAt = null;
        if (apiCallStatus.required) {
          const res = await getDataFromServer(
            unit.query,
            unit.id,
            unit.dashboard_id,
            durationWithSavedFrequency,
            refresh,
            activeProject.id
          );
          if (!hasComponentUnmounted.current) {
            onDataLoadSuccess({ unitId: unit.id });
          }
          if (
            queryType === QUERY_TYPE_FUNNEL &&
            !hasComponentUnmounted.current
          ) {
            lastRefreshedAt = _.get(
              res,
              'data.cache_meta.last_computed_at',
              null
            );
            setResultState({
              ...initialState,
              data: res.data.result
            });
          } else if (
            queryType === QUERY_TYPE_PROFILE &&
            !hasComponentUnmounted.current
          ) {
            lastRefreshedAt = _.get(
              res,
              'data.cache_meta.last_computed_at',
              null
            );
            setResultState({
              ...initialState,
              data: res.data.result
            });
          } else if (
            queryType === QUERY_TYPE_ATTRIBUTION &&
            !hasComponentUnmounted.current
          ) {
            lastRefreshedAt = _.get(
              res,
              'data.cache_meta.last_computed_at',
              null
            );
            setResultState({
              ...initialState,
              data: res.data.result,
              apiCallStatus
            });
          } else if (
            queryType === QUERY_TYPE_CAMPAIGN &&
            !hasComponentUnmounted.current
          ) {
            lastRefreshedAt = _.get(
              res,
              'data.cache_meta.last_computed_at',
              null
            );
            setResultState({
              ...initialState,
              data: res.data.result
            });
          } else if (
            queryType === QUERY_TYPE_KPI &&
            !hasComponentUnmounted.current
          ) {
            lastRefreshedAt = _.get(
              res,
              'data.cache_meta.last_computed_at',
              null
            );
            setResultState({
              ...initialState,
              data: res.data.result || res.data
            });
          } else if (!hasComponentUnmounted.current) {
            lastRefreshedAt = _.get(
              res,
              'data.cache_meta.last_computed_at',
              null
            );
            const resultGroup = res.data.result.result_group;
            const equivalentQuery = getStateQueryFromRequestQuery(
              unit.query.query.query_group[0]
            );
            const appliedBreakdown = [
              ...equivalentQuery.breakdown.event,
              ...equivalentQuery.breakdown.global
            ];

            if (unit.query.query.query_group.length === 1) {
              setResultState({
                ...initialState,
                data: resultGroup[0]
              });
            } else if (unit.query.query.query_group.length === 3) {
              const userData = formatApiData(resultGroup[0], resultGroup[1]);
              const sessionsData = resultGroup[2];
              const activeUsersData = calculateActiveUsersData(
                userData,
                sessionsData,
                appliedBreakdown
              );
              setResultState({
                ...initialState,
                data: activeUsersData
              });
            } else if (unit.query.query.query_group.length === 4) {
              const eventsData = formatApiData(resultGroup[0], resultGroup[1]);
              const userData = formatApiData(resultGroup[2], resultGroup[3]);
              const frequencyData = calculateFrequencyData(
                eventsData,
                userData,
                appliedBreakdown
              );
              setResultState({
                ...initialState,
                data: frequencyData
              });
            } else {
              setResultState({
                ...initialState,
                data: formatApiData(resultGroup[0], resultGroup[1])
              });
            }
          }
          if (lastRefreshedAt != null && !hasComponentUnmounted.current) {
            setOldestRefreshTime((currValue) => {
              if (currValue == null || lastRefreshedAt < currValue) {
                return lastRefreshedAt;
              }
              return currValue;
            });
          }
        } else {
          setResultState({
            ...initialState,
            apiCallStatus
          });
        }
      } catch (err) {
        console.log(err);
        console.log(err.response);
        if (!hasComponentUnmounted.current) {
          onDataLoadSuccess({ unitId: unit.id });
        }
        setResultState({
          ...initialState,
          error: true
        });
      }
    },
    [
      unit.query,
      unit.id,
      unit.dashboard_id,
      durationWithSavedFrequency,
      activeProject.id,
      onDataLoadSuccess,
      setOldestRefreshTime
    ]
  );

  useEffect(() => {
    getData();
    return () => {
      hasComponentUnmounted.current = true;
    };
  }, [getData, durationWithSavedFrequency]);

  useEffect(() => {
    if (dashboardRefreshState.widgetIdGettingFetched === unit.id) {
      getData(true);
    }
  }, [dashboardRefreshState.widgetIdGettingFetched, unit.id, getData]);

  useEffect(() => {
    if (unit?.query?.settings != null) {
      if (unit.query.settings.attributionMetrics != null) {
        setAttributionMetrics(
          getSavedAttributionMetrics(
            JSON.parse(unit.query.settings.attributionMetrics)
          )
        );
      }
      if (unit.query.settings.tableFilters != null) {
        setTableFilters(JSON.parse(unit.query.settings.tableFilters));
      }
    }
  }, [unit?.query?.settings]);

  const handleDelete = useCallback(() => {
    showDeleteWidgetModal(unit);
  }, [unit, showDeleteWidgetModal]);

  const onWidgetRefresh = useCallback(() => {
    handleWidgetRefresh(unit.id);
  }, [unit.id, handleWidgetRefresh]);

  const getMenu = () => (
    <Menu>
      <Menu.Item key='0'>
        <a onClick={handleDelete} href='#!'>
          Delete Widget
        </a>
      </Menu.Item>
      <Menu.Item key='1'>
        <a onClick={onWidgetRefresh} href='#!'>
          Refresh
        </a>
      </Menu.Item>
    </Menu>
  );

  const changeCardSize = useCallback(
    (cardSize) => {
      const unitIndex = activeDashboardUnits.data.findIndex(
        (au) => au.id === unit.id
      );
      const updatedUnit = {
        ...unit,
        className: cardClassNames[cardSize],
        cardSize
      };
      const newState = [
        ...activeDashboardUnits.data.slice(0, unitIndex),
        updatedUnit,
        ...activeDashboardUnits.data.slice(unitIndex + 1)
      ];
      onDrop(newState);
    },
    [unit, activeDashboardUnits.data, onDrop]
  );

  const handleEditQuery = useCallback(() => {
    if (metadata?.DashboardUnitWiseResult) {
      const insightsItem = metadata?.DashboardUnitWiseResult[unit.id];
      if (insightsItem) {
        dispatch({
          type: 'SET_ACTIVE_INSIGHT',
          payload: {
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
          const insightsLen =
            Object.keys(insightsItem?.InsightsRange)?.length || 0;
          fetchWeeklyInsights(
            activeProject.id,
            unit.id,
            Object.keys(insightsItem.InsightsRange)[insightsLen - 1],
            insightsItem.InsightsRange[
              Object.keys(insightsItem.InsightsRange)[insightsLen - 1]
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
        query: { ...unit.query, settings: unit.query.settings },
        global_search: true,
        navigatedFromDashboard: unit
      }
    });
  }, [
    history,
    unit,
    activeProject.id,
    dispatch,
    fetchWeeklyInsights,
    metadata?.DashboardUnitWiseResult
  ]);

  const contextValue = useMemo(
    () => ({
      attributionMetrics,
      tableFilters,
      setAttributionMetrics,
      handleEditQuery
    }),
    [attributionMetrics, tableFilters, handleEditQuery]
  );

  return (
    <div
      className={`${unit?.query?.title.split(' ').join('-')} ${
        unit.className
      } py-3 flex widget-card-top-div`}
    >
      <div
        id={`card-${unit.id}`}
        ref={cardRef}
        className={`fa-dashboard--widget-card h-full w-full flex ${styles.widgetCardCustomCSS}`}
      >
        <div className='flex justify-between items-start w-full'>
          <div className='w-full flex flex-1 flex-col h-full justify-between'>
            <div
              className={`${styles.widgetCard} flex items-center justify-between px-4`}
            >
              <div
                className='widget-card--title-container py-3 flex truncate cursor-pointer items-center w-full mr-2'
                onClick={handleEditQuery}
              >
                <div className='flex  items-center'>
                  <Tooltip title={unit?.query?.title} mouseEnterDelay={0.2}>
                    <Text
                      ellipsis
                      type='title'
                      level={6}
                      weight='bold'
                      extraClass='widget-card--title m-0 mr-1 flex'
                    >
                      {unit?.query?.title}
                    </Text>
                  </Tooltip>
                </div>
                <SVG
                  extraClass='widget-card--expand-icon ml-1'
                  size={20}
                  color='grey'
                  name='arrowright'
                />
              </div>
              <div className='flex items-center'>
                {resultState.apiCallStatus &&
                resultState.apiCallStatus.required &&
                resultState.apiCallStatus.message ? (
                  <Tooltip
                    mouseEnterDelay={0.2}
                    title={resultState.apiCallStatus.message}
                  >
                    <div className='cursor-pointer'>
                      <SVG color='#dea069' name='warning' />
                    </div>
                  </Tooltip>
                ) : null}
                <Dropdown
                  placement='bottomRight'
                  overlay={getMenu()}
                  trigger={['hover']}
                >
                  <Button
                    type='text'
                    icon={<SVG size={20} name='threedot' color='grey' />}
                  />
                </Dropdown>
              </div>
            </div>
            <DashboardContext.Provider value={contextValue}>
              <CardContent
                durationObj={durationWithSavedFrequency}
                unit={unit}
                resultState={resultState}
              />
            </DashboardContext.Provider>
          </div>
        </div>
      </div>
      <div
        id={`resize-${unit.id}`}
        className='fa-widget-card--resize-container'
      >
        <span className='fa-widget-card--resize-contents'>
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

export default connect(null, {
  fetchWeeklyInsights: fetchWeeklyInsightsAction
})(React.memo(WidgetCard));
