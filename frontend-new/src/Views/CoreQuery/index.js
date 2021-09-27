import React, {
  useState,
  useCallback,
  useEffect,
  useMemo,
  useReducer,
} from 'react';
import MomentTz from 'Components/MomentTz';
import { bindActionCreators } from 'redux';
import { connect, useSelector, useDispatch } from 'react-redux';
import QueryComposer from '../../components/QueryComposer';
import AttrQueryComposer from '../../components/AttrQueryComposer';
import CampQueryComposer from '../../components/CampQueryComposer';
import CoreQueryHome from '../CoreQueryHome';
import { Drawer, Button, Collapse, Modal } from 'antd';
import {
  Text,
  SVG,
  FaErrorComp,
  FaErrorLog,
} from '../../components/factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import {
  deleteGroupByForEvent,
  getCampaignConfigData,
} from '../../reducers/coreQuery/middleware';
import {
  calculateFrequencyData,
  calculateActiveUsersData,
  formatApiData,
  getQuery,
  initialState,
  getFunnelQuery,
  DefaultDateRangeFormat,
  getAttributionQuery,
  getCampaignsQuery,
  isComparisonEnabled,
} from './utils';
import {
  getEventsData,
  getFunnelData,
  getAttributionsData,
  getCampaignsData,
} from '../../reducers/coreQuery/services';
import {
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_CAMPAIGN,
  QUERY_TYPE_ATTRIBUTION,
  TOTAL_EVENTS_CRITERIA,
  TOTAL_USERS_CRITERIA,
  ACTIVE_USERS_CRITERIA,
  FREQUENCY_CRITERIA,
  EACH_USER_TYPE,
  REPORT_SECTION,
  INITIAL_SESSION_ANALYTICS_SEQ,
  ATTRIBUTION_METRICS,
} from '../../utils/constants';
import { SHOW_ANALYTICS_RESULT } from '../../reducers/types';
import AnalysisResultsPage from './AnalysisResultsPage';
import AnalysisHeader from './AnalysisResultsPage/AnalysisHeader';
import {
  SET_CAMP_DATE_RANGE,
  SET_ATTR_DATE_RANGE,
} from '../../reducers/coreQuery/actions';
import { CoreQueryContext } from '../../contexts/CoreQueryContext';
import CoreQueryReducer from './CoreQueryReducer';
import {
  CORE_QUERY_INITIAL_STATE,
  SET_COMPARISON_ENABLED,
  COMPARISON_DATA_LOADING,
  COMPARISON_DATA_FETCHED,
  RESET_COMPARISON_DATA,
  SET_COMPARISON_SUPPORTED,
  SET_COMPARE_DURATION,
  SET_NAVIGATED_FROM_DASHBOARD,
  UPDATE_CHART_TYPES,
  SET_SAVED_QUERY_SETTINGS,
} from './constants';
import {
  getValidGranularityOptions,
  shouldDataFetch,
} from '../../utils/dataFormatter';

const { Panel } = Collapse;

function CoreQuery({
  activeProject,
  deleteGroupByForEvent,
  location,
  getCampaignConfigData,
}) {
  const [coreQueryState, localDispatch] = useReducer(
    CoreQueryReducer,
    CORE_QUERY_INITIAL_STATE
  );
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [queryType, setQueryType] = useState(QUERY_TYPE_EVENT);
  const [activeKey, setActiveKey] = useState('0');
  const [showResult, setShowResult] = useState(false);
  const [appliedQueries, setAppliedQueries] = useState([]);
  const [appliedBreakdown, setAppliedBreakdown] = useState([]);
  const [resultState, setResultState] = useState(initialState);
  const [requestQuery, updateRequestQuery] = useState(null);
  const [rowClicked, setRowClicked] = useState(false);
  const [querySaved, setQuerySaved] = useState(false);
  const [breakdownType, setBreakdownType] = useState(EACH_USER_TYPE);
  const [queries, setQueries] = useState([]);
  const [queryOptions, setQueryOptions] = useState({
    groupBy: [
      {
        prop_category: '', // user / event
        property: '', // user/eventproperty
        prop_type: '', // categorical  /numberical
        eventValue: '', // event name (funnel only)
        eventName: '', // eventName $present for global user breakdown
        eventIndex: 0,
      },
    ],
    globalFilters: [],
    event_analysis_seq: '',
    session_analytics_seq: INITIAL_SESSION_ANALYTICS_SEQ,
    date_range: { ...DefaultDateRangeFormat },
  });
  const [attributionsState, setAttributionsState] = useState({
    eventGoal: {},
    touchpoint: '',
    models: [],
    linkedEvents: [],
    date_range: {},
    attr_dimensions: [],
  });

  const [campaignState, setCampaignState] = useState({
    channel: '',
    select_metrics: [],
    filters: [],
    group_by: [],
    date_range: {},
  });

  const [attributionMetrics, setAttributionMetrics] = useState([
    ...ATTRIBUTION_METRICS,
  ]);

  const dispatch = useDispatch();
  const {
    groupBy,
    eventGoal,
    touchpoint,
    touchpoint_filters,
    attr_query_type,
    models,
    window,
    linkedEvents,
    camp_channels,
    camp_measures,
    camp_filters,
    camp_groupBy,
    camp_dateRange,
    attr_dateRange,
    eventNames,
    attr_dimensions,
  } = useSelector((state) => state.coreQuery);

  const [activeTab, setActiveTab] = useState(1);

  const [queryOpen, setQueryOpen] = useState(true);

  const {
    show_criteria: result_criteria,
    performance_criteria: user_type,
  } = useSelector((state) => state.analyticsQuery);

  const dateRange = queryOptions.date_range;
  const { session_analytics_seq } = queryOptions;
  const { globalFilters } = queryOptions;

  useEffect(() => {
    if (activeProject && activeProject.id) {
      getCampaignConfigData(activeProject.id, 'all_ads');
    }
  }, [activeProject, getCampaignConfigData]);

  const updateResultState = useCallback((newState) => {
    setResultState(newState);
  }, []);

  const updateAppliedBreakdown = useCallback(() => {
    const newAppliedBreakdown = [...groupBy.event, ...groupBy.global];
    setAppliedBreakdown(newAppliedBreakdown);
  }, [groupBy]);

  const updateLocalReducer = useCallback((type, payload) => {
    localDispatch({ type, payload });
  }, []);

  const updateChartTypes = useCallback(
    (payload) => {
      updateLocalReducer(UPDATE_CHART_TYPES, payload);
    },
    [updateLocalReducer]
  );

  const resetComparisonData = useCallback(() => {
    updateLocalReducer(RESET_COMPARISON_DATA);
  }, [updateLocalReducer]);

  const handleCompareWithClick = useCallback(() => {
    updateLocalReducer(SET_COMPARISON_ENABLED, true);
  }, [updateLocalReducer]);

  const setNavigatedFromDashboard = useCallback(
    (payload) => {
      updateLocalReducer(SET_NAVIGATED_FROM_DASHBOARD, payload);
    },
    [updateLocalReducer]
  );

  const updateSavedQuerySettings = useCallback(
    (payload) => {
      updateLocalReducer(SET_SAVED_QUERY_SETTINGS, payload);
    },
    [updateLocalReducer]
  );

  const configActionsOnRunningQuery = useCallback(
    (isQuerySaved) => {
      closeDrawer();
      dispatch({ type: SHOW_ANALYTICS_RESULT, payload: true });
      setShowResult(true);
      setQuerySaved(isQuerySaved);
      if (!isQuerySaved) {
        setNavigatedFromDashboard(false);
        updateSavedQuerySettings({});
      }
      localDispatch({
        type: SET_COMPARISON_SUPPORTED,
        payload: isComparisonEnabled(queryType, queries, groupBy, models),
      });
      if (queryType === QUERY_TYPE_FUNNEL || queryType === QUERY_TYPE_EVENT) {
        setAppliedQueries(queries.map((elem) => elem.alias? elem.alias : elem.label));
        updateAppliedBreakdown();
      }
    },
    [
      dispatch,
      groupBy,
      queries,
      queryType,
      models,
      updateAppliedBreakdown,
      setNavigatedFromDashboard,
      updateSavedQuerySettings,
    ]
  );

  const getDashboardConfigs = useCallback(
    (isQuerySaved) => {
      // use cache urls when user expands the dashboard widget
      if (isQuerySaved && coreQueryState.navigatedFromDashboard) {
        return {
          id: coreQueryState.navigatedFromDashboard.dashboard_id,
          unit_id: coreQueryState.navigatedFromDashboard.id,
          refresh: false,
        };
      }
      return null;
    },
    [coreQueryState.navigatedFromDashboard]
  );

  const runQuery = useCallback(
    async (
      isQuerySaved,
      durationObj,
      isGranularityChange = false,
      isCompareQuery = false
    ) => {
      try {
        if (!durationObj) {
          durationObj = dateRange;
        }
        const query = getQuery(
          groupBy,
          queries,
          result_criteria,
          user_type,
          durationObj,
          globalFilters
        );
        if (!isCompareQuery) {
          configActionsOnRunningQuery(isQuerySaved);
          setBreakdownType(user_type);
          updateRequestQuery(query);
          updateResultState({ ...initialState, loading: true });
        } else {
          updateLocalReducer(COMPARISON_DATA_LOADING);
        }
        const res = await getEventsData(
          activeProject.id,
          query,
          getDashboardConfigs(isGranularityChange ? false : isQuerySaved) //we need to call fresh query when granularity is changed
        );
        const data = res.data.result || res.data;
        if (result_criteria === TOTAL_EVENTS_CRITERIA) {
          updateResultState({
            ...initialState,
            data: formatApiData(data.result_group[0], data.result_group[1]),
          });
        } else if (result_criteria === TOTAL_USERS_CRITERIA) {
          if (user_type === EACH_USER_TYPE) {
            updateResultState({
              ...initialState,
              data: formatApiData(data.result_group[0], data.result_group[1]),
            });
          } else {
            updateResultState({
              ...initialState,
              data: data.result_group[0],
            });
          }
        } else if (result_criteria === ACTIVE_USERS_CRITERIA) {
          const userData = formatApiData(
            data.result_group[0],
            data.result_group[1]
          );
          const sessionsData = data.result_group[2];
          const activeUsersData = calculateActiveUsersData(
            userData,
            sessionsData,
            [...groupBy.global, ...groupBy.event]
          );
          updateResultState({ ...initialState, data: activeUsersData });
        } else if (result_criteria === FREQUENCY_CRITERIA) {
          const eventData = formatApiData(
            data.result_group[0],
            data.result_group[1]
          );
          const userData = formatApiData(
            data.result_group[2],
            data.result_group[3]
          );
          const frequencyData = calculateFrequencyData(eventData, userData, [
            ...groupBy.global,
            ...groupBy.event,
          ]);
          updateResultState({ ...initialState, data: frequencyData });
        }
      } catch (err) {
        console.log(err);
        updateResultState({ ...initialState, error: true });
      }
    },
    [
      queries,
      dateRange,
      result_criteria,
      user_type,
      activeProject.id,
      groupBy,
      globalFilters,
      updateResultState,
      configActionsOnRunningQuery,
      updateLocalReducer,
      getDashboardConfigs,
    ]
  );

  const runFunnelQuery = useCallback(
    async (isQuerySaved, durationObj, isCompareQuery) => {
      try {
        if (!durationObj) {
          durationObj = dateRange;
          resetComparisonData();
        }
        const query = getFunnelQuery(
          groupBy,
          queries,
          session_analytics_seq,
          durationObj,
          globalFilters
        );
        if (!isCompareQuery) {
          configActionsOnRunningQuery(isQuerySaved);
          updateResultState({ ...initialState, loading: true });
          updateRequestQuery(query);
        } else {
          updateLocalReducer(COMPARISON_DATA_LOADING);
        }
        const res = await getFunnelData(
          activeProject.id,
          query,
          getDashboardConfigs(isQuerySaved)
        );
        if (isCompareQuery) {
          updateLocalReducer(
            COMPARISON_DATA_FETCHED,
            res.data.result || res.data
          );
        } else {
          updateResultState({
            ...initialState,
            data: res.data.result || res.data,
          });
        }
      } catch (err) {
        console.log(err);
        updateResultState({ ...initialState, error: true });
      }
    },
    [
      queries,
      session_analytics_seq,
      activeProject.id,
      groupBy,
      globalFilters,
      dateRange,
      updateResultState,
      configActionsOnRunningQuery,
      updateLocalReducer,
      resetComparisonData,
      getDashboardConfigs,
    ]
  );

  const runAttributionQuery = useCallback(
    async (isQuerySaved, durationObj, isCompareQuery) => {
      try {
        if (!durationObj) {
          durationObj = attr_dateRange;
          resetComparisonData();
        }
        const query = getAttributionQuery(
          eventGoal,
          touchpoint,
          attr_dimensions,
          touchpoint_filters,
          attr_query_type,
          models,
          window,
          linkedEvents,
          durationObj
        );
        if (!isCompareQuery) {
          configActionsOnRunningQuery(isQuerySaved);
          updateResultState({ ...initialState, loading: true });
          updateRequestQuery(query);
          setAttributionsState({
            eventGoal,
            touchpoint,
            models,
            linkedEvents,
            attr_dimensions,
            date_range: { ...durationObj },
          });
        } else {
          updateLocalReducer(COMPARISON_DATA_LOADING);
        }
        let apiCallStatus = { required: true, message: null };
        if (
          isQuerySaved &&
          coreQueryState.navigatedFromDashboard &&
          !isCompareQuery
        ) {
          apiCallStatus = shouldDataFetch(durationObj);
        }
        if (apiCallStatus.required) {
          const res = await getAttributionsData(
            activeProject.id,
            query,
            getDashboardConfigs(isQuerySaved)
          );
          if (isCompareQuery) {
            updateLocalReducer(
              COMPARISON_DATA_FETCHED,
              res.data.result || res.data
            );
          } else {
            updateResultState({
              ...initialState,
              data: res.data.result || res.data,
              apiCallStatus,
            });
          }
        } else {
          updateResultState({
            ...initialState,
            apiCallStatus,
          });
        }
      } catch (err) {
        console.log(err);
        updateResultState({
          ...initialState,
          error: true,
        });
      }
    },
    [
      activeProject.id,
      eventGoal,
      linkedEvents,
      models,
      touchpoint,
      touchpoint_filters,
      attr_query_type,
      window,
      attr_dateRange,
      updateResultState,
      attr_dimensions,
      getDashboardConfigs,
      configActionsOnRunningQuery,
      resetComparisonData,
      updateLocalReducer,
      coreQueryState.navigatedFromDashboard,
    ]
  );

  const runCampaignsQuery = useCallback(
    async (isQuerySaved, durationObj = null, isGranularityChange = false) => {
      try {
        closeDrawer();
        dispatch({ type: SHOW_ANALYTICS_RESULT, payload: true });
        setShowResult(true);
        setQuerySaved(isQuerySaved);
        if (!isQuerySaved) {
          setNavigatedFromDashboard(false);
        }
        updateResultState({
          ...initialState,
          loading: true,
        });
        if (!durationObj) {
          durationObj = camp_dateRange;
        }
        const query = getCampaignsQuery(
          camp_channels,
          camp_measures,
          camp_filters,
          camp_groupBy,
          durationObj
        );
        setCampaignState({
          channel: query.query_group[0].channel,
          filters: query.query_group[0].filters,
          select_metrics: query.query_group[0].select_metrics,
          group_by: query.query_group[0].group_by,
          date_range: { ...durationObj },
        });
        updateRequestQuery(query);
        const res = await getCampaignsData(
          activeProject.id,
          query,
          getDashboardConfigs(isGranularityChange ? false : isQuerySaved) //we need to call fresh query when granularity is changed
        );
        updateResultState({
          ...initialState,
          data: res.data.result || res.data,
        });
      } catch (err) {
        console.log(err);
        updateResultState({
          ...initialState,
          error: true,
        });
      }
    },
    [
      dispatch,
      activeProject.id,
      camp_measures,
      camp_filters,
      camp_groupBy,
      camp_channels,
      camp_dateRange,
      updateResultState,
      setNavigatedFromDashboard,
      getDashboardConfigs,
    ]
  );

  const handleGranularityChange = useCallback(
    ({ key: frequency }) => {
      if (queryType === QUERY_TYPE_EVENT) {
        const appliedDateRange = {
          ...queryOptions.date_range,
          frequency,
        };
        setQueryOptions((currState) => {
          return {
            ...currState,
            date_range: appliedDateRange,
          };
        });
        runQuery(querySaved, appliedDateRange, true);
      }
      if (queryType === QUERY_TYPE_CAMPAIGN) {
        const payload = {
          ...camp_dateRange,
          frequency,
        };
        dispatch({ type: SET_CAMP_DATE_RANGE, payload });
        runCampaignsQuery(querySaved, payload, true);
      }
    },
    [
      queryOptions.date_range,
      querySaved,
      runQuery,
      camp_dateRange,
      dispatch,
      queryType,
      runCampaignsQuery,
    ]
  );

  const handleDurationChange = useCallback(
    (dates, isCompareDate) => {
      let from,
        to,
        frequency = 'date',
        { dateType } = dates;

      if (Array.isArray(dates.startDate)) {
        from = dates.startDate[0];
        to = dates.startDate[1];
      } else {
        from = dates.startDate;
        to = dates.endDate;
      }

      if (queryType === QUERY_TYPE_EVENT || queryType === QUERY_TYPE_CAMPAIGN) {
        frequency = getValidGranularityOptions({ from, to }, queryType)[0];
      }

      const payload = {
        from: MomentTz(from).startOf('day'),
        to: MomentTz(to).endOf('day'),
        frequency,
        dateType,
      };

      if (!isCompareDate) {
        setQueryOptions((currState) => {
          return {
            ...currState,
            date_range: {
              ...currState.date_range,
              ...payload,
            },
          };
        });
      }

      if (isCompareDate) {
        localDispatch({
          type: SET_COMPARE_DURATION,
          payload: { from, to, frequency, dateType },
        });
      }

      const appliedDateRange = {
        ...queryOptions.date_range,
        ...payload,
      };

      if (queryType === QUERY_TYPE_FUNNEL) {
        runFunnelQuery(querySaved, appliedDateRange, isCompareDate);
      }
      if (queryType === QUERY_TYPE_EVENT) {
        runQuery(querySaved, appliedDateRange);
      }

      if (queryType === QUERY_TYPE_CAMPAIGN) {
        dispatch({ type: SET_CAMP_DATE_RANGE, payload });
        runCampaignsQuery(querySaved, payload);
      }

      if (queryType === QUERY_TYPE_ATTRIBUTION) {
        if (!isCompareDate) {
          // set range in reducer only when original date is changed and not the comparisom date
          dispatch({ type: SET_ATTR_DATE_RANGE, payload });
        }
        runAttributionQuery(querySaved, payload, isCompareDate);
      }
    },
    [
      queryType,
      runFunnelQuery,
      runQuery,
      querySaved,
      queryOptions.date_range,
      dispatch,
      runCampaignsQuery,
      runAttributionQuery,
    ]
  );

  useEffect(() => {
    if (rowClicked) {
      if (rowClicked.queryType === QUERY_TYPE_FUNNEL) {
        runFunnelQuery(rowClicked.queryName);
      } else if (rowClicked.queryType === QUERY_TYPE_ATTRIBUTION) {
        runAttributionQuery(rowClicked.queryName);
      } else if (rowClicked.queryType === QUERY_TYPE_CAMPAIGN) {
        runCampaignsQuery(rowClicked.queryName);
      } else {
        runQuery(rowClicked.queryName);
      }
      setRowClicked(false);
    }
  }, [
    rowClicked,
    runFunnelQuery,
    runQuery,
    runAttributionQuery,
    runCampaignsQuery,
  ]);

  useEffect(() => {
    return () => {
      dispatch({ type: SHOW_ANALYTICS_RESULT, payload: false });
    };
  }, [dispatch]);

  const queryChange = (newEvent, index, changeType = 'add') => {
    const queryupdated = [...queries];
    if (queryupdated[index]) {
      if (changeType === 'add') {
        if (JSON.stringify(queryupdated[index]) !== JSON.stringify(newEvent)) {
          deleteGroupByForEvent(newEvent, index);
        }
        queryupdated[index] = newEvent;
      } else {
        if (changeType === 'filters_updated') {
          // dont remove group by if filter is changed
          queryupdated[index] = newEvent;
        } else {
          deleteGroupByForEvent(newEvent, index);
          queryupdated.splice(index, 1);
        }
      }
    } else {
      queryupdated.push(newEvent);
    }
    setQueries(queryupdated);
  };

  const closeDrawer = () => {
    setDrawerVisible(false);
  };

  const setExtraOptions = (options) => {
    setQueryOptions(options);
  };

  const IconAndTextSwitchQueryType = (queryType) => {
    switch (queryType) {
      case QUERY_TYPE_EVENT:
        return {
          text: 'Analyse Events',
          icon: 'events_cq',
        };
      case QUERY_TYPE_FUNNEL:
        return {
          text: 'Find event funnel for',
          icon: 'funnels_cq',
        };
      case QUERY_TYPE_CAMPAIGN:
        return {
          text: 'Campaign Analytics',
          icon: 'campaigns_cq',
        };
      case QUERY_TYPE_ATTRIBUTION:
        return {
          text: 'Attributions',
          icon: 'attributions_cq',
        };
      default:
        return {
          text: 'Templates',
          icon: 'templates_cq',
        };
    }
  };

  const title = () => {
    const IconAndText = IconAndTextSwitchQueryType(queryType);
    return (
      <div className={'flex justify-between items-center'}>
        <div className={'flex items-center'}>
          <SVG name={IconAndText.icon} size='24px'></SVG>
          <Text
            type={'title'}
            level={4}
            weight={'bold'}
            extraClass={'ml-2 m-0'}
          >
            {IconAndText.text}
          </Text>
        </div>
        <div className={'flex justify-end items-center'}>
          <Button size={'large'} type='text' onClick={() => closeDrawer()}>
            <SVG name='times'></SVG>
          </Button>
        </div>
      </div>
    );
  };

  const campaignsArrayMapper = useMemo(() => {
    return campaignState.select_metrics.map((metric, index) => {
      return {
        eventName: metric,
        index,
        mapper: `event${index + 1}`,
      };
    });
  }, [campaignState.select_metrics]);

  const arrayMapper = useMemo(() => {
    return appliedQueries.map((q, index) => {
      return {
        eventName: q,
        index,
        mapper: `event${index + 1}`,
        displayName: eventNames[q],
      };
    });
  }, [appliedQueries, eventNames]);

  const checkIfnewComposer = () => {
    return (
      queryType === QUERY_TYPE_FUNNEL ||
      queryType === QUERY_TYPE_EVENT ||
      queryType === QUERY_TYPE_ATTRIBUTION
    );
  };

  const renderQueryComposer = () => {
    if (queryType === QUERY_TYPE_FUNNEL || queryType === QUERY_TYPE_EVENT) {
      return (
        <QueryComposer
          queries={queries}
          runQuery={runQuery}
          eventChange={queryChange}
          queryType={queryType}
          queryOptions={queryOptions}
          setQueryOptions={setExtraOptions}
          runFunnelQuery={runFunnelQuery}
          activeKey={activeKey}
        />
      );
    }

    if (queryType === QUERY_TYPE_ATTRIBUTION) {
      return <AttrQueryComposer runAttributionQuery={runAttributionQuery} />;
    }

    if (queryType === QUERY_TYPE_CAMPAIGN) {
      return (
        <CampQueryComposer
          handleRunQuery={runCampaignsQuery}
        ></CampQueryComposer>
      );
    }
  };

  // const renderQueryHeader = () => {
  //   if (true) return null;

  //   let eventList = [];
  //   if (requestQuery.cl === 'attribution') {
  //     eventList = [requestQuery.query.ce];
  //   } else {
  //     eventList = isArray(requestQuery)
  //       ? requestQuery[0].ewp
  //       : requestQuery.ewp;
  //   }
  //   return (
  //     <div className={`${styles.query_header} flex flex-col fa-act-header`}>
  //       <span className={`${styles.query_header__link}`}>
  //         {queryType} Analysis
  //       </span>
  //       <div className={`${styles.query_header__content}`}>
  //         {requestQuery &&
  //           eventList?.map((ev, index) => {
  //             return (
  //               <>
  //                 <Text
  //                   level={6}
  //                   type={'title'}
  //                   extraClass={'m-0'}
  //                   weight={'bold'}
  //                 >
  //                   {ev?.na}
  //                 </Text>
  //                 <span className={`${styles.query_header__content__trail}`}>
  //                   ...
  //                 </span>
  //               </>
  //             );
  //           })}
  //       </div>
  //     </div>
  //   );
  // };

  const renderQueryComposerNew = () => {
    if (
      queryType === QUERY_TYPE_FUNNEL ||
      queryType === QUERY_TYPE_EVENT ||
      queryType === QUERY_TYPE_ATTRIBUTION
    ) {
      return (
        <div
          className={queryOpen ? `query_card_open-add` : `query_card_close`}
          onClick={() => !queryOpen && setQueryOpen(true)}
        >
          {renderQueryComposer()}
        </div>
      );

      // return (
      //   <Collapse activeKey={true? ["1"] : ["0"]} onChange={() => setQueryOpen(!queryOpen)} className={`fa-query-edit ${queryOpen? 'query-open': ''}`} expandIcon={() =>
      //     <SVG  name={`sliders`} extraClass={`query_header_icon`}></SVG>}>
      //     <Panel id={true? 'query_header' : ''} header={renderQueryHeader()} key="1">
      //       <div>
      //         {renderQueryComposer()}
      //       </div>
      //     </Panel>
      //   </Collapse>
      // )
    }
    return null;
  };

  const handleBreadCrumbClick = () => {
    setShowResult(false);
    setNavigatedFromDashboard(false);
    setQuerySaved(false);
    updateRequestQuery(null);
    closeDrawer();
  };

  function changeTab(key) {
    // console.log('current tab is=-->>',key);
    setActiveTab(key);
  }

  const renderCreateQFlow = () => {
    return (
      <CoreQueryContext.Provider
        value={{
          coreQueryState,
          attributionMetrics,
          setAttributionMetrics,
          setNavigatedFromDashboard,
          resetComparisonData,
          handleCompareWithClick,
        }}
      >
        <Modal
          title={
            <AnalysisHeader
              requestQuery={requestQuery}
              onBreadCrumbClick={handleBreadCrumbClick}
              queryType={queryType}
              queryTitle={querySaved}
              setQuerySaved={setQuerySaved}
              breakdownType={breakdownType}
              changeTab={changeTab}
              activeTab={activeTab}
              getCurrentSorter={() => {}}
            />
          }
          visible={drawerVisible}
          footer={null}
          centered={false}
          // zIndex={1005}
          mask={false}
          closable={false}
          className={'fa-modal--full-width'}
        >

          <div className='mt-8 px-20'>
            <ErrorBoundary
              fallback={
                <FaErrorComp
                  size={'medium'}
                  title={'Analyse Results Error'}
                  subtitle={
                    'We are facing trouble loading Analyse results. Drop us a message on the in-app chat.'
                  }
                />
              }
              onError={FaErrorLog}
            >
              {Number(activeTab) === 1 && <>{renderQueryComposerNew()}</>}
            </ErrorBoundary>
          </div>
        </Modal>
      </CoreQueryContext.Provider>
    );
  };

  const composerFunctions = {
    runQuery,
    queryChange,
    setExtraOptions,
    runFunnelQuery,
    runAttributionQuery,
    activeKey,
    queries,
    showResult,
  };

  const closeResultPage = (flag = false) => {
    setQuerySaved(false);
    setDrawerVisible(flag);
  };

  return (
    <>
      <ErrorBoundary
        fallback={
          <FaErrorComp
            size={'medium'}
            title={'Analyse Error'}
            subtitle={
              'We are facing trouble loading Analyse. Drop us a message on the in-app chat.'
            }
          />
        }
        onError={FaErrorLog}
      >
        {
          <Drawer
            title={title()}
            placement='left'
            closable={false}
            visible={drawerVisible && !checkIfnewComposer()}
            onClose={closeDrawer}
            getContainer={false}
            width={'650px'}
            className={'fa-drawer'}
          >
            <ErrorBoundary
              fallback={
                <FaErrorComp subtitle={'Facing issues with Query Builder'} />
              }
              onError={FaErrorLog}
            >
              {renderQueryComposer()}
            </ErrorBoundary>
          </Drawer>
        }

        {!showResult && drawerVisible && checkIfnewComposer()
          ? renderCreateQFlow()
          : !showResult && (
              <CoreQueryHome
                setQueryType={setQueryType}
                setDrawerVisible={closeResultPage}
                setQueries={setQueries}
                setQueryOptions={setExtraOptions}
                setRowClicked={setRowClicked}
                location={location}
                setActiveKey={setActiveKey}
                setBreakdownType={setBreakdownType}
                setNavigatedFromDashboard={setNavigatedFromDashboard}
                updateSavedQuerySettings={updateSavedQuerySettings}
              />
            )}

        {showResult ? (
          <CoreQueryContext.Provider
            value={{
              coreQueryState,
              attributionMetrics,
              setAttributionMetrics,
              setNavigatedFromDashboard,
              resetComparisonData,
              handleCompareWithClick,
            }}
          >
            <AnalysisResultsPage
              queryType={queryType}
              resultState={resultState}
              setDrawerVisible={closeResultPage}
              requestQuery={requestQuery}
              queries={appliedQueries}
              breakdown={appliedBreakdown}
              setShowResult={() => {
                setShowResult(false);
                updateRequestQuery(false);
              }}
              querySaved={querySaved}
              setQuerySaved={setQuerySaved}
              durationObj={queryOptions.date_range}
              handleDurationChange={handleDurationChange}
              arrayMapper={arrayMapper}
              queryOptions={queryOptions}
              attributionsState={attributionsState}
              breakdownType={breakdownType}
              campaignState={campaignState}
              eventPage={result_criteria}
              section={REPORT_SECTION}
              runAttrCmprQuery={null}
              cmprResultState={null}
              campaignsArrayMapper={campaignsArrayMapper}
              handleGranularityChange={handleGranularityChange}
              updateChartTypes={updateChartTypes}
              composerFunctions={composerFunctions}
            />
          </CoreQueryContext.Provider>
        ) : null}
      </ErrorBoundary>
    </>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      deleteGroupByForEvent,
      getCampaignConfigData,
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(CoreQuery);
